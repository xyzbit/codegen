package parser

import (
	_ "embed"
	"testing"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/stretchr/testify/assert"

	"github.com/xyzbit/codegen/sqlgen/pkg/spec"
)

//go:embed test.sql
var testSql string
var testParser *parser.Parser

func TestMain(m *testing.M) {
	testParser = parser.New()
	m.Run()
}

func TestParse(t *testing.T) {
	t.Run("ParseError", func(t *testing.T) {
		_, err := Parse("delete from where id = ?")
		assert.NotNil(t, err)
	})

	t.Run("ParseError", func(t *testing.T) {
		_, err := Parse("alter table foo add column name varchar(255);")
		assert.ErrorIs(t, err, errorUnsupportedStmt)
	})

	t.Run("success", func(t *testing.T) {
		dxl, err := Parse(testSql)
		assert.Nil(t, err)

		_, err = spec.From(dxl)
		assert.Nil(t, err)
	})
}

func Test_parseDML(t *testing.T) {
	t.Run("InsertStmt", func(t *testing.T) {
		stmt, _, err := testParser.Parse(`
-- fn:
insert into foo (name) values(?);
-- fn:
select * from foo;
-- fn:
delete from foo where id = ?;
-- fn:
update foo set name = ? where id = ?;
-- fn:
alter table foo add column bar varchar(255);
`, "", "")
		assert.NoError(t, err)
		for _, v := range stmt {
			_, err := parseDML(v)
			assert.NotNil(t, err)
		}
	})
}

func Test_parseTableRefsClause(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		_, err := parseTableRefsClause(nil)
		assert.ErrorIs(t, err, errorMissingTable)
	})

	t.Run("joinNil", func(t *testing.T) {
		_, err := parseTableRefsClause(&ast.TableRefsClause{})
		assert.ErrorIs(t, err, errorMissingTable)
	})

	t.Run("joinLeftNil", func(t *testing.T) {
		_, err := parseTableRefsClause(&ast.TableRefsClause{
			TableRefs: &ast.Join{
				Left:  &ast.TableName{},
				Right: &ast.TableName{},
			},
		})
		assert.ErrorIs(t, err, errorMultipleTable)
	})

	t.Run("parseResultSetNode", func(t *testing.T) {
		_, err := parseTableRefsClause(&ast.TableRefsClause{
			TableRefs: &ast.Join{
				Left: &ast.SelectStmt{},
			},
		})
		assert.ErrorIs(t, err, errorUnsupportedNestedQuery)
	})
}

func Test_parseTransaction(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		_, err := parseTransaction(nil)
		assert.ErrorIs(t, err, errorMissingTransaction)
	})
}
