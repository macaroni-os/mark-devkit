/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmddiag

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
)

func KitMergeCommand(config *specs.MarkDevkitConfig) *cobra.Command {

	var cmd = &cobra.Command{
		Use:     "merge",
		Aliases: []string{"m", "me"},
		Short:   "Merge packages between kits.",
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
			deep, _ := cmd.Flags().GetInt("deep")
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			skipPullSources, _ := cmd.Flags().GetBool("skip-pull-sources")
			skipGenReposcan, _ := cmd.Flags().GetBool("skip-reposcan-generation")
			to, _ := cmd.Flags().GetString("to")
			signatureName, _ := cmd.Flags().GetString("signature-name")
			signatureEmail, _ := cmd.Flags().GetString("signature-email")
			verbose, _ := cmd.Flags().GetBool("verbose")

			log.InfoC(log.Aurora.Bold(
				fmt.Sprintf(":mask:Loading specfile %s", specfile)),
			)

			mergeOpts := kit.NewMergeBotOpts()
			mergeOpts.Concurrency = concurrency
			mergeOpts.GitDeepFetch = deep
			mergeOpts.PullSources = !skipPullSources
			mergeOpts.GenReposcan = !skipGenReposcan
			mergeOpts.Verbose = verbose
			mergeOpts.SignatureName = signatureName
			mergeOpts.SignatureEmail = signatureEmail

			mergeBot := kit.NewMergeBot(config)
			mergeBot.SetWorkDir(to)

			err := mergeBot.Run(specfile, mergeOpts)
			if err != nil {
				log.Fatal(err.Error())
			}

			log.InfoC(log.Aurora.Bold(":party_popper:All done"))
		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfiles of the jobs.")
	flags.String("to", "workdir", "Override default work directory.")
	flags.Bool("verbose", false, "Show additional informations.")
	flags.Int("deep", 5, "Define the limit of commits to fetch.")
	flags.Int("concurrency", 3, "Define the elaboration concurrency.")
	flags.Bool("skip-reposcan-generation", false,
		"Skip reposcan files generation.")
	flags.Bool("skip-pull-sources", false,
		"Skip pull of sources repositories.")

	flags.String("signature-name", "", "Specify the name of the user for the commits.")
	flags.String("signature-email", "", "Specify the email of the user for the commits.")

	return cmd
}
