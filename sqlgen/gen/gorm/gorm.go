package gorm

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/xyzbit/codegen/pkg/templatex"
	"github.com/xyzbit/codegen/sqlgen/pkg/spec"
	"github.com/xyzbit/codegen/sqlgen/pkg/types"
)

//go:embed gorm_adapter.go.tpl
var gormAdapterTpl string

//go:embed gorm_repo.go.tpl
var gormRepoTpl string

//go:embed gorm_entity.go.tpl
var gormEntityTpl string

//go:embed gorm_docker_mysql_mock.go.tpl
var gormDockerMySQLMockTpl string

//go:embed gorm_sqlite_mock.go.tpl
var gormSQLiteMockTpl string

// 模版数据
type TempData struct {
	spec.Context
	AdapterPackageName string
	RepoPackage        string
	RepoPackageName    string
	EntityPackage      string
	AutoAudit          bool
}

func Run(list []spec.Context, arg types.RunArg) error {
	for _, ctx := range list {
		td := TempData{
			Context:       ctx,
			RepoPackage:   arg.RepoPackage,
			EntityPackage: arg.EntityPackage,
			AutoAudit:     arg.AutoAudit,
		}
		adapterTmps := strings.Split(arg.Output, "/")
		td.AdapterPackageName = adapterTmps[len(adapterTmps)-1]
		repoTmps := strings.Split(arg.RepoOutput, "/")
		td.RepoPackageName = repoTmps[len(repoTmps)-1]

		adpterFilename := filepath.Join(arg.Output, fmt.Sprintf("%s_adpter.go", ctx.Table.Name))
		repoFilename := filepath.Join(arg.RepoOutput, fmt.Sprintf("%s_repo.go", ctx.Table.Name))
		entityFilename := filepath.Join(arg.EntityOutput, fmt.Sprintf("%s_entity.go", ctx.Table.Name))

		// 生成基础文件
		if err := generateFile(adpterFilename, gormAdapterTpl, td, funcMap, template.FuncMap{
			"IsPrimary": func(name string) bool {
				return ctx.Table.IsPrimary(name)
			},
			"IsExtraResult": func(name string) bool {
				return name != strcase.ToCamel(ctx.Table.Name)
			},
		}); err != nil {
			return err
		}

		if err := generateFile(repoFilename, gormRepoTpl, td, nil, nil); err != nil {
			return err
		}

		if err := generateFile(entityFilename, gormEntityTpl, td, funcMap, nil); err != nil {
			return err
		}

		// 根据 MockTypes 参数生成对应的 mock 文件
		for _, mockType := range arg.MockTypes {
			switch mockType {
			case types.MockDocker:
				dockerMockFilename := filepath.Join(arg.Output, fmt.Sprintf("%s_docker_mock_adapter.go", ctx.Table.Name))
				if err := generateFile(dockerMockFilename, gormDockerMySQLMockTpl, td, funcMap, template.FuncMap{
					"IsPrimary": func(name string) bool {
						return ctx.Table.IsPrimary(name)
					},
				}); err != nil {
					return err
				}
			case types.MockSQLite:
				sqliteMockFilename := filepath.Join(arg.Output, fmt.Sprintf("%s_sqlite_mock_adapter.go", ctx.Table.Name))
				if err := generateFile(sqliteMockFilename, gormSQLiteMockTpl, td, funcMap, template.FuncMap{
					"IsPrimary": func(name string) bool {
						return ctx.Table.IsPrimary(name)
					},
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// generateFile 生成文件的辅助函数
func generateFile(filename string, tpl string, data interface{}, baseFuncMap template.FuncMap, extraFuncMap template.FuncMap) error {
	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("[ignore] %s already exists\n", filename)
		return nil
	}

	gen := templatex.New()
	if baseFuncMap != nil {
		gen.AppendFuncMap(baseFuncMap)
	}
	if extraFuncMap != nil {
		gen.AppendFuncMap(extraFuncMap)
	}
	gen.MustParse(tpl)
	gen.MustExecute(data)
	gen.MustSaveAs(filename, true)
	return nil
}
