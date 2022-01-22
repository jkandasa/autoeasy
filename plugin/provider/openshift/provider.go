package openshift

// update defaults
// action.Config.TimeoutConfig.UpdateDefaults()

import (
	"fmt"

	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	axnCS "github.com/jkandasa/autoeasy/plugin/provider/openshift/action/catalog_source"
	axnDeployment "github.com/jkandasa/autoeasy/plugin/provider/openshift/action/deployment"
	axnICSP "github.com/jkandasa/autoeasy/plugin/provider/openshift/action/image_content_source_policy"
	axnNS "github.com/jkandasa/autoeasy/plugin/provider/openshift/action/namespace"
	axnSubscription "github.com/jkandasa/autoeasy/plugin/provider/openshift/action/subscription"
	clusterAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/cluster"
	k8s "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	OpenshiftStore "github.com/jkandasa/autoeasy/plugin/provider/openshift/store"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	providerPluginTY "github.com/jkandasa/autoeasy/plugin/provider/types"
	"go.uber.org/zap"
)

const (
	PluginName = "openshift"
)

type Openshift struct {
	Config openshiftTY.PluginConfig
}

func New(config map[string]interface{}) (providerPluginTY.Plugin, error) {
	openshiftCfg := openshiftTY.PluginConfig{}
	err := formatterUtils.YamlInterfaceToStruct(config, &openshiftCfg)
	if err != nil {
		return nil, err
	}
	return &Openshift{Config: openshiftCfg}, nil
}

func (o *Openshift) Name() string {
	return PluginName
}

func (o *Openshift) Start() error {
	cfg := &o.Config
	if !cfg.LoadClient {
		zap.L().Info("loading openshift client is disabled. you have to load it via action")
		return nil
	}

	if !cfg.LoadFromConfig {
		err := cfg.Validate()
		if err != nil {
			return err
		}
	}

	// load client
	kubeClient, err := k8s.NewClient(cfg)
	if err != nil {
		zap.L().Fatal("error on loading kubernetes client", zap.Error(err))
	}
	OpenshiftStore.K8SClient = kubeClient

	// load client set
	kubeClientSet, err := k8s.NewClientset(cfg)
	if err != nil {
		zap.L().Fatal("error on loading kubernetes client set", zap.Error(err))
	}
	OpenshiftStore.K8SClientSet = kubeClientSet

	zap.L().Info("kubernetes client loaded successfully")
	clusterAPI.PrintClusterInfo()
	return nil
}

func (o *Openshift) Close() error {
	return nil
}

func (o *Openshift) Execute(action *templateTY.Action) error {
	config := &openshiftTY.ProviderConfig{}
	err := formatterUtils.YamlInterfaceToStruct(action.Input, config)
	if err != nil {
		return err
	}

	switch config.Kind {
	case openshiftTY.KindCatalogSource:
		return axnCS.Run(config)

	case openshiftTY.KindImageContentSourcePolicy:
		return axnICSP.Run(config)

	case openshiftTY.KindNamespace:
		return axnNS.Run(config)

	case openshiftTY.KindSubscription:
		return axnSubscription.Run(config)

	case openshiftTY.KindDeployment:
		return axnDeployment.Run(config)

	default:
		return fmt.Errorf("invalid kind:[%s]", config.Kind)

	}
}
