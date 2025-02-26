package sqlgen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/xyzbit/codegen/pkg/patterns"
	"github.com/xyzbit/codegen/sqlgen/gen/gorm"
	"github.com/xyzbit/codegen/sqlgen/pkg/parser"
	"github.com/xyzbit/codegen/sqlgen/pkg/spec"
	"github.com/xyzbit/codegen/sqlgen/pkg/types"
)

const sqlExt = ".sql"

func Run(arg types.RunArg) {
	var err error
	if len(arg.DSN) > 0 {
		err = runFromDSN(arg)
	} else if len(arg.Filename) > 0 {
		err = runFromSQL(arg)
	} else {
		err = fmt.Errorf("missing dsn or filename")
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func runFromSQL(arg types.RunArg) error {
	var list []string
	for _, item := range arg.Filename {
		filename, err := filepath.Abs(item)
		if err != nil {
			return err
		}

		dir := filepath.Dir(filename)
		base := filepath.Base(filename)
		fileInfo, err := ioutil.ReadDir(dir)
		if err != nil {
			return err
		}
		var filenames []string
		for _, item := range fileInfo {
			if item.IsDir() {
				continue
			}

			ext := filepath.Ext(item.Name())
			if ext != sqlExt {
				continue
			}

			f := filepath.Join(dir, item.Name())
			filenames = append(filenames, f)
		}
		p := patterns.New(base)
		matchSQLFile := p.Match(filenames...)

		list = append(list, matchSQLFile...)
	}

	if len(list) == 0 {
		return fmt.Errorf("no sql file found")
	}

	var ret spec.DXL
	for _, file := range list {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		dxl, err := parser.Parse(string(data))
		if err != nil {
			return err
		}

		ret.DDL = append(ret.DDL, dxl.DDL...)
		ret.DML = append(ret.DML, dxl.DML...)
	}

	return run(&ret, arg.Mode, arg)
}

func runFromDSN(arg types.RunArg) error {
	dxl, err := parser.From(arg.DSN, arg.Table...)
	if err != nil {
		return err
	}

	return run(dxl, arg.Mode, arg)
}

var funcMap = map[types.Mode]func(context []spec.Context, arg types.RunArg) error{
	// SQL:  sql.Run,
	types.GORM: gorm.Run,
	// XORM: xorm.Run,
	// SQLX: sqlx.Run,
	// BUN:  bun.Run,
}

func run(dxl *spec.DXL, mode types.Mode, arg types.RunArg) error {
	ctx, err := spec.From(dxl)
	if err != nil {
		return err
	}

	fn, ok := funcMap[mode]
	if !ok {
		return nil
	}

	return fn(ctx, arg)
}
