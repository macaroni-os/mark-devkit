/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cliName = `Copyright (c) 2024-2025 Macaroni OS - Daniele Rondina

M.A.R.K. Development Tool

Distributed under the terms of the GNU General Public License version 3
this tool comes with ABSOLUTELY NO WARRANTY; This is free software, and you
are welcome to redistribute it under certain conditions.
`
)

var (
	BuildTime   string
	BuildCommit string
)

func initConfig(config *specs.MarkDevkitConfig) {
	// Set env variable
	config.Viper.SetEnvPrefix(specs.MARKDEVKIT_ENV_PREFIX)
	config.Viper.BindEnv("config")
	config.Viper.SetDefault("config", "")
	config.Viper.SetDefault("etcd-config", false)

	config.Viper.AutomaticEnv()

	// Create EnvKey Replacer for handle complex structure
	replacer := strings.NewReplacer(".", "__", "-", "_")
	config.Viper.SetEnvKeyReplacer(replacer)

	// Set config file name (without extension)
	config.Viper.SetConfigName(specs.MARKDEVKIT_CONFIGNAME)

	config.Viper.SetTypeByDefaultValue(true)
}

func initCommand(rootCmd *cobra.Command, config *specs.MarkDevkitConfig) {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "", "MARK Devkit configuration file")
	pflags.BoolP("debug", "d", config.Viper.GetBool("general.debug"),
		"Enable debug output.")

	config.Viper.BindPFlag("config", pflags.Lookup("config"))
	config.Viper.BindPFlag("general.debug", pflags.Lookup("debug"))

	rootCmd.AddCommand(
		autogenCmdCommand(config),
		autogenThinCmdCommand(config),
		metroCmdCommand(config),
		diagnoseCmdCommand(config),
		kitCmdCommand(config),
	)
}

func Execute() {
	// Create Main Instance Config object
	var config *specs.MarkDevkitConfig = specs.NewMarkDevkitConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        cliName,
		Version:      fmt.Sprintf("%s-g%s %s", specs.MARKDEVKIT_VERSION, BuildCommit, BuildTime),
		Args:         cobra.OnlyValidArgs,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			var v *viper.Viper = config.Viper

			v.SetConfigType("yml")
			if v.Get("config") == "" {
				config.Viper.AddConfigPath(".")
			} else {
				v.SetConfigFile(v.Get("config").(string))
			}

			// Parse configuration file
			err = config.Unmarshal()
			if err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					// Config file not found; ignore error if desired
				} else {
					fmt.Println(err)
					os.Exit(1)
				}
			}

			// Initialize logger
			log := logger.NewMarkDevkitLogger(config)
			log.SetAsDefault()
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
