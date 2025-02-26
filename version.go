package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "v0.0.3"

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version)
	},
}
