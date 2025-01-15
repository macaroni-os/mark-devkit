/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/autogen"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
)

func autogenThinCmdCommand(config *specs.MarkDevkitConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "autogen-thin",
		Aliases: []string{"doit"},
		Short:   "Autogen specs without merging and/or sync.",
		Long:    `Executes minimal Autogen elaboration for testing purpose.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()
			specfile, _ := cmd.Flags().GetString("specfile")
			kitfile, _ := cmd.Flags().GetString("kitfile")

			if specfile == "" {
				log.Fatal("No specfile param defined.")
			}

			if kitfile == "" {
				log.Fatal("No kitfile param defined.")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()
			specfile, _ := cmd.Flags().GetString("specfile")
			kitfile, _ := cmd.Flags().GetString("kitfile")
			deep, _ := cmd.Flags().GetInt("deep")
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			to, _ := cmd.Flags().GetString("to")
			downloadDir, _ := cmd.Flags().GetString("download-dir")
			verbose, _ := cmd.Flags().GetBool("verbose")
			showValues, _ := cmd.Flags().GetBool("show-values")

			backendOpts := make(map[string]string, 0)

			log.InfoC(log.Aurora.Bold(
				fmt.Sprintf(":mask:Loading specfile %s", specfile)),
			)

			autogenOpts := autogen.NewAutogenBotOpts()
			autogenOpts.Concurrency = concurrency
			autogenOpts.GitDeepFetch = deep
			autogenOpts.Verbose = verbose
			autogenOpts.Push = false
			autogenOpts.PullRequest = false
			autogenOpts.SyncFiles = false
			autogenOpts.CleanWorkingDir = false
			autogenOpts.PullSources = false
			autogenOpts.GenReposcan = false
			autogenOpts.MergeAutogen = false
			autogenOpts.ShowGeneratedValues = showValues

			autogenBot := autogen.NewAutogenBot(config)
			autogenBot.SetWorkDir(to)
			if downloadDir != "" {
				autogenBot.SetDownloadDir(downloadDir)
			}
			err := autogenBot.SetupFetcher("dir", backendOpts)
			if err != nil {
				log.Fatal(err.Error())
			}

			err = autogenBot.Run(specfile, kitfile, autogenOpts)
			if err != nil {
				log.Fatal(err.Error())
			}

			log.InfoC(log.Aurora.Bold(":party_popper:All done"))
		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfile with the rules of the packages to autogen.")
	flags.StringP("kitfile", "k", "", "The YAML with the target kit definition.")
	flags.String("to", "workdir", "Override default work directory.")
	flags.String("download-dir", "", "Override the default ${workdir}/downloads directory.")
	flags.Bool("verbose", false, "Show additional informations.")
	flags.Int("deep", 5, "Define the limit of commits to fetch.")
	flags.Int("concurrency", 3, "Define the elaboration concurrency.")
	flags.Bool("show-values", false,
		"For debug purpose print generated values for any elaborated package in YAML format.")

	return cmd
}
