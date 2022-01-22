package types

import (
	"errors"
	"time"
)

// PluginConfig struct
type PluginConfig struct {
	LoadClient     bool          `yaml:"load_client"`
	LoadFromConfig bool          `yaml:"load_from_config"`
	ConfigFile     string        `yaml:"config_file"`
	Server         string        `yaml:"server"`
	Username       string        `yaml:"username"`
	Password       string        `yaml:"password"`
	Token          string        `yaml:"token"`
	Insecure       bool          `yaml:"insecure"`
	TimeoutConfig  TimeoutConfig `yaml:"timeout_config"`
}

func (p *PluginConfig) Validate() error {
	if p.Server == "" {
		return errors.New("server url can not be empty")
	}
	if p.Username == "" {
		return errors.New("username can not be empty")
	}

	if p.Token == "" && p.Password == "" {
		return errors.New("either token or password can not be empty")
	}
	return nil
}

// ProviderConfig struct
type ProviderConfig struct {
	Kind   string        `yaml:"kind"`
	Action string        `yaml:"function"`
	Config ActionConfig  `yaml:"config"`
	Data   []interface{} `yaml:"data"`
}

type ActionConfig struct {
	Recreate      bool          `yaml:"recreate"`
	OnFailure     string        `yaml:"on_failure"`
	TimeoutConfig TimeoutConfig `yaml:"timeout_config"`
}

type TimeoutConfig struct {
	Timeout              time.Duration `yaml:"timeout"`
	ScanInterval         time.Duration `yaml:"scan_interval"`
	ExpectedSuccessCount int           `yaml:"success_count"`
}

func (tc *TimeoutConfig) UpdateDefaults() {
	if tc.Timeout == 0 {
		tc.Timeout = time.Minute * 3
	}
	if tc.ScanInterval == 0 {
		tc.ScanInterval = time.Second * 10
	}

	if tc.ExpectedSuccessCount == 0 {
		tc.ExpectedSuccessCount = 3
	}
}
