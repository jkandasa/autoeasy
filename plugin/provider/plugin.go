package provider

import (
	"fmt"

	providerTY "github.com/jkandasa/autoeasy/plugin/provider/types"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(config map[string]interface{}) (providerTY.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Plugins() []string {
	plugins := []string{}
	for pluginName := range creators {
		plugins = append(plugins, pluginName)
	}
	return plugins
}

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(name string, config map[string]interface{}) (p providerTY.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("provider plugin [%s] is not registered", name)
	}
	return
}
