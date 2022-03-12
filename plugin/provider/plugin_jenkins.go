package provider

import (
	jenkinsPlugin "github.com/jkandasa/autoeasy/plugin/provider/jenkins"
)

func init() {
	Register(jenkinsPlugin.PluginName, jenkinsPlugin.New)
}
