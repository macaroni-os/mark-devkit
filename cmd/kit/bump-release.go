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

func KitBumpReleaseCommand(config *specs.MarkDevkitConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "bump-release",
		Aliases: []string{"br", "bump", "release"},
		Short:   "Bump a new kits release.",
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
			to, _ := cmd.Flags().GetString("to")
			verbose, _ := cmd.Flags().GetBool("verbose")
			signatureName, _ := cmd.Flags().GetString("signature-name")
			signatureEmail, _ := cmd.Flags().GetString("signature-email")
			push, _ := cmd.Flags().GetBool("push")

			log.InfoC(log.Aurora.Bold(
				fmt.Sprintf(":mask:Loading specfile %s", specfile)),
			)

			releaseOpts := kit.NewReleaseOpts()
			releaseOpts.Verbose = verbose
			releaseOpts.GitDeepFetch = deep
			releaseOpts.SignatureName = signatureName
			releaseOpts.SignatureEmail = signatureEmail
			releaseOpts.Push = push

			releaseBot := kit.NewReleaseBot(config)
			releaseBot.SetWorkDir(to)

			err := releaseBot.Run(specfile, releaseOpts)
			if err != nil {
				log.Fatal(err.Error())
			}

			log.InfoC(log.Aurora.Bold(":party_popper:All done"))
		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfiles of the jobs.")
	flags.String("to", "workdir", "Override default work directory.")
	flags.Int("deep", 5, "Define the limit of commits to fetch.")
	flags.Bool("verbose", false, "Show additional informations.")
	flags.Bool("push", false, "Push commits to origin.")

	flags.String("signature-name", "", "Specify the name of the user for the commits.")
	flags.String("signature-email", "", "Specify the email of the user for the commits.")

	return cmd
}
