package cmd

import efinProxy "github.com/artilugio0/efin-proxy/pkg/cmd"

var proxyCmd = efinProxy.NewProxyCmd("proxy")

func init() {
	rootCmd.AddCommand(proxyCmd)
}
