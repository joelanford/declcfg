package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func New(log *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{Use: "declcfg"}
	cmd.AddCommand(
		newInheritChannelsCmd(log),
		newInlineBundlesCmd(log),
		newSemverCmd(log),
		newVersionCmd(),
	)
	return cmd
}
