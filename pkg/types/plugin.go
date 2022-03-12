package types

// Plugin details
type PluginFile struct {
	Provider map[string]ProviderData `yaml:"provider"`
}

// Provider config details
type ProviderData struct {
	PluginName string                 `yaml:"plugin"`
	Config     map[string]interface{} `yaml:"config"`
}
