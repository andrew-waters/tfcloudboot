package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cloudBootCmd = &cobra.Command{
		Use:   "tfcloudboot",
		Short: "A tool for generating Terraform Cloud bootstrap files",
		Long:  ``,
		Run:   func(cmd *cobra.Command, args []string) {},
	}
)

func init() {
	cloudBootCmd.AddCommand(strapCmd)
}

// Execute Cobra
func Execute() {
	if err := cloudBootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
