package semver

import (
	"fmt"
	"io"
	"os"

	"github.com/joelanford/declcfg/internal/action"
	oraction "github.com/operator-framework/operator-registry/alpha/action"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	output := ""
	channelNames := []string{}
	skipPatch := false
	cmd := &cobra.Command{
		Use:  "semver <indexRef> <packageName>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			packageName := args[1]

			var write func(declcfg.DeclarativeConfig, io.Writer) error
			switch output {
			case "json":
				write = declcfg.WriteJSON
			case "yaml":
				write = declcfg.WriteYAML
			default:
				return fmt.Errorf("invalid output format %q", output)
			}

			r := oraction.Render{
				Refs:           []string{ref},
				AllowedRefMask: oraction.RefDCDir | oraction.RefDCImage | oraction.RefSqliteFile | oraction.RefSqliteImage,
			}

			cfg, err := r.Run(cmd.Context())
			if err != nil {
				logrus.Fatalf("render index %q: %v", ref, err)
			}

			s := action.Semver{
				Configs:      *cfg,
				PackageName:  packageName,
				ChannelNames: channelNames,
				SkipPatch:    skipPatch,
			}
			out, err := s.Run()
			if err != nil {
				logrus.Fatalf("semver %q: %v", packageName, err)
			}

			if err := write(*out, os.Stdout); err != nil {
				logrus.Fatal(err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "json", "Output format (json|yaml)")
	cmd.Flags().BoolVar(&skipPatch, "skip-patch", false, "Add skips for intermediate semver patch versions")
	cmd.Flags().StringSliceVarP(&channelNames, "channels", "c", nil, "Channels to order as semver (default: all channels in package")
	return cmd
}
