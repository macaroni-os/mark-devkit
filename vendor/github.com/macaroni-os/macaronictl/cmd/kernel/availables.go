/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmdkernel

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/macaroni-os/macaronictl/pkg/kernel"
	"github.com/macaroni-os/macaronictl/pkg/logger"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"

	"github.com/logrusorgru/aurora"
	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type KernelAvailable struct {
	*specs.Stone `json:"stone" yaml:"stone"`
	Annotation   *specs.KernelAnnotation `json:"kernel_data" yaml:"kernel_data"`
}

type KernelsAvailables struct {
	Kernels []*KernelAvailable `json:"kernels" yaml:"kernels"`
}

func NewAvailableCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "available",
		Aliases: []string{"availables", "a"},
		Short:   "List available kernels to install.",
		Long: `Shows kernels available in configured repositories.

$ macaronictl kernel availables

NOTE: It works only if the repositories are synced.
`,
		Run: func(cmd *cobra.Command, args []string) {

			log := logger.GetDefaultLogger()
			jsonOutput, _ := cmd.Flags().GetBool("json")
			lts, _ := cmd.Flags().GetBool("lts")

			stones, err := kernel.AvailableKernels(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			stonesInstalled, err := kernel.InstalledKernels(config)
			if err != nil {
				fmt.Println("Error on retrieve installed kernel: " + err.Error())
				os.Exit(1)
			}

			kimap := make(map[string]bool)
			for _, s := range stonesInstalled.Stones {
				kimap[s.HumanReadableString()] = true
			}
			stonesInstalled = nil

			kernels := &KernelsAvailables{
				Kernels: []*KernelAvailable{},
			}

			kMap := make(map[string]bool, 0)

			// Create response struct
			for _, s := range stones.Stones {

				if _, ok := kMap[s.HumanReadableString()]; ok {
					// The stone is already been catched from
					// another repo. I take only one time a specific kernel.
					continue
				}

				kMap[s.HumanReadableString()] = true

				a, err := kernel.ParseKernelAnnotations(s)
				if err != nil {
					log.Warning("[%s/%s] Error on parse annotation: %s",
						s.Category, s.Name, err.Error())
					continue
				}

				if lts && !a.Lts {
					continue
				}

				kernels.Kernels = append(kernels.Kernels,
					&KernelAvailable{
						Stone:      s,
						Annotation: a,
					},
				)
			}

			if !jsonOutput {

				if len(kernels.Kernels) == 0 {
					fmt.Println("No kernels availables. Check repositories configurations and sync.")
					os.Exit(1)
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{
					Left: true, Top: false, Right: true, Bottom: false,
				})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"Kernel",
					"Kernel Version",
					"Package Version",
					"Eol",
					"LTS",
					"Released",
					"Type",
				})

				for _, k := range kernels.Kernels {
					ltsstr := "false"
					if k.Annotation.Lts {
						ltsstr = "true"
					}

					pversion := k.Stone.Version
					version, ok := k.Stone.Labels["package.version"]
					if !ok {
						version = ""
					}

					_, installed := kimap[k.Stone.HumanReadableString()]

					if installed {
						version = fmt.Sprintf("%s", aurora.Bold(version))
						pversion = fmt.Sprintf("%s", aurora.Bold(pversion))
					}

					row := []string{
						k.Annotation.Suffix,
						version,
						pversion,
						k.Annotation.EoL,
						ltsstr,
						k.Annotation.Released,
						k.Annotation.Type,
					}

					table.Append(row)
				}

				table.Render()

			} else {
				data, err := json.Marshal(kernels)
				if err != nil {
					fmt.Println(fmt.Errorf("Error on convert data to json: %s", err.Error()))
					os.Exit(1)
				}
				fmt.Println(string(data))
			}

		},
	}

	flags := c.Flags()
	flags.Bool("json", false, "JSON output")
	flags.Bool("lts", false, "Show only LTS kernels.")

	return c
}
