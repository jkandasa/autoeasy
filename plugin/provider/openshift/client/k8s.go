package k8s

import (
	"flag"
	"path/filepath"
	"sync"

	jaegerv1 "github.com/jaegertracing/jaeger-operator/apis/v1"
	osoperatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	parseConfig sync.Once
	kubeconfig  *string
)

type K8SClient struct {
	restConfig *rest.Config
	Config     *openshiftTY.PluginConfig
}

func New(cfg *openshiftTY.PluginConfig) *K8SClient {
	return &K8SClient{Config: cfg}
}

func (k8s *K8SClient) loadConfigFromFile() error {
	parseConfig.Do(func() { // run only once
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
	})

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
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
			return err
		}
	} else {
		// load config from input
		k8s.restConfig = &rest.Config{
			Host:     cfg.Server,
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	// update insecure
	k8s.restConfig.TLSClientConfig.Insecure = cfg.Insecure

	return nil

}

func (k8s *K8SClient) NewClientset() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(k8s.restConfig)
}

func (k8s *K8SClient) NewClient() (client.Client, error) {
	client, err := client.New(k8s.restConfig, client.Options{})
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
