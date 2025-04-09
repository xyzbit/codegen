package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xyzbit/codegen/pkg/install"
	"github.com/xyzbit/codegen/sqlgen"
)

var rootCmd = &cobra.Command{
	Use:   "codegen",
	Short: "代码生成工具",
}

func init() {
	rootCmd.AddCommand(sqlgen.Cmd)

	// 添加 install 子命令
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "安装 codegen 规则到 Cursor IDE",
		RunE: func(cmd *cobra.Command, args []string) error {
			return install.NewCommand().Run()
		},
	}
	rootCmd.AddCommand(installCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
