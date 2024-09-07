/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmdmetro

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/executor"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/metro"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
)

func RunJobCommand(config *specs.MarkDevkitConfig) *cobra.Command {

	var cmd = &cobra.Command{
		Use:     "run",
		Aliases: []string{"r", "j"},
		Short:   "Run one or more job.",
		PreRun: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()
			specfile, _ := cmd.Flags().GetString("specfile")
			job, _ := cmd.Flags().GetString("job")

			if specfile == "" {
				log.Fatal("No specfile param defined.")
			}
			if job == "" {
				log.Fatal("No job defined.")
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()

			specfile, _ := cmd.Flags().GetString("specfile")
			job, _ := cmd.Flags().GetString("job")
			cleanup, _ := cmd.Flags().GetBool("cleanup")
			cpu, _ := cmd.Flags().GetString("cpu")
			fchrootDebug, _ := cmd.Flags().GetBool("fchroot-debug")
			quiet, _ := cmd.Flags().GetBool("quiet")
			skipSourcePhase, _ := cmd.Flags().GetBool("skip-source-phase")
			skipPackerPhase, _ := cmd.Flags().GetBool("skip-packer-phase")
			skipHooksPhase, _ := cmd.Flags().GetBool("skip-hooks-phase")

			log.Info(fmt.Sprintf(":guard:Loading specfile %s", specfile))

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

			fopts := executor.NewFchrootOpts()
			fopts.Verbose = true
			fopts.Debug = fchrootDebug
			fopts.Cpu = cpu

			ropts := &metro.RunOpts{
				CleanupRootfs: cleanup,
				Quiet:         quiet,
				SkipSource:    skipSourcePhase,
				SkipPacker:    skipPackerPhase,
				SkipHooks:     skipHooksPhase,
				Opts:          fopts,
			}

			err = m.RunJob(jrender, ropts)
			if err != nil {
				log.Fatal(err.Error())
			}

			log.Info(fmt.Sprintf(":party_popper:Job %s completed.", job))
		},
	}

	flags := cmd.Flags()
	flags.String("specfile", "", "The specfiles of the jobs.")
	flags.String("job", "", "The job to diagnose.")
	flags.String("cpu", "", "Specify specific CPU type for QEMU to use")
	flags.Bool("cleanup", true, "Cleanup rootfs directory.")
	flags.Bool("skip-source-phase", false, "Skip source phase.")
	flags.Bool("skip-packer-phase", false, "Skip packer phase.")
	flags.Bool("skip-hooks-phase", false,
		"Skip hooks executions. For development only.")
	flags.Bool("fchroot-debug", false, "Enable debug on fchroot.")
	flags.Bool("quiet", false, "Avoid to see the hooks command output.")

	return cmd
}
