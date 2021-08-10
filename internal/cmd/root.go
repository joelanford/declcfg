package cmd

import (
	"github.com/joelanford/declcfg/internal/cmd/inheritchannels"
	"github.com/joelanford/declcfg/internal/cmd/semver"
	"github.com/joelanford/declcfg/internal/cmd/version"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{Use: "declcfg"}
	cmd.AddCommand(
		inheritchannels.NewCmd(),
		semver.NewCmd(),
		version.NewCmd(),
	)
	return cmd
}
