package main

import (
	"context"
	"iptv/cmd/iptv/cmds"

	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(cmds.NewRootCLI().ExecuteContext(context.Background()))
}
