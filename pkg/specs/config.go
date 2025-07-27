/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	rg "github.com/geaaru/rest-guard/pkg/specs"
	v "github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	MARKDEVKIT_CONFIGNAME = "mark-devkit"
	MARKDEVKIT_ENV_PREFIX = "MARKDEVKIT"
	MARKDEVKIT_VERSION    = `0.18.0`
)

type MarkDevkitConfig struct {
	Viper *v.Viper `yaml:"-" json:"-"`

	General        MarkDevkitGeneral        `mapstructure:"general" json:"general,omitempty" yaml:"general,omitempty"`
	Logging        MarkDevkitLogging        `mapstructure:"logging" json:"logging,omitempty" yaml:"logging,omitempty"`
	Authentication MarkDevkitAuthentication `mapstructure:"authentication" json:"authentication,omitempty" yaml:"authentication,omitempty"`
	RgConfig       *rg.RestGuardConfig      `mapstructure:"rest" json:"rest,omitempty" yaml:"rest,omitempty"`

	Storage map[string]interface{} `mapstructure:"-" json:"-" yaml:"-"`
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

type MarkDevkitAuthentication struct {
	Remotes map[string]*MarkDevkitRemoteAuth `mapstructure:"-,inline" json:"-,inline" yaml:"-,inline"`
}

type MarkDevkitRemoteAuth struct {
	Username   string `mapstructure:"username,omitempty" json:"username,omitempty" yaml:"username,omitempty"`
	Password   string `mapstructure:"password,omitempty" json:"password,omitempty" yaml:"password,omitempty"`
	Token      string `mapstructure:"token,omitempty" json:"token,omitempty" yaml:"token,omitempty"`
	ApiVersion string `mapstructure:"api_version,omitempty" json:"api_version,omitempty" yaml:"api_version,omitempty"`
	Url        string `mapstructure:"url,omitempty" json:"url,omitempty" yaml:"url,omitempty"`
}

func (a *MarkDevkitAuthentication) GetRemote(r string) (*MarkDevkitRemoteAuth, bool) {
	remote, present := a.Remotes[r]
	return remote, present
}

func NewMarkDevkitConfig(viper *v.Viper) *MarkDevkitConfig {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &MarkDevkitConfig{Viper: viper}
}

func (c *MarkDevkitConfig) GetAuthentication() *MarkDevkitAuthentication {
	return &c.Authentication
}

func (c *MarkDevkitConfig) GetGeneral() *MarkDevkitGeneral {
	return &c.General
}

func (c *MarkDevkitConfig) GetLogging() *MarkDevkitLogging {
	return &c.Logging
}

func (c *MarkDevkitConfig) GetRest() *rg.RestGuardConfig {
	if c.RgConfig == nil {
		c.RgConfig = rg.NewConfig()
	}
	return c.RgConfig
}

func (c *MarkDevkitConfig) GetStorage() *map[string]interface{} {
	return &c.Storage
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

	viper.SetDefault("rest.reqs_timeout", 3600)
	viper.SetDefault("rest.user_agent", "MARK Devkit Bot")
}

func (g *MarkDevkitGeneral) HasDebug() bool {
	return g.Debug
}
