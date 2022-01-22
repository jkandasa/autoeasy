package provider

import (
	localCmdPlugin "github.com/jkandasa/autoeasy/plugin/provider/local_command"
)

func init() {
	Register(localCmdPlugin.PluginName, localCmdPlugin.New)
}
