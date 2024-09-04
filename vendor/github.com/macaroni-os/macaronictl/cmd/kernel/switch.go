/*
Copyright © 2021-2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmdkernel

import (
	"fmt"
	"os"
	"strings"

	"github.com/macaroni-os/macaronictl/pkg/kernel"
	"github.com/macaroni-os/macaronictl/pkg/logger"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"

	"github.com/spf13/cobra"
)

func NewSwitchCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "switch <kernel>@<kernel-branch> [OPTIONS]",
		Aliases: []string{"s"},
		Short:   "Switch to a specified kernel.",
		Long: `Switch an installed kernel from a branch to another.

$ macaronictl kernel switch macaroni@6.1 --purge

$ macaronictl kernel switch macaroni@6.1 --from 5.15

NOTE: It works only if the repositories are synced and the branch
      is not yet installed.
      Please, use --purge carefully. Often on switch it's better
      to maintain the old until the new is been verified.
      This command requires root privilege.
`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				fmt.Println("More of one kernel defined. Only one is accepted.")
				os.Exit(1)
			} else if len(args) == 0 {
				fmt.Println("Missing mandatory argument.")
				os.Exit(1)
			}

			if strings.Index(args[0], "@") < 0 {
				fmt.Println("Malformed argument.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			kType, _ := cmd.Flags().GetString("type")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			from, _ := cmd.Flags().GetString("from")
			fromType, _ := cmd.Flags().GetString("from-type")

			// Parse input argument
			param := args[0]
			requiredKernel := param[0:strings.Index(param, "@")]
			requiredBranch := param[strings.Index(param, "@")+1:]

			log := logger.GetDefaultLogger()

			// Retrieve the installed kernels
			installed, err := kernel.InstalledKernels(config)
			if err != nil {
				fmt.Println("Error on retrieve installed kernels: " + err.Error())
				os.Exit(1)
			}

			for _, s := range installed.Stones {
				a, err := kernel.ParseKernelAnnotations(s)
				if err != nil {
					log.Error("[%s/%s] Error on parse annotation: %s",
						s.Category, s.Name, err.Error())
					os.Exit(1)
				}

				if a.Suffix != requiredKernel {
					// The required kernel is different. Go ahead.
					continue
				}

				if a.Type != kType {
					continue
				}

				if s.Category == "kernel-"+requiredBranch {
					log.Error(fmt.Sprintf(
						"The kernel %s and branch %s is already installed.",
						requiredKernel, requiredBranch,
					))
					os.Exit(1)
				}
			}

			available, err := kernel.AvailableKernels(config)
			if err != nil {
				fmt.Println("Error on retrieve available kernels: " + err.Error())
				os.Exit(1)
			}
			log.Debug(fmt.Sprintf(
				"Found %d available kernels.", len(available.Stones)))

			var candidate *specs.Stone = nil
			for _, s := range available.Stones {
				a, err := kernel.ParseKernelAnnotations(s)
				if err != nil {
					log.Error("[%s/%s] Error on parse annotation: %s",
						s.Category, s.Name, err.Error())
					os.Exit(1)
				}

				if a.Suffix != requiredKernel {
					continue
				}

				if a.Type != kType {
					continue
				}

				if s.Category == "kernel-"+requiredBranch {
					candidate = s
					break
				}
			}

			if candidate == nil {
				fmt.Println("No valid kernel candidate found.")
				os.Exit(1)
			}

			// Retrieve installed extra modules.
			availableInstMods, err := kernel.AvailableExtraModules(
				from, fromType, true, config,
			)
			if err != nil {
				fmt.Println("Error on retrieve installed kernel modules: " + err.Error())
				os.Exit(1)
			}

			kextraModsMap := make(map[string]*specs.Stone, 0)
			sourceCategoryPrefix := "kernel-"
			switch fromType {
			case "zen":
				sourceCategoryPrefix = "kernel-zen-"
			}

			// Prepare map of all installed module
			for _, s := range availableInstMods.Stones {
				if from != "" && s.Category != sourceCategoryPrefix+from {
					continue
				}
				kextraModsMap[s.Name] = s
				log.Debug("Found module", s.Name)
			}

			availableModules, err := kernel.AvailableExtraModules(
				requiredBranch, kType, false, config,
			)
			if err != nil {
				fmt.Println("Error on retrieve available kernel modules: " + err.Error())
				os.Exit(1)
			}

			candidateModules := []*specs.Stone{}
			for _, s := range availableModules.Stones {
				if _, present := kextraModsMap[s.Name]; present {
					candidateModules = append(candidateModules, s)
				}
			}

			fmt.Println(fmt.Sprintf(
				"Found kernel candidate %s...",
				candidate.HumanReadableString()))

			if len(candidateModules) > 0 {
				fmt.Println("Modules extra to install:")
				for _, m := range candidateModules {
					fmt.Println("- " + m.HumanReadableString())
				}
			}

			if !dryRun {
				err = kernel.InstallPackages(candidate, candidateModules)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
			}

		},
	}

	flags := c.Flags()
	flags.Bool("purge", false, "Purge the installed kernels.")
	flags.Bool("dry-run", false, "Dry run installation and show candidates.")
	flags.String("type", "vanilla", "Define the kernel type to use.")
	flags.String("from", "", "Define the kernel branch to replace.")
	flags.String("from-type", "vanilla", "Define the type of the kernel used to retrieve the list of installed modules.")

	return c
}
