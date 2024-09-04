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
	MACARONICTL_CONFIGNAME = "macaronios"
	MACARONICTL_ENV_PREFIX = "MACARONICTL"
)

type MacaroniCtlConfig struct {
	Viper *v.Viper `yaml:"-" json:"-"`

	General   MacaroniCtlGeneral   `mapstructure:"general" json:"general,omitempty" yaml:"general,omitempty"`
	Logging   MacaroniCtlLogging   `mapstructure:"logging" json:"logging,omitempty" yaml:"logging,omitempty"`
	EnvUpdate MacaroniCtlEnvUpdate `mapstructure:"env-update,omitempty" json:"env-update,omitempty" yaml:"env-update,omitempty"`

	KernelProfilesDir string `mapstructure:"kernel-profiles-dir,omitempty" json:"kernel-profiles-dir,omitempty" yaml:"kernel-profiles-dir,omitempty"`
}

type MacaroniCtlGeneral struct {
	Debug bool `mapstructure:"debug,omitempty" json:"debug,omitempty" yaml:"debug,omitempty"`
}

type MacaroniCtlEnvUpdate struct {
	Ldconfig bool `mapstructure:"ldconfig,omitempty" json:"ldconfig,omitempty" yaml:"ldconfig,omitempty"`
	Csh      bool `mapstructure:"csh,omitempty" json:"csh,omitempty" yaml:"csh,omitempty"`
	Prelink  bool `mapstructure:"prelink,omitempty" json:"prelink,omitempty" yaml:"prelink,omitempty"`
	Systemd  bool `mapstructure:"systemd,omitempty" json:"systemd,omitempty" yaml:"systemd,omitempty"`
}

type MacaroniCtlLogging struct {
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

func NewMacaroniCtlConfig(viper *v.Viper) *MacaroniCtlConfig {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &MacaroniCtlConfig{Viper: viper}
}

func (c *MacaroniCtlConfig) GetGeneral() *MacaroniCtlGeneral {
	return &c.General
}

func (c *MacaroniCtlConfig) GetLogging() *MacaroniCtlLogging {
	return &c.Logging
}

func (c *MacaroniCtlConfig) GetEnvUpdate() *MacaroniCtlEnvUpdate {
	return &c.EnvUpdate
}

func (c *MacaroniCtlConfig) Unmarshal() error {
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

func (c *MacaroniCtlConfig) Yaml() ([]byte, error) {
	return yaml.Marshal(c)
}

func GenDefault(viper *v.Viper) {
	viper.SetDefault("general.debug", false)
	viper.SetDefault("kernel-profiles-dir", "/etc/macaroni/kernels-profiles/")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.enable_logfile", false)
	viper.SetDefault("logging.path", "/var/log/macaroni/macaronictl.log")
	viper.SetDefault("logging.json_format", false)
	viper.SetDefault("logging.enable_emoji", true)
	viper.SetDefault("logging.color", true)

	viper.SetDefault("env-update.csh", false)
	viper.SetDefault("env-update.ldconfig", true)
	viper.SetDefault("env-update.systemd", false)
	viper.SetDefault("env-update.prelink", false)
}

func (g *MacaroniCtlGeneral) HasDebug() bool {
	return g.Debug
}

func (c *MacaroniCtlConfig) GetKernelProfilesDir() string { return c.KernelProfilesDir }
