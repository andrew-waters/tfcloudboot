package cmd

import (
	"fmt"
	"os"

	entities "github.com/andrew-waters/tfcloudboot/entities"
	"github.com/spf13/cobra"
)

var (
	inputFile   string
	outputDir   string
	outputName  string
	secretsFile string

	strapCmd = &cobra.Command{
		Use:   "strap",
		Short: "Bootstrap a new Terraform Cloud workspace",
		Long:  `Creates a new Terraform Cloud workspace stanza based on the input yaml`,
		Run:   strap,
	}
)

func init() {
	strapCmd.Flags().StringVarP(&inputFile, "file", "f", "", "The input yaml file")
	strapCmd.Flags().StringVarP(&secretsFile, "secrets", "s", "", "The location of a yaml file containing secrets to merge")
	strapCmd.Flags().StringVarP(&outputDir, "output", "o", "", "The output directory (leave blank for pwd)")
	strapCmd.Flags().StringVarP(&outputName, "name", "n", "", "The filename to save the generated workspace with")
	strapCmd.MarkFlagRequired("file")
}

func strap(ccmd *cobra.Command, args []string) {

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "Input not found - the provided input file does not exist")
		return
	}

	if secretsFile != "" {
		if _, err := os.Stat(secretsFile); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Secrets file not found - the provided secret file does not exist")
			return
		}
	}

	if outputDir == "" {
		outputDir = "./"
	}

	if outputDir != "./" {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			fmt.Fprintln(os.Stderr, "Could not make output path")
			return
		}
	}

	if outputName == "" {
		outputName = "workspace"
	}

	// create the workspace and output the rendered terraform files
	ws := entities.NewWorkspace(inputFile)
	ws.Output(outputDir, outputName, secretsFile)
}
