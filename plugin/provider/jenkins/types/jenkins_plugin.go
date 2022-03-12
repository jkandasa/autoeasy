package jenkins_provider

import "time"

// plugin configuration
type PluginConfig struct {
	ServerURL string        `yaml:"server_url"`
	Insecure  bool          `yaml:"insecure"`
	Username  string        `yaml:"username"`
	Password  string        `yaml:"password"`
	Timeout   time.Duration `yaml:"timeout"`
}
