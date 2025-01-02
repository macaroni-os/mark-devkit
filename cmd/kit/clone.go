/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmdkit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	kitops "github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

func generateReposcanFiles(ra *specs.ReposcanAnalysis,
	sourceDir, kitCacheDir string, concurrency int,
	verbose bool) error {

	log := logger.GetDefaultLogger()
	eclassDirs, err := ra.GetKitsEclassDirs(sourceDir)
	if err != nil {
		return err
	}

	err = helpers.EnsureDirWithoutIds(kitCacheDir, 0755)
	if err != nil {
		return err
	}

	for _, source := range ra.Kits {
		sourceDir := filepath.Join(sourceDir, source.Name)
		targetFile := filepath.Join(kitCacheDir, source.Name+"-"+source.Branch)

		log.Debug(fmt.Sprintf("Generating kit-cache file for kit %s...",
			source.Name))

		err = kitops.RunReposcanGenerate(sourceDir, source.Name, source.Branch, targetFile,
			eclassDirs, concurrency, verbose)
		if err != nil {
			return err
		}
	}

	return nil
}

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
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			showSummary, _ := cmd.Flags().GetBool("show-summary")
			generateReposcan, _ := cmd.Flags().GetBool("generate-reposcan-files")
			reposcanDir, _ := cmd.Flags().GetString("kit-cache-dir")
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

			if generateReposcan {
				err = generateReposcanFiles(analysis, to, reposcanDir,
					concurrency, !showSummary)
				if err != nil {
					log.Fatal(err.Error())
				}
			}

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
	flags.Int("concurrency", 3, "Define the elaboration concurrency.")
	flags.Bool("show-summary", false, "Show YAML summary results")
	flags.Bool("generate-reposcan-files", false, "Generate reposcan files of the pulled kits.")
	flags.String("kit-cache-dir", "kit-cache", "Directory where generate reposcan files.")
	flags.Int("deep", 5, "Define the limit of commits to fetch.")
	flags.String("write-summary-file", "",
		"Write the sync summary to the specified file in YAML format.")

	return cmd
}
