package cmd

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

func newInheritChannelsCmd(log *logrus.Logger) *cobra.Command {
	output := ""
	cmd := &cobra.Command{
		Use:  "inherit-channels <indexRef> <packageName>",
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
				log.Fatalf("render index %q: %v", ref, err)
			}

			i := action.InheritChannels{
				Configs:     *cfg,
				PackageName: packageName,
			}
			out, err := i.Run()
			if err != nil {
				log.Fatalf("inherit channels for package %q: %v", packageName, err)
			}

			if err := write(*out, os.Stdout); err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "json", "Output format (json|yaml)")
	return cmd
}
