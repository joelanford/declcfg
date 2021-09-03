package cmd

import (
	"fmt"
	"github.com/blang/semver/v4"
	"io"
	"os"

	"github.com/joelanford/declcfg/internal/action"
	oraction "github.com/operator-framework/operator-registry/alpha/action"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newSemverCmd(log *logrus.Logger) *cobra.Command {
	output := ""
	templates := []string{}
	skipPatch := false
	semverRangeStr := ""
	cmd := &cobra.Command{
		Use:  "semver <indexRef> <packageName>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			packageName := args[1]

			var (
				semverRange semver.Range
				err error
			)
			if semverRangeStr == "" {
				semverRange = func(semver.Version) bool { return true }
			} else {
				semverRange, err = semver.ParseRange(semverRangeStr)
				if err != nil {
					return fmt.Errorf("invalid semver range %q", semverRangeStr)
				}
			}

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

			s := action.Semver{
				Configs:      *cfg,
				PackageName:  packageName,
				SkipPatch:    skipPatch,
				TemplateStrings: templates,
				SemverRange: semverRange,
			}
			out, err := s.Run()
			if err != nil {
				log.Fatalf("semver %q: %v", packageName, err)
			}

			if err := write(*out, os.Stdout); err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "json", "Output format (json|yaml)")
	cmd.Flags().BoolVar(&skipPatch, "skip-patch", false, "Add skips for intermediate semver patch versions")
	cmd.Flags().StringSliceVarP(&templates, "templates", "t", []string{"default"}, "Template strings evaluated against semver versions to generate channel names")
	cmd.Flags().StringVarP(&semverRangeStr, "semver-range", "r", "", "Semver range of bundles to consider when building channels" )
	return cmd
}
