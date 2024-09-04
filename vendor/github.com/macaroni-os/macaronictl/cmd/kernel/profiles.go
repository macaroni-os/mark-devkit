/*
	Copyright © 2021 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmdkernel

import (
	"encoding/json"
	"fmt"
	"os"

	kernelspecs "github.com/macaroni-os/macaronictl/pkg/kernel/specs"
	"github.com/macaroni-os/macaronictl/pkg/profile"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewProfilesCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "profiles",
		Aliases: []string{"p"},
		Short:   "List available kernels profiles.",
		Long: `Shows kernels available in your system

$ macaronictl kernel profiles

`,
		Run: func(cmd *cobra.Command, args []string) {

			jsonOutput, _ := cmd.Flags().GetBool("json")
			kernelProfilesDir, _ := cmd.Flags().GetString("kernel-profiles-dir")

			types := []kernelspecs.KernelType{}
			if kernelProfilesDir != "" {
				types, _ = profile.LoadKernelProfiles(kernelProfilesDir)
			} else {
				types, _ = profile.LoadKernelProfiles(config.GetKernelProfilesDir())
			}
			if len(types) == 0 {
				types = profile.GetDefaultKernelProfiles()
			}

			if jsonOutput {
				data, err := json.Marshal(types)
				if err != nil {
					fmt.Println(fmt.Errorf("Error on convert data to json: %s", err.Error()))
					os.Exit(1)
				}
				fmt.Println(string(data))

			} else {

				if len(types) == 0 {
					fmt.Println("No kernel profiles availables. I will use default profiles.")
					os.Exit(0)
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{
					Left: true, Top: false, Right: true, Bottom: false,
				})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"Name",
					"Kernel Prefix",
					"Initrd Prefix",
					"Suffix",
					"Type",
					"With Arch",
				})

				for _, kt := range types {

					table.Append([]string{
						kt.GetName(),
						kt.GetKernelPrefixSanitized(),
						kt.GetInitrdPrefixSanitized(),
						kt.GetSuffix(),
						kt.GetType(),
						fmt.Sprintf("%v", kt.WithArch),
					})
				}

				table.Render()
			}
		},
	}

	flags := c.Flags()
	flags.Bool("json", false, "JSON output")
	flags.String("kernel-profiles-dir", "",
		"Specify the directory where read the kernel types profiles supported.")

	return c
}
