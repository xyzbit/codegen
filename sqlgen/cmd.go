package sqlgen

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xyzbit/codegen/sqlgen/pkg/types"
)

const buildVersion = "0.0.1"

var (
	arg        = types.DefaultRunArg()
	configFile string
)

var Cmd = &cobra.Command{
	Use:   "dbrepo",
	Short: "A cli for mysql generator",
}

var gormCmd = &cobra.Command{
	Use:   "gorm",
	Short: "Generate gorm model",
	PreRun: func(cmd *cobra.Command, args []string) {
		// 如果指定了配置文件，从配置文件加载
		if configFile != "" {
			config, err := types.LoadConfig(configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "加载配置文件失败: %v\n", err)
				os.Exit(1)
			}
			// 保留命令行模式
			mode := arg.Mode
			arg = *config
			arg.Mode = mode
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		arg.Mode = types.GORM
		Run(arg)
	},
}

func init() {
	// flags init
	persistentFlags := Cmd.PersistentFlags()
	persistentFlags.StringVarP(&configFile, "config", "c", "", "配置文件路径")
	persistentFlags.StringVarP(&arg.DSN, "dsn", "d", "", "Mysql address")
	persistentFlags.StringSliceVarP(&arg.Table, "table", "t", []string{"*"}, "Patterns of table name")
	persistentFlags.StringSliceVarP(&arg.Filename, "filename", "f", []string{"*.sql"}, "Patterns of SQL filename")
	persistentFlags.StringVarP(&arg.Output, "output", "o", ".", "The adapter output directory")
	persistentFlags.StringVarP(&arg.EntityOutput, "entity-output", "e", ".", "The entity output directory")
	persistentFlags.StringVarP(&arg.RepoOutput, "repo-output", "i", ".", "The port output directory")
	persistentFlags.StringVarP(&arg.RepoPackage, "repo-package", "p", "", "The port packge full name")
	persistentFlags.StringVarP(&arg.EntityPackage, "entity-package", "E", "", "The entity packge full name")
	persistentFlags.BoolVarP(&arg.AutoAudit, "auto-audit", "a", false, "Whether to turn on automatic audit mode")
	persistentFlags.StringSliceVar(&arg.MockTypes, "mock-type", nil, "Types of mock files to generate (sqlite, docker)")

	// sub commands init
	Cmd.AddCommand(gormCmd)
	Cmd.Version = buildVersion
	Cmd.CompletionOptions.DisableDefaultCmd = true
}
