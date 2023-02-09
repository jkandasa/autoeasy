package k8s

import (
	"flag"
	"os"
	"path/filepath"
	"sync"

	jaegerv1 "github.com/jaegertracing/jaeger-operator/apis/v1"
	osoperatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"go.uber.org/zap"

	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"

	"k8s.io/client-go/kubernetes"
	restClient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	parseConfig        sync.Once
	kubeConfigFilename *string
)

type K8SClient struct {
	restConfig *restClient.Config
	Config     *openshiftTY.PluginConfig
}

func New(cfg *openshiftTY.PluginConfig) *K8SClient {
	return &K8SClient{Config: cfg}
}

func (k8s *K8SClient) loadConfigFromFile() error {
	parseConfig.Do(func() { // run only once
		if os.Getenv("KUBECONFIG") != "" {
			_kubeConfigFilename := os.Getenv("KUBECONFIG")
			kubeConfigFilename = &_kubeConfigFilename
		} else if home := homedir.HomeDir(); home != "" {
			kubeConfigFilename = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeConfigFilename = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
	})

	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfigFilename)
	if err != nil {
		return err
	}

	k8s.restConfig = config
	return nil
}

func (k8s *K8SClient) Login(cfg *openshiftTY.PluginConfig, forceRelogin bool) error {
	if !forceRelogin && k8s.restConfig != nil {
		return nil
	}

	// load config from file
	if cfg.LoadFromConfig {
		err := k8s.loadConfigFromFile()
		if err != nil {
			zap.L().Debug("error on getting kube config, trying with in-cluster configuration", zap.Error(err))
			// if no config files found or error with config file login, try with in-cluster configurations
			_restConfig, err := restClient.InClusterConfig()
			if err != nil {
				return err
			}
			k8s.restConfig = _restConfig
		}
	} else {
		// load config from input
		k8s.restConfig = &restClient.Config{
			Host:     cfg.Server,
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	// update insecure
	k8s.restConfig.TLSClientConfig.Insecure = cfg.Insecure

	return nil

}

func (k8s *K8SClient) GetRestConfig() *restClient.Config {
	return k8s.restConfig
}

func (k8s *K8SClient) NewClientset() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(k8s.restConfig)
}

func NewClientFromRestConfig(restConfig *restClient.Config) (client.Client, error) {
	client, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, err
	}

	// update schema
	err = registerSchema(client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (k8s *K8SClient) NewClient() (client.Client, error) {
	return NewClientFromRestConfig(k8s.restConfig)
}

func registerSchema(client client.Client) error {
	err := jaegerv1.AddToScheme(client.Scheme())
	if err != nil {
		return err
	}

	err = corsosv1alpha1.AddToScheme(client.Scheme())
	if err != nil {
		return err
	}

	err = osoperatorv1alpha1.AddToScheme(client.Scheme())
	if err != nil {
		return err
	}

	err = routev1.AddToScheme(client.Scheme())
	if err != nil {
		return err
	}

	return nil
}

func GetKubernetesClient() client.Client {
	k8sClientConf := GetK8SClientConfig()
	k8sClient, err := k8sClientConf.NewClient()
	if err != nil {
		zap.L().Fatal("error on loading k8s client", zap.Error(err))
	}

	return k8sClient
}

func GetK8SClientConfig() *K8SClient {
	zap.L().Debug("loading k8s client")
	k8sClientConf := &K8SClient{Config: &openshiftTY.PluginConfig{LoadClient: true, LoadFromConfig: true, Insecure: true}}
	err := k8sClientConf.Login(k8sClientConf.Config, false)
	if err != nil {
		zap.L().Fatal("error on login into k8s cluster", zap.Error(err))
	}
	return k8sClientConf
}
