package local_command

import (
	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	localCmdTY "github.com/jkandasa/autoeasy/plugin/provider/local_command/types"
	providerPluginTY "github.com/jkandasa/autoeasy/plugin/provider/types"
)

const (
	PluginName = "local_command"
)

type LocalCommand struct {
	Config localCmdTY.ProviderConfig
}

func New(config map[string]interface{}) (providerPluginTY.Plugin, error) {
	cfg := localCmdTY.ProviderConfig{}
	err := formatterUtils.YamlInterfaceToStruct(config, &cfg)
	if err != nil {
		return nil, err
	}
	cfg.UpdateDefaults()
	return &LocalCommand{Config: cfg}, nil
}

func (lc *LocalCommand) Name() string {
	return PluginName
}

func (lc *LocalCommand) Start() error {
	return nil
}

func (lc *LocalCommand) Close() error {
	return nil
}

func (lc *LocalCommand) Execute(task *templateTY.Task) error {
	return lc.run(task)
}
