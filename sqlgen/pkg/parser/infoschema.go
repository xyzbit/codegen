package parser

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"text/template"

	sql "github.com/go-sql-driver/mysql"
	"github.com/iancoleman/strcase"
	"github.com/pingcap/parser/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/xyzbit/codegen/pkg/patterns"
	"github.com/xyzbit/codegen/pkg/stringx"
	"github.com/xyzbit/codegen/sqlgen/pkg/infoschema"
	"github.com/xyzbit/codegen/sqlgen/pkg/spec"
)

var errMissingSchema = errors.New("missing schema")

func From(dsn string, pattern ...string) (*spec.DXL, error) {
	schema, url, err := parseDSN(dsn)
	if err != nil {
		return nil, err
	}

	conn := sqlx.NewMysql(url)
	model := infoschema.NewInformationSchemaModel(conn)
	tables, err := model.GetAllTables(schema)
	if err != nil {
		return nil, err
	}

	p := patterns.New(pattern...)
	matchTables := p.Match(tables...)
	var dxl spec.DXL
	for _, table := range matchTables {
		modelTable, err := model.FindColumns(schema, table)
		if err != nil {
			return nil, err
		}

		ddl, err := convertDDL(modelTable)
		if err != nil {
			return nil, err
		}

		dml, err := convertDML(ddl.Table)
		if err != nil {
			return nil, err
		}

		dxl.DDL = append(dxl.DDL, ddl)
		dxl.DML = append(dxl.DML, dml...)
	}

	return &dxl, nil
}

func parseDSN(dsn string) (db, url string, err error) {
	cfg, err := sql.ParseDSN(dsn)
	if err != nil {
		return "", "", err
	}

	if cfg.DBName == "" {
		return "", "", errMissingSchema
	}

	url = fmt.Sprintf("%s:%s@tcp(%s)/%s", cfg.User, cfg.Passwd, cfg.Addr, "information_schema")
	db = cfg.DBName
	return
}

//go:embed init.tpl.sql
var initSql string

func convertDML(in *spec.Table) ([]spec.DML, error) {
	t, err := template.New("sql").Parse(initSql)
	if err != nil {
		return nil, err
	}

	var sqlBuffer bytes.Buffer

	columnList := getSafeColumnList(in)
	if err = t.Execute(&sqlBuffer, map[string]interface{}{
		"insert_columns":  strings.Join(columnList, ", "),
		"insert_table":    in.Name,
		"insert_values":   stringx.RepeatJoin("?", ", ", len(columnList)),
		"unique_indexes":  getUniques(in),
		"general_indexes": getIndexes(in),
	}); err != nil {
		return nil, err
	}

	dxl, err := Parse(sqlBuffer.String())
	if err != nil {
		return nil, err
	}

	return dxl.DML, nil
}

func getSafeColumnList(in *spec.Table) []string {
	columnList := []string{}
	for _, c := range in.ColumnList() {
		columnList = append(columnList, fmt.Sprintf("`%s`", c))
	}
	return columnList
}

// Unique is a unique index info.
type Unique struct {
	SelectColumns  string
	Table          string
	UpdateSet      string
	WhereClause    string
	UniqueNameJoin string
}

func getUniques(in *spec.Table) []Unique {
	var list []Unique
	columns := strings.Join(getSafeColumnList(in), ", ")
	updateSet := strings.Join(getSafeColumnList(in), " = ?,") + " = ?"
	m := map[Unique]struct{}{}
	for _, c := range in.Constraint.PrimaryKey {
		item := Unique{
			SelectColumns:  columns,
			Table:          in.Name,
			UpdateSet:      updateSet,
			WhereClause:    strings.Join(c, " = ? AND ") + " = ?",
			UniqueNameJoin: strcase.ToCamel(strings.Join(c, "")),
		}
		if _, ok := m[item]; ok {
			continue
		}
		m[item] = struct{}{}
		list = append(list, item)
	}
	for _, c := range in.Constraint.UniqueKey {
		item := Unique{
			SelectColumns:  columns,
			Table:          in.Name,
			UpdateSet:      updateSet,
			WhereClause:    strings.Join(c, " = ? AND ") + " = ?",
			UniqueNameJoin: strcase.ToCamel(strings.Join(c, "")),
		}
		if _, ok := m[item]; ok {
			continue
		}
		m[item] = struct{}{}
		list = append(list, item)
	}

	return list
}

// Index is simple indx info
type Index struct {
	Items []Item
}
type Item struct {
	SelectColumns string
	Table         string
	UpdateSet     string
	WhereClause   string
	IndexNameJoin string
}

func getIndexes(in *spec.Table) []Index {
	var list []Index
	columns := strings.Join(getSafeColumnList(in), ", ")
	updateSet := strings.Join(getSafeColumnList(in), " = ?,") + " = ?"
	for _, cs := range in.Constraint.Index {
		var items []Item
		wcs := indexWhereColumns(cs)
		for _, c := range wcs {
			item := Item{
				SelectColumns: columns,
				Table:         in.Name,
				UpdateSet:     updateSet,
				WhereClause:   strings.Join(c, " = ? AND ") + " = ?",
				IndexNameJoin: strcase.ToCamel(strings.Join(c, "")),
			}
			items = append(items, item)
		}
		list = append(list, Index{Items: items})
	}
	return list
}

func indexWhereColumns(cs []string) [][]string {
	if len(cs) == 0 {
		return [][]string{}
	}
	whereColumns := make([][]string, 0, 1)
	for i := len(cs); i > 0; i-- {
		temp := []string{}
		for j := 0; j < i; j++ {
			temp = append(temp, cs[j])
		}
		whereColumns = append(whereColumns, temp)
	}
	return whereColumns
}

func convertDDL(in *infoschema.Table) (*spec.DDL, error) {
	var ddl spec.DDL
	constraint := spec.NewConstraint()
	getConstraint(in.Columns, constraint)
	var table spec.Table
	table.Name = in.Table
	table.Schema = in.Db
	if !constraint.IsEmpty() {
		table.Constraint = *constraint
	}

	for _, c := range in.Columns {
		extra := c.Extra
		autoIncrement := strings.Contains(extra, "auto_increment")
		unsigned := strings.Contains(c.DataType, "unsigned")
		tp, err := dbTypeMapper(c.DataType)
		if err != nil {
			return nil, err
		}

		table.Columns = append(table.Columns, spec.Column{
			ColumnOption: spec.ColumnOption{
				AutoIncrement:   autoIncrement,
				Comment:         stringx.TrimNewLine(c.Comment),
				HasDefaultValue: c.ColumnDefault != nil,
				NotNull:         !strings.EqualFold(c.IsNullAble, "yes"),
				Unsigned:        unsigned,
			},
			Name: c.Name,
			TP:   tp,
		})
	}

	ddl.Table = &table
	return &ddl, nil
}

func getConstraint(columns []*infoschema.Column, constraint *spec.Constraint) {
	for _, c := range columns {
		index := c.Index
		if index == nil {
			continue
		}
		indexName := index.IndexName
		if strings.EqualFold(indexName, "primary") {
			constraint.AppendPrimaryKey(indexName, c.Name)
		}
		if index.NonUnique == 0 {
			constraint.AppendUniqueKey(indexName, c.Name)
		} else {
			constraint.AppendIndex(indexName, c.Name)
		}
	}
}

var str2Type = map[string]byte{
	"bit":         mysql.TypeBit,
	"text":        mysql.TypeBlob,
	"date":        mysql.TypeDate,
	"datetime":    mysql.TypeDatetime,
	"unspecified": mysql.TypeUnspecified,
	"decimal":     mysql.TypeNewDecimal,
	"double":      mysql.TypeDouble,
	"enum":        mysql.TypeEnum,
	"float":       mysql.TypeFloat,
	"geometry":    mysql.TypeGeometry,
	"mediumint":   mysql.TypeInt24,
	"json":        mysql.TypeJSON,
	"int":         mysql.TypeLong,
	"bigint":      mysql.TypeLonglong,
	"longtext":    mysql.TypeLongBlob,
	"mediumtext":  mysql.TypeMediumBlob,
	"null":        mysql.TypeNull,
	"set":         mysql.TypeSet,
	"smallint":    mysql.TypeShort,
	"char":        mysql.TypeString,
	"time":        mysql.TypeDuration,
	"timestamp":   mysql.TypeTimestamp,
	"tinyint":     mysql.TypeTiny,
	"tinytext":    mysql.TypeTinyBlob,
	"varchar":     mysql.TypeVarchar,
	"var_string":  mysql.TypeVarString,
	"year":        mysql.TypeYear,
}

func dbTypeMapper(tp string) (byte, error) {
	l := strings.ToLower(tp)
	ret, ok := str2Type[l]
	if !ok {
		return 0, fmt.Errorf("unsupported type:%s", tp)
	}
	return ret, nil
}
