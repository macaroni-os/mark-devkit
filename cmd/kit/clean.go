/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmdkit

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
)

func KitCleanCommand(config *specs.MarkDevkitConfig) *cobra.Command {

	var cmd = &cobra.Command{
		Use:     "clean",
		Aliases: []string{"purge", "c"},
		Short:   "Clean old packages from kits.",
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
			keepWorkdir, _ := cmd.Flags().GetBool("keep-workdir")
			push, _ := cmd.Flags().GetBool("push")
			githubUser, _ := cmd.Flags().GetString("github-user")
			pr, _ := cmd.Flags().GetBool("pr")
			atoms, _ := cmd.Flags().GetStringArray("pkg")

			log.InfoC(log.Aurora.Bold(
				fmt.Sprintf(":mask:Loading specfile %s", specfile)),
			)

			mergeOpts := kit.NewMergeBotOpts()
			mergeOpts.Concurrency = concurrency
			mergeOpts.GitDeepFetch = deep
			mergeOpts.PullSources = !skipPullSources
			mergeOpts.GenReposcan = !skipGenReposcan
			mergeOpts.Push = push
			mergeOpts.PullRequest = pr
			mergeOpts.Verbose = verbose
			mergeOpts.SignatureName = signatureName
			mergeOpts.SignatureEmail = signatureEmail
			mergeOpts.CleanWorkingDir = !keepWorkdir
			mergeOpts.Atoms = atoms

			if githubUser != "" {
				mergeOpts.GithubUser = githubUser
			}

			mergeBot := kit.NewMergeBot(config)
			mergeBot.SetWorkDir(to)

			err := mergeBot.Clean(specfile, mergeOpts)
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
	flags.Bool("push", false, "Push commits to origin.")
	flags.Bool("keep-workdir", false, "Avoid to remove the working directory.")
	flags.Bool("pr", false, "Push commit over specific branch and as Pull Request.")
	flags.StringArray("pkg", []string{}, "Elaborate only specified packages.")

	flags.String("signature-name", "", "Specify the name of the user for the commits.")
	flags.String("signature-email", "", "Specify the email of the user for the commits.")
	flags.String("github-user", "", "Override the default Github user used for PR.")

	return cmd
}
