package parser

import (
	"fmt"

	"github.com/xyzbit/codegen/pkg/buffer"
	"github.com/xyzbit/codegen/pkg/set"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/test_driver"

	"github.com/xyzbit/codegen/sqlgen/pkg/spec"
)

var p *parser.Parser

type stmts []stmt

type stmt interface {
	nodes() []ast.StmtNode
}

type createTableStmt struct {
	stmt *ast.CreateTableStmt
}

func (c createTableStmt) nodes() []ast.StmtNode {
	return []ast.StmtNode{c.stmt}
}

type queryStmt struct {
	stmt ast.StmtNode
}

func (q queryStmt) nodes() []ast.StmtNode {
	return []ast.StmtNode{q.stmt}
}

type transactionStmt struct {
	startTransactionStmt ast.StmtNode
	queryList            stmts
	commitStmt           ast.StmtNode
}

func (t transactionStmt) nodes() []ast.StmtNode {
	var list []ast.StmtNode
	for _, v := range t.queryList {
		stmt, ok := v.(*queryStmt)
		if ok {
			list = append(list, stmt.stmt)
		}
	}
	return list
}

func init() {
	p = parser.New()
}

// Parse parses a SQL statement string and returns a spec.DXL.
func Parse(sql string) (*spec.DXL, error) {
	// fmt.Printf("sql %s \n", sql)
	stmtNodes, _, err := p.Parse(sql, "", "")
	if err != nil {
		return nil, err
	}

	stmt, err := splits(stmtNodes)
	if err != nil {
		return nil, err
	}

	var ret spec.DXL
	for _, stmtNode := range stmt {
		switch node := stmtNode.(type) {
		case *createTableStmt:
			ddl, err := parseDDL(node.stmt)
			if err != nil {
				return nil, err
			}
			ret.DDL = append(ret.DDL, ddl)
		case *queryStmt:
			dml, err := parseDML(node.stmt)
			if err != nil {
				return nil, err
			}
			ret.DML = append(ret.DML, dml)
		case *transactionStmt:
			if node.queryList.hasTransactionStmt() {
				return nil, errorUnsupportedNestedTransaction
			}
			if len(node.queryList) == 0 {
				continue
			}
			dml, err := parseTransaction(node)
			if err != nil {
				return nil, err
			}
			ret.DML = append(ret.DML, dml)
		default:
			// ignores other statements
		}
	}

	if err = ret.Validate(); err != nil {
		return nil, err
	}

	return &ret, nil
}

func splits(stmtNodes []ast.StmtNode) ([]stmt, error) {
	var list stmts
	var transactionMode bool
	for _, v := range stmtNodes {
		switch node := v.(type) {
		case *ast.CreateTableStmt:
			if transactionMode {
				return nil, fmt.Errorf("missing begin stmt near by '%s'", v.Text())
			}
			list = append(list, &createTableStmt{stmt: node})
		case *ast.InsertStmt, *ast.SelectStmt, *ast.DeleteStmt, *ast.UpdateStmt:
			if transactionMode {
				transactionNode := list[len(list)-1].(*transactionStmt)
				transactionNode.queryList = append(transactionNode.queryList, &queryStmt{stmt: node})
			} else {
				list = append(list, &queryStmt{stmt: node})
			}
		case *ast.BeginStmt:
			if transactionMode {
				transactionNode := list[len(list)-1].(*transactionStmt)
				transactionNode.queryList = append(transactionNode.queryList, &transactionStmt{startTransactionStmt: node})
			} else {
				transactionMode = true
				list = append(list, &transactionStmt{startTransactionStmt: v})
			}
		case *ast.CommitStmt:
			if transactionMode {
				transactionNode := list[len(list)-1].(*transactionStmt)
				transactionNode.commitStmt = v
				transactionMode = false
			} else {
				return nil, fmt.Errorf("missing begin stmt near by '%s'", v.Text())
			}
		default:
			return nil, errorUnsupportedStmt
		}
	}
	if transactionMode {
		return nil, errorMissingCommit
	}
	return list, nil
}

func (s stmts) hasTransactionStmt() bool {
	for _, v := range s {
		if _, ok := v.(*transactionStmt); ok {
			return true
		}
	}
	return false
}

func parseDDL(node *ast.CreateTableStmt) (*spec.DDL, error) {
	var ddl spec.DDL
	ddl.Table = parseCreateTableStmt(node)
	return &ddl, nil
}

func parseDelete(stmt *ast.DeleteStmt) (spec.DML, error) {
	var ret spec.DeleteStmt
	text := stmt.Text()
	comment, err := parseLineComment(text)
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	sql, err := NewSqlScanner(text).ScanAndTrim()
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	if stmt.IsMultiTable {
		return nil, errorNearBy(errorMultipleTable, text)
	}

	tableName, err := parseTableRefsClause(stmt.TableRefs)
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	if stmt.Where != nil {
		where, err := parseExprNode(stmt.Where, tableName, exprTypeWhereClause)
		if err != nil {
			return nil, errorNearBy(err, text)
		}

		ret.Where = where
	}

	if stmt.Order != nil {
		orderBy, err := parseOrderBy(stmt.Order, tableName)
		if err != nil {
			return nil, errorNearBy(err, text)
		}

		ret.OrderBy = orderBy
	}

	if stmt.Limit != nil {
		limit, err := parseLimit(stmt.Limit)
		if err != nil {
			return nil, errorNearBy(err, text)
		}

		ret.Limit = limit
	}

	ret.Comment = comment
	ret.SQL = sql
	ret.Action = spec.ActionDelete
	ret.From = tableName
	return &ret, nil
}

func parseDML(node ast.StmtNode) (spec.DML, error) {
	switch v := node.(type) {
	case *ast.InsertStmt:
		return parseInsert(v)
	case *ast.SelectStmt:
		return parseSelect(v)
	case *ast.DeleteStmt:
		return parseDelete(v)
	case *ast.UpdateStmt:
		return parseUpdate(v)
	default:
		return nil, errorUnsupportedStmt
	}
}

func parseTableRefsClause(clause *ast.TableRefsClause) (string, error) {
	if clause == nil {
		return "", errorMissingTable
	}

	join := clause.TableRefs
	if join == nil {
		return "", errorMissingTable
	}

	if join.Left == nil {
		return "", errorMissingTable
	}

	if join.Right != nil {
		return "", errorMultipleTable
	}

	tableName, err := parseResultSetNode(join.Left)
	if err != nil {
		return "", err
	}

	return tableName, nil
}

func parseTransaction(node *transactionStmt) (spec.DML, error) {
	if node == nil {
		return nil, errorMissingTransaction
	}
	sqlBuilder := buffer.New()
	beginText := node.startTransactionStmt.Text()
	commitText := node.commitStmt.Text()
	beginSQL, err := NewSqlScanner(beginText).ScanAndTrim()
	if err != nil {
		return nil, errorNearBy(err, beginText)
	}
	commitSQL, err := NewSqlScanner(commitText).ScanAndTrim()
	if err != nil {
		return nil, errorNearBy(err, commitText)
	}

	comment, err := parseLineComment(beginText)
	if err != nil {
		return nil, err
	}

	sqlBuilder.Write(beginSQL)
	var ret spec.Transaction
	ret.Action = spec.ActionTransaction
	for _, v := range node.nodes() {
		dml, err := parseDML(v)
		if err != nil {
			return nil, err
		}
		sqlBuilder.Write(dml.SQLText())
		ret.Statements = append(ret.Statements, dml)
	}
	sqlBuilder.Write(commitSQL)
	ret.SQL = sqlBuilder.String()
	ret.Comment = comment
	return &ret, nil
}

func parseColumnDef(col *ast.ColumnDef) (*spec.Column, *spec.Constraint) {
	if col == nil || col.Name == nil {
		return nil, nil
	}

	var column spec.Column
	constraint := spec.NewConstraint()
	tp := col.Tp
	if tp != nil {
		column.Unsigned = mysql.HasUnsignedFlag(tp.Flag)
		column.TP = tp.Tp
	}

	column.Name = col.Name.String()
	for _, opt := range col.Options {
		tp := opt.Tp
		switch tp {
		case ast.ColumnOptionNotNull:
			column.NotNull = true
		case ast.ColumnOptionAutoIncrement:
			column.AutoIncrement = true
		case ast.ColumnOptionDefaultValue:
			column.HasDefaultValue = true
		case ast.ColumnOptionComment:
			expr := opt.Expr
			if expr != nil {
				value, ok := expr.(*test_driver.ValueExpr)
				if ok {
					column.Comment = value.GetString()
				}
			}
		case ast.ColumnOptionUniqKey:
			constraint.AppendUniqueKey(column.Name, column.Name)
		case ast.ColumnOptionPrimaryKey:
			constraint.AppendPrimaryKey(column.Name, column.Name)
		default:
			// ignore other options
		}
	}

	return &column, constraint
}

func parseConstraint(constraint *ast.Constraint) *spec.Constraint {
	if constraint == nil {
		return nil
	}

	columns := parseColumnFromKeys(constraint.Keys)
	if len(columns) == 0 {
		return nil
	}

	ret := spec.NewConstraint()
	key := constraint.Name
	switch constraint.Tp {
	case ast.ConstraintPrimaryKey:
		ret.AppendPrimaryKey(key, columns...)
	case ast.ConstraintKey, ast.ConstraintIndex:
		ret.AppendIndex(key, columns...)
	case ast.ConstraintUniq, ast.ConstraintUniqKey, ast.ConstraintUniqIndex:
		ret.AppendUniqueKey(key, columns...)
	default:
		// ignore other constraints
	}

	return ret
}

func parseColumnFromKeys(keys []*ast.IndexPartSpecification) []string {
	columnSet := set.From()
	for _, key := range keys {
		if key.Column == nil {
			continue
		}

		columnName := key.Column.String()
		columnSet.Add(columnName)
	}

	return columnSet.String()
}

func parseInsert(stmt *ast.InsertStmt) (*spec.InsertStmt, error) {
	text := stmt.Text()
	comment, err := parseLineComment(text)
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	sql, err := NewSqlScanner(text).ScanAndTrim()
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	var ret spec.InsertStmt
	ret.Comment = comment
	tableName, err := parseTableRefsClause(stmt.Table)
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	columns, err := parseColumns(stmt.Columns, tableName)
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	ret.Table = tableName
	ret.Action = spec.ActionCreate
	ret.SQL = sql
	ret.Columns = columns

	return &ret, nil
}

func parseCreateTableStmt(stmt *ast.CreateTableStmt) *spec.Table {
	var table spec.Table
	if stmt.Table != nil {
		table.Name = stmt.Table.Name.String()
	}

	constraint := spec.NewConstraint()
	for _, col := range stmt.Cols {
		column, con := parseColumnDef(col)
		if column != nil {
			table.Columns = append(table.Columns, *column)
		}
		constraint.Merge(con)
	}

	for _, c := range stmt.Constraints {
		constraint.Merge(parseConstraint(c))
	}

	table.Constraint = *constraint
	return &table
}

func parseUpdate(stmt *ast.UpdateStmt) (spec.DML, error) {
	var ret spec.UpdateStmt
	text := stmt.Text()
	comment, err := parseLineComment(text)
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	sql, err := NewSqlScanner(text).ScanAndTrim()
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	if stmt.MultipleTable {
		return nil, errorNearBy(errorMultipleTable, text)
	}

	tableName, err := parseTableRefsClause(stmt.TableRefs)
	if err != nil {
		return nil, errorNearBy(err, text)
	}

	if stmt.Where != nil {
		where, err := parseExprNode(stmt.Where, tableName, exprTypeWhereClause)
		if err != nil {
			return nil, errorNearBy(err, text)
		}

		ret.Where = where
	}

	if stmt.Order != nil {
		orderBy, err := parseOrderBy(stmt.Order, tableName)
		if err != nil {
			return nil, errorNearBy(err, text)
		}

		ret.OrderBy = orderBy
	}

	if stmt.Limit != nil {
		limit, err := parseLimit(stmt.Limit)
		if err != nil {
			return nil, errorNearBy(err, text)
		}

		ret.Limit = limit
	}

	for _, a := range stmt.List {
		colName, err := parseColumn(a.Column, tableName)
		if err != nil {
			return nil, errorNearBy(err, text)
		}

		if len(colName) > 0 {
			ret.Columns = append(ret.Columns, colName)
		}
	}

	ret.Comment = comment
	ret.SQL = sql
	ret.Action = spec.ActionUpdate
	ret.Table = tableName

	return &ret, nil
}
