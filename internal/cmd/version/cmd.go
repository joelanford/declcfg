package version

import (
	"fmt"

	"github.com/joelanford/declcfg/internal/version"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("%#v\n", version.Version)
		},
	}
	return cmd
}
