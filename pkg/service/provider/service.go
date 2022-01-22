package provider

import (
	"fmt"

	providerPlugin "github.com/jkandasa/autoeasy/plugin/provider"
	providerPluginTY "github.com/jkandasa/autoeasy/plugin/provider/types"
)

var store = make(map[string]providerPluginTY.Plugin)

func Start(cfg map[string]interface{}) error {
	for _, providerName := range providerPlugin.Plugins() {
		providerCfg := make(map[string]interface{})
		rawCfg, found := cfg[providerName]
		if found {
			provCfg, ok := rawCfg.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid configuration. format should be in map[sting]interface format. provderName:%s", providerName)
			}
			providerCfg = provCfg
		}

		provider, err := providerPlugin.Create(providerName, providerCfg)
		if err != nil {
			return err
		}
		err = provider.Start()
		if err != nil {
			return err
		}
		store[providerName] = provider
	}
	return nil
}

func GetProvider(name string) providerPluginTY.Plugin {
	if provider, found := store[name]; found {
		return provider
	}
	return nil
}
