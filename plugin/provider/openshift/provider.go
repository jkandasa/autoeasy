package openshift

// update defaults
// task.Config.TimeoutConfig.UpdateDefaults()

import (
	"errors"
	"fmt"

	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	clusterAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/cluster"
	k8s "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	taskCS "github.com/jkandasa/autoeasy/plugin/provider/openshift/task/catalog_source"
	taskDeployment "github.com/jkandasa/autoeasy/plugin/provider/openshift/task/deployment"
	taskICSP "github.com/jkandasa/autoeasy/plugin/provider/openshift/task/image_content_source_policy"
	taskNS "github.com/jkandasa/autoeasy/plugin/provider/openshift/task/namespace"
	taskRoute "github.com/jkandasa/autoeasy/plugin/provider/openshift/task/route"
	taskSubscription "github.com/jkandasa/autoeasy/plugin/provider/openshift/task/subscription"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	providerPluginTY "github.com/jkandasa/autoeasy/plugin/provider/types"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PluginName = "openshift"
)

type Openshift struct {
	Config       openshiftTY.PluginConfig
	Client       *k8s.K8SClient
	K8SClient    client.Client
	K8SClientSet *kubernetes.Clientset
}

func New(config map[string]interface{}) (providerPluginTY.Plugin, error) {
	openshiftCfg := openshiftTY.PluginConfig{}
	err := formatterUtils.YamlInterfaceToStruct(config, &openshiftCfg)
	if err != nil {
		return nil, err
	}
	return &Openshift{Config: openshiftCfg, Client: k8s.New(&openshiftCfg)}, nil
}

func (o *Openshift) Name() string {
	return PluginName
}

func (o *Openshift) Start() error {
	cfg := &o.Config
	if !cfg.LoadClient {
		zap.L().Info("loading openshift client is disabled. you have to load it via task")
		return nil
	}

	return o.login(cfg)
}

func (o *Openshift) Close() error {
	return nil
}

func (o *Openshift) Execute(task *templateTY.Task) (interface{}, error) {
	config := &openshiftTY.ProviderConfig{}
	err := formatterUtils.YamlInterfaceToStruct(task.Input, config)
	if err != nil {
		return nil, err
	}

	switch config.Kind {
	case openshiftTY.KindCatalogSource:
		return taskCS.Run(o.K8SClient, config)

	case openshiftTY.KindImageContentSourcePolicy:
		return taskICSP.Run(o.K8SClient, config)

	case openshiftTY.KindNamespace:
		return taskNS.Run(o.K8SClient, config)

	case openshiftTY.KindSubscription:
		return taskSubscription.Run(o.K8SClient, config)

	case openshiftTY.KindDeployment:
		return taskDeployment.Run(o.K8SClient, config)

	case openshiftTY.KindRoute:
		return taskRoute.Run(o.K8SClient, config)

	case openshiftTY.KindInternal:
		return o.runInternal(config)

	default:
		return nil, fmt.Errorf("invalid kind:[%s]", config.Kind)

	}
}

func (o *Openshift) runInternal(cfg *openshiftTY.ProviderConfig) (interface{}, error) {
	switch cfg.Function {
	case openshiftTY.FuncLogin:
		if len(cfg.Data) == 0 {
			return nil, fmt.Errorf("no data supplied. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
		}
		osCfg := &openshiftTY.PluginConfig{}
		err := formatterUtils.YamlInterfaceToStruct(cfg.Data[0], osCfg)
		if err != nil {
			return nil, err
		}
		return nil, o.login(osCfg)

	case openshiftTY.FuncLogout:
		return nil, o.logout()

	case openshiftTY.FuncPrintInfo:
		return nil, errors.New("not implemented yet")
	}

	return nil, fmt.Errorf("unknown function. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
}

func (o *Openshift) login(cfg *openshiftTY.PluginConfig) error {
	if !cfg.LoadFromConfig {
		err := cfg.Validate()
		if err != nil {
			return err
		}
	}

	// login
	err := o.Client.Login(cfg, true)
	if err != nil {
		zap.L().Error("error on doing login", zap.Error(err))
		return err
	}

	// load client
	kubeClient, err := o.Client.NewClient()
	if err != nil {
		zap.L().Error("error on loading kubernetes client", zap.Error(err))
		return err
	}
	o.K8SClient = kubeClient

	// load client set
	kubeClientSet, err := o.Client.NewClientset()
	if err != nil {
		zap.L().Error("error on loading kubernetes client set", zap.Error(err))
		return err
	}
	o.K8SClientSet = kubeClientSet

	zap.L().Info("kubernetes client loaded successfully")
	clusterAPI.PrintClusterInfo(o.K8SClient, o.K8SClientSet)
	return nil
}

func (o *Openshift) logout() error {
	o.K8SClient = nil
	o.K8SClientSet = nil
	return nil
}
