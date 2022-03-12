package jenkins_provider

import (
	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	jenkinsProviderTY "github.com/jkandasa/autoeasy/plugin/provider/jenkins/types"
	providerPluginTY "github.com/jkandasa/autoeasy/plugin/provider/types"
	jenkins "github.com/jkandasa/jenkinsctl/pkg/jenkins"
	jenkinsTY "github.com/jkandasa/jenkinsctl/pkg/types"
	jenkinsCfgTY "github.com/jkandasa/jenkinsctl/pkg/types/config"
	"go.uber.org/zap"
)

const (
	PluginName = "jenkins"
)

type Jenkins struct {
	Config jenkinsProviderTY.PluginConfig
	Client *jenkins.Client
}

func New(config map[string]interface{}) (providerPluginTY.Plugin, error) {
	cfg := jenkinsProviderTY.PluginConfig{}
	err := formatterUtils.YamlInterfaceToStruct(config, &cfg)
	if err != nil {
		return nil, err
	}

	return &Jenkins{Config: cfg}, nil
}

func (j *Jenkins) Name() string {
	return PluginName
}

// Start loads jenkins client
func (j *Jenkins) Start() error {
	cfg := &jenkinsCfgTY.Config{
		URL:      j.Config.ServerURL,
		Insecure: j.Config.Insecure,
		Username: j.Config.Username,
		Password: j.Config.Password,
	}
	cfg.EncodePassword() // encode the password
	client, err := jenkins.NewClient(cfg, &jenkinsTY.IOStreams{})
	if err != nil {
		return err
	}
	j.Client = client
	zap.L().Debug("jenkins server", zap.Any("version", client.Version()))
	return nil
}

func (j *Jenkins) Close() error {
	return nil
}

func (j *Jenkins) Execute(task *templateTY.Task) (interface{}, error) {
	config := &jenkinsProviderTY.ProviderConfig{}
	err := formatterUtils.YamlInterfaceToStruct(task.Input, config)
	if err != nil {
		return nil, err
	}
	return j.run(config)
}
