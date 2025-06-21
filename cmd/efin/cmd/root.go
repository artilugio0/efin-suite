package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "efin",
	Short: "Suite of Web Security Tools",
	Long: `Efin Suite is a collection of web security tools
designed to be run on the command line.`,
}

func Execute() error {
	return rootCmd.Execute()
}
