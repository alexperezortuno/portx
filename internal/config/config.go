package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Tunnels  []TunnelConfig `mapstructure:"tunnels"`
	Provider string         `mapstructure:"provider"`
	LogLevel string         `mapstructure:"log_level"`
}

type TunnelConfig struct {
	Name         string `mapstructure:"name"`
	Provider     string `mapstructure:"provider"`
	LocalAddr    string `mapstructure:"local_addr"`
	RemoteAddr   string `mapstructure:"remote_addr"`
	SSHUser      string `mapstructure:"ssh_user"`
	SSHH         string `mapstructure:"ssh_host"`
	SSHPort      int    `mapstructure:"ssh_port"`
	SSHPKey      string `mapstructure:"ssh_private_key"`
	SSHPassword  string `mapstructure:"ssh_password"`
	PortXDServer string `mapstructure:"portxd_server"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("$HOME/.portx")
		v.AddConfigPath(".")
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if configPath == "" {
		err := v.ReadInConfig()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	} else {
		err := v.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("reading config file %s: %w", configPath, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.LogLevel != "" {
		level := strings.ToLower(c.LogLevel)
		switch level {
		case "debug", "info", "warn", "error":
		default:
			return fmt.Errorf("invalid log_level: %s", c.LogLevel)
		}
	}
	for i, t := range c.Tunnels {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("tunnel[%d]: %w", i, err)
		}
	}
	return nil
}

func (t *TunnelConfig) Validate() error {
	if t.LocalAddr == "" {
		return fmt.Errorf("local_addr is required")
	}
	return nil
}
