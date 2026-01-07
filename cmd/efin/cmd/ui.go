package cmd

import efinui "github.com/artilugio0/efin-ui/pkg/cmd"

func init() {
	rootCmd.AddCommand(efinui.NewUICmd("ui"))
}
