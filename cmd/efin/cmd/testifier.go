package cmd

import efinTestifier "github.com/artilugio0/efin-testifier/pkg/cmd"

func init() {
	rootCmd.AddCommand(efinTestifier.NewTestifierCmd("testifier"))
}
