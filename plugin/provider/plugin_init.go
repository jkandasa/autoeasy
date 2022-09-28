package provider

import (
	jenkinsPlugin "github.com/jkandasa/autoeasy/plugin/provider/jenkins"
	localCmdPlugin "github.com/jkandasa/autoeasy/plugin/provider/local_command"
	openshiftPlugin "github.com/jkandasa/autoeasy/plugin/provider/openshift"
)

func init() {
	Register(jenkinsPlugin.PluginName, jenkinsPlugin.New)
	Register(localCmdPlugin.PluginName, localCmdPlugin.New)
	Register(openshiftPlugin.PluginName, openshiftPlugin.New)
}
