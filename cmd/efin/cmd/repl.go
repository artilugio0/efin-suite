package cmd

import (
	"github.com/artilugio0/efin-suite/internal/repl"
	"github.com/spf13/cobra"
)

var (
	replDBFile string
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Run Efin REPL",
	Long: `Efin REPL is a Read-Evaluate-Print-Loop
program that lets you interact with all of the Efin Suite
tools interactively`,
	Run: func(cmd *cobra.Command, args []string) {
		repl.Run(replDBFile)
	},
}

func init() {
	replCmd.Flags().StringVarP(&replDBFile, "db-file", "D", "./proxy.db", "Requests DB file path")
	rootCmd.AddCommand(replCmd)
}
