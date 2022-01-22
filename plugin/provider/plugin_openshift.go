package provider

import (
	openshiftPlugin "github.com/jkandasa/autoeasy/plugin/provider/openshift"
)

func init() {
	Register(openshiftPlugin.PluginName, openshiftPlugin.New)
}
