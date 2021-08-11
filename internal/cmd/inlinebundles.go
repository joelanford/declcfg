package cmd

import (
	"fmt"
	"github.com/joelanford/declcfg/internal/action"
	oraction "github.com/operator-framework/operator-registry/alpha/action"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
)



func newInlineBundlesCmd() *cobra.Command {
	pruneFromNonChannelHeads := false
	bundleImages := []string{}
	output := ""

	cmd := &cobra.Command{
		Use:  "inline-bundles <indexRef> <packageName>",
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

			i := action.InlineBundles{
				Configs:                  *cfg,
				PackageName:              packageName,
				BundleImages:             bundleImages,
				PruneFromNonChannelHeads: pruneFromNonChannelHeads,
				Logger:                   logrus.StandardLogger(),
			}
			out, err := i.Run(cmd.Context())
			if err != nil {
				logrus.Fatalf("inline bundles for package %q: %v", packageName, err)
			}

			if err := write(*out, os.Stdout); err != nil {
				logrus.Fatal(err)
			}

			return nil
		},
	}
	cmd.Flags().BoolVarP(&pruneFromNonChannelHeads, "prune", "p", false, "Prune objects for bundles that are not channel heads.")
	cmd.Flags().StringSliceVarP(&bundleImages, "bundles", "b", nil, "Specific bundle image references to inline objects for.")
	cmd.Flags().StringVarP(&output, "output", "o", "json", "Output format (json|yaml)")
	return cmd
}
