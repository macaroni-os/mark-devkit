/*
Copyright © 2021 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	v "github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

const (
	MARKDEVKIT_CONFIGNAME = "mark-devkit"
	MARKDEVKIT_ENV_PREFIX = "MARKDEVKIT"
)

type MarkDevkitConfig struct {
	Viper *v.Viper `yaml:"-" json:"-"`

	General MarkDevkitGeneral `mapstructure:"general" json:"general,omitempty" yaml:"general,omitempty"`
	Logging MarkDevkitLogging `mapstructure:"logging" json:"logging,omitempty" yaml:"logging,omitempty"`
}

type MarkDevkitGeneral struct {
	Debug bool `mapstructure:"debug,omitempty" json:"debug,omitempty" yaml:"debug,omitempty"`
}

type MarkDevkitLogging struct {
	// Path of the logfile
	Path string `mapstructure:"path,omitempty" json:"path,omitempty" yaml:"path,omitempty"`
	// Enable/Disable logging to file
	EnableLogFile bool `mapstructure:"enable_logfile,omitempty" json:"enable_logfile,omitempty" yaml:"enable_logfile,omitempty"`
	// Enable JSON format logging in file
	JsonFormat bool `mapstructure:"json_format,omitempty" json:"json_format,omitempty" yaml:"json_format,omitempty"`

	// Log level
	Level string `mapstructure:"level,omitempty" json:"level,omitempty" yaml:"level,omitempty"`

	// Enable emoji
	EnableEmoji bool `mapstructure:"enable_emoji,omitempty" json:"enable_emoji,omitempty" yaml:"enable_emoji,omitempty"`
	// Enable/Disable color in logging
	Color bool `mapstructure:"color,omitempty" json:"color,omitempty" yaml:"color,omitempty"`
}

func NewMarkDevkitConfig(viper *v.Viper) *MarkDevkitConfig {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &MarkDevkitConfig{Viper: viper}
}

func (c *MarkDevkitConfig) GetGeneral() *MarkDevkitGeneral {
	return &c.General
}

func (c *MarkDevkitConfig) GetLogging() *MarkDevkitLogging {
	return &c.Logging
}

func (c *MarkDevkitConfig) Unmarshal() error {
	var err error

	if c.Viper.InConfig("etcd-config") &&
		c.Viper.GetBool("etcd-config") {
		err = c.Viper.ReadRemoteConfig()
	} else {
		err = c.Viper.ReadInConfig()
	}

	if err != nil {
		if _, ok := err.(v.ConfigFileNotFoundError); !ok {
			return err
		}
		// else: Config file not found; ignore error
	}

	err = c.Viper.Unmarshal(&c)

	return err
}

func (c *MarkDevkitConfig) Yaml() ([]byte, error) {
	return yaml.Marshal(c)
}

func GenDefault(viper *v.Viper) {
	viper.SetDefault("general.debug", false)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.enable_logfile", false)
	viper.SetDefault("logging.path", "/var/log/macaroni/mark-devkit.log")
	viper.SetDefault("logging.json_format", false)
	viper.SetDefault("logging.enable_emoji", true)
	viper.SetDefault("logging.color", true)
}

func (g *MarkDevkitGeneral) HasDebug() bool {
	return g.Debug
}
