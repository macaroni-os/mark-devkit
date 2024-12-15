/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmddiag

import (
	"fmt"
	"os"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	kitops "github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

func KitCloneCommand(config *specs.MarkDevkitConfig) *cobra.Command {

	var cmd = &cobra.Command{
		Use:     "clone",
		Aliases: []string{"c", "cl", "sync"},
		Short:   "Clone/Sync kits locally from a YAML specs rules file.",
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
			verbose, _ := cmd.Flags().GetBool("verbose")
			singleBranch, _ := cmd.Flags().GetBool("single-branch")
			deep, _ := cmd.Flags().GetInt("deep")
			showSummary, _ := cmd.Flags().GetBool("show-summary")
			writeSummaryFile, _ := cmd.Flags().GetString(
				"write-summary-file")

			if showSummary {
				config.GetLogging().Level = "error"
			}

			log.InfoC(log.Aurora.Bold(
				fmt.Sprintf(":mask:Loading specfile %s", specfile)),
			)

			analysis, err := specs.NewReposcanAnalysis(specfile)
			if err != nil {
				log.Fatal(err.Error())
			}

			if len(analysis.Kits) == 0 {
				log.InfoC(":warning:No kits to sync defined in the specfile.")
				os.Exit(1)
			}

			err = helpers.EnsureDirWithoutIds(to, 0740)
			if err != nil {
				log.Fatal(err.Error())
			}

			opts := &kitops.CloneOptions{
				GitCloneOptions: &git.CloneOptions{
					SingleBranch: singleBranch,
					RemoteName:   "origin",
					Depth:        deep,
				},
				Verbose: verbose,
				Summary: showSummary || writeSummaryFile != "",
				Results: []*specs.ReposcanKit{},
			}

			if !showSummary {
				opts.GitCloneOptions.Progress = os.Stdout
			}

			err = kitops.CloneKits(analysis, to, opts)
			if err != nil {
				log.Fatal(err.Error())
			}

			log.Info(":party_popper:Kits synced.")

			if showSummary || writeSummaryFile != "" {
				summary := &specs.ReposcanAnalysis{
					Kits: opts.Results,
				}

				if writeSummaryFile != "" {
					err = summary.WriteYamlFile(writeSummaryFile)
					if err != nil {
						log.Fatal(err.Error())
					}
				}

				if showSummary {
					data, _ := summary.Yaml()
					fmt.Println(string(data))
				}
			}

		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfiles of the jobs.")
	flags.String("to", "output", "Target dir where sync kits.")
	flags.Bool("verbose", false, "Show additional informations.")
	flags.Bool("single-branch", true, "Pull only the used branch.")
	flags.Bool("show-summary", false, "Show YAML summary results")
	flags.Int("deep", 5, "Define the limit of commits to fetch.")
	flags.String("write-summary-file", "",
		"Write the sync summary to the specified file in YAML format.")

	return cmd
}
