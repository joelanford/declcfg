package cmd

import (
	"fmt"

	"github.com/joelanford/declcfg/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("%#v\n", version.Version)
		},
	}
	return cmd
}
