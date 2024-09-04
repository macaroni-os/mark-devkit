/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmdkernel

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/macaroni-os/macaronictl/pkg/kernel"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewModulesCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "modules",
		Aliases: []string{"m"},
		Short:   "List available extra kernel modules to install.",
		Long: `Shows all extra kernel modules available in configured repositories.

$ macaronictl kernel modules

NOTE: It works only if the repositories are synced.
`,
		Run: func(cmd *cobra.Command, args []string) {

			jsonOutput, _ := cmd.Flags().GetBool("json")
			installed, _ := cmd.Flags().GetBool("installed")
			kBranch, _ := cmd.Flags().GetString("kernel-branch")
			kType, _ := cmd.Flags().GetString("kernel-type")

			stones, err := kernel.AvailableExtraModules(
				kBranch, kType, installed, config,
			)
			if err != nil {
				fmt.Println("Error on retrieve package list: " + err.Error())
				os.Exit(1)
			}

			if jsonOutput {
				data, err := json.Marshal(stones)
				if err != nil {
					fmt.Println(fmt.Errorf("Error on convert data to json: %s", err.Error()))
					os.Exit(1)
				}
				fmt.Println(string(data))
			} else {

				if len(stones.Stones) == 0 {
					fmt.Println(
						"No extra kernel modules available. Check repositories configurations and sync.")
					os.Exit(1)
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{
					Left: true, Top: false, Right: true, Bottom: false,
				})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"Package",
					"Package Version",
					"Kernel Branch",
					"Kernel Type",
					"Kernel Version",
					"Repository",
				})

				// Create response struct
				for _, s := range stones.Stones {

					kcat := "kernel-"
					version, ok := s.Labels["kernel.version"]
					if !ok {
						version = ""
					}
					ktype, ok := s.Labels["kernel.type"]
					if !ok {
						ktype = ""
					} else if ktype == "zen" {
						kcat = "kernel-zen-"
					}

					row := []string{
						s.GetName(),
						s.Version,
						strings.ReplaceAll(s.Category, kcat, ""),
						ktype,
						version,
						s.Repository,
					}

					table.Append(row)
				}

				table.Render()
			}

		},
	}

	flags := c.Flags()
	flags.Bool("json", false, "JSON output")
	flags.BoolP("installed", "i", false, "Show only installed modules. (Requires root permission)")
	flags.StringP("kernel-branch", "b", "", "Filter for a specific kernel branch.")
	flags.StringP("kernel-type", "t", "", "Filter for a specific kernel type.")

	return c
}
