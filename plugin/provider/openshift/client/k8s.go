package k8s

import (
	"flag"
	"path/filepath"

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

var restConfig *rest.Config

func GetConfig() (*rest.Config, error) {
	if restConfig == nil {
		cfg, err := getConfig()
		if err != nil {
			return nil, err
		}
		restConfig = cfg
	}
	return restConfig, nil
}

func getConfig() (*rest.Config, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	restConfig = config

	return config, nil
}

func NewClient(cfg *openshiftTY.PluginConfig) (client.Client, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}

	config.TLSClientConfig.Insecure = cfg.Insecure

	client, err := client.New(config, client.Options{})
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

func NewClientset(cfg *openshiftTY.PluginConfig) (*kubernetes.Clientset, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}

	config.TLSClientConfig.Insecure = cfg.Insecure

	return kubernetes.NewForConfig(config)
}
