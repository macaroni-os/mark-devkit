/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmdkit

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type DistfilesReport struct {
	Stats         *kit.AtomsStats  `json:"stats,omitempty" yaml:"stats,omitempty"`
	AtomsInErrors []*kit.AtomError `json:"atoms_errors,omitempty" yaml;"atoms_errors,omitempty"`
}

func (r *DistfilesReport) Yaml() ([]byte, error) {
	return yaml.Marshal(r)
}

func (r *DistfilesReport) Json() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (r *DistfilesReport) WriteJsonFile(f string) error {
	data, err := r.Json()
	if err != nil {
		return err
	}

	return os.WriteFile(f, data, 0644)
}

func (r *DistfilesReport) WriteYamlFile(f string) error {
	data, err := r.Yaml()
	if err != nil {
		return err
	}

	return os.WriteFile(f, data, 0644)
}

func KitDistfilesSyncCommand(config *specs.MarkDevkitConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "distfiles-sync",
		Aliases: []string{"ds", "distfiles"},
		Short:   "Sync distfiles for a list of kits.",
		PreRun: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()
			specfile, _ := cmd.Flags().GetString("specfile")

			if specfile == "" {
				log.Fatal("No specfile param defined.")
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()

			specfile, _ := cmd.Flags().GetString("specfile")
			to, _ := cmd.Flags().GetString("to")
			backend, _ := cmd.Flags().GetString("backend")
			verbose, _ := cmd.Flags().GetBool("verbose")
			skipGenReposcan, _ := cmd.Flags().GetBool("skip-reposcan-generation")
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			showSummary, _ := cmd.Flags().GetBool("show-summary")
			writeSummaryFile, _ := cmd.Flags().GetString(
				"write-summary-file")
			summaryFormat, _ := cmd.Flags().GetString("summary-format")
			keepWorkdir, _ := cmd.Flags().GetBool("keep-workdir")

			if showSummary {
				config.GetLogging().Level = "error"
			}

			log.InfoC(log.Aurora.Bold(
				fmt.Sprintf(":mask:Loading specfile %s", specfile)),
			)

			fetchOpts := kit.NewFetchOpts()
			fetchOpts.Concurrency = concurrency
			fetchOpts.Verbose = verbose
			fetchOpts.GenReposcan = !skipGenReposcan
			fetchOpts.CleanWorkingDir = !keepWorkdir

			fetcher, err := kit.NewFetcher(config, backend)
			if err != nil {
				log.Fatal(err.Error())
			}

			fetcher.SetWorkDir(to)
			err = fetcher.Sync(specfile, fetchOpts)
			if err != nil {
				log.Fatal(err.Error())
			}

			report := &DistfilesReport{
				Stats:         fetcher.GetStats(),
				AtomsInErrors: *fetcher.GetAtomsInError(),
			}

			if writeSummaryFile != "" {
				if summaryFormat == "json" {
					err = report.WriteJsonFile(writeSummaryFile)
				} else {
					err = report.WriteYamlFile(writeSummaryFile)
				}
				if err != nil {
					log.Fatal(err.Error())
				}
			}

			if showSummary {
				var data []byte

				if summaryFormat == "json" {
					data, err = report.Json()
				} else {
					data, err = report.Yaml()
				}
				if err != nil {
					log.Fatal(err.Error())
				}

				fmt.Println(string(data))
			} else {
				log.InfoC(log.Aurora.Bold(":party_popper:All done"))
			}
		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfiles of the jobs.")
	flags.String("to", "workdir", "Override default work directory.")
	flags.String("backend", "dir", "Set the fetcher backend to use: dir|s3.")
	flags.Int("concurrency", 3, "Define the elaboration concurrency.")
	flags.Bool("skip-reposcan-generation", false,
		"Skip reposcan files generation.")
	flags.Bool("verbose", false, "Show additional informations.")
	flags.Bool("show-summary", false, "Show YAML/JSON summary results")
	flags.String("write-summary-file", "",
		"Write the sync summary to the specified file in YAML/JSON format.")
	flags.String("summary-format", "yaml", "Specificy the summary format: json|yaml")
	flags.Bool("keep-workdir", false, "Avoid to remove the working directory.")

	return cmd
}
