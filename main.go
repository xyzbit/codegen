package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xyzbit/codegen/sqlgen"
)

var rootCmd = &cobra.Command{
	Use:   "codegen",
	Short: "codegen",
}

func init() {
	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(sqlgen.Cmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
