/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmddiag

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/metro"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
)

func DiagnoseJobCommand(config *specs.MarkDevkitConfig) *cobra.Command {

	var cmd = &cobra.Command{
		Use:     "job",
		Aliases: []string{"j"},
		Short:   "Render Jobs for debug.",
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()

			specfile, _ := cmd.Flags().GetString("specfile")
			job, _ := cmd.Flags().GetString("job")

			log.InfoC(log.Aurora.Bold(
				fmt.Sprintf(":mask:Loading specfile %s", specfile)),
			)

			m := metro.NewMetro(config)
			mspec, err := m.Load(specfile)
			if err != nil {
				log.Fatal(err.Error())
			}

			j := mspec.GetJob(job)
			if j == nil {
				log.Fatal(fmt.Sprintf("No job with name %s found",
					job))
			}

			jrender, err := j.Render(specfile)
			if err != nil {
				log.Fatal(err.Error())
			}

			data, err := jrender.Yaml()
			if err != nil {
				log.Fatal(err.Error())
			}

			fmt.Println(string(data))
		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfiles of the jobs.")
	flags.String("job", "", "The job to diagnose.")

	return cmd
}
