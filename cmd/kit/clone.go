/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmddiag

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

			opts := git.CloneOptions{
				SingleBranch: singleBranch,
				RemoteName:   "origin",
				Progress:     os.Stdout,
				Depth:        1,
			}

			for _, kit := range analysis.Kits {

				kitdir := filepath.Join(to, kit.Name)

				log.InfoC(log.Aurora.Bold(
					fmt.Sprintf(":factory:[%s] Syncing ...", kit.Name)),
				)
				err = kitops.Clone(kit, kitdir, opts, verbose)
				if err != nil {
					log.Fatal(err.Error())
				}
			}

			log.Info(":party_popper:Kits synced.")
		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfiles of the jobs.")
	flags.String("to", "output", "Target dir where sync kits.")
	flags.Bool("verbose", false, "Show additional informations.")
	flags.Bool("single-branch", true, "Pull only the used branch.")

	return cmd
}
