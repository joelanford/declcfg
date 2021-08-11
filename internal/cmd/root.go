package cmd

import (
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{Use: "declcfg"}
	cmd.AddCommand(
		newInheritChannelsCmd(),
		newInlineBundlesCmd(),
		newSemverCmd(),
		newVersionCmd(),
	)
	return cmd
}
