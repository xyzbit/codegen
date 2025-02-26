package gorm

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xyzbit/codegen/sqlgen/gen/testdata"
	"github.com/xyzbit/codegen/sqlgen/pkg/parser"
	"github.com/xyzbit/codegen/sqlgen/pkg/spec"
	"github.com/xyzbit/codegen/sqlgen/pkg/types"
)

func TestRun(t *testing.T) {
	dxl, err := parser.Parse(testdata.TestSql)
	assert.NoError(t, err)
	ctx, err := spec.From(dxl)
	assert.NoError(t, err)
	err = Run(ctx, types.RunArg{
		Output:        t.TempDir(),
		RepoOutput:    t.TempDir(),
		EntityOutput:  t.TempDir(),
		RepoPackage:   "",
		EntityPackage: "",
		AutoAudit:     false,
	})
	assert.NoError(t, err)
}
