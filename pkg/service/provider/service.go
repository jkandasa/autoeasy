package provider

import (
	"fmt"

	"github.com/jkandasa/autoeasy/pkg/types"
	providerPlugin "github.com/jkandasa/autoeasy/plugin/provider"
	providerPluginTY "github.com/jkandasa/autoeasy/plugin/provider/types"
	"go.uber.org/zap"
)

var store = make(map[string]providerPluginTY.Plugin)

func Start(cfg map[string]types.ProviderData) error {
	// load given providers
	for providerName, providerData := range cfg {
		if providerData.PluginName == "" {
			return fmt.Errorf("plugin name can not be empty. providerName:%s", providerName)
		}
		zap.L().Debug("creating plugin", zap.String("name", providerName), zap.String("plugin", providerData.PluginName))
		provider, err := providerPlugin.Create(providerData.PluginName, providerData.Config)
		if err != nil {
			return err
		}
		zap.L().Debug("starting plugin", zap.String("name", providerName), zap.String("plugin", providerData.PluginName))
		err = provider.Start()
		if err != nil {
			zap.L().Error("error on starting a provider", zap.String("providerName", providerName), zap.Error(err))
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
