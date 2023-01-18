package jaeger

import (
	"strings"

	jaegerv1 "github.com/jaegertracing/jaeger-operator/apis/v1"
	openshiftInstallCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/install"
	jaegerAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/jaeger"
	nsAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/namespace"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	jaegerDeploymentStrategy    string
	jaegerNamespace             string
	jaegerUseEsStorage          bool
	jaegerEsNodeCount           uint32
	jaegerEsResourceLimitCPU    string
	jaegerEsResourceLimitMemory string
	jaegerEsOptions             []string
	jaegerForceRecreate         bool
)

func init() {
	openshiftInstallCmd.AddCommand(installJaegerCmd)
	installJaegerCmd.Flags().StringVar(&jaegerDeploymentStrategy, "strategy", string(jaegerv1.DeploymentStrategyAllInOne), "deployment strategy options: allinone, production, streaming")
	installJaegerCmd.Flags().StringVar(&jaegerNamespace, "namespace", "jaeger-ns", "namespace of the jaeger")
	installJaegerCmd.Flags().BoolVar(&jaegerUseEsStorage, "use-es-storage", false, "use elasticsearch storage")
	installJaegerCmd.Flags().Uint32Var(&jaegerEsNodeCount, "es-node-count", 1, "elasticsearch node count")
	installJaegerCmd.Flags().StringVar(&jaegerEsResourceLimitCPU, "es-cpu-limit", "2", "elasticsearch node container memory limit")
	installJaegerCmd.Flags().StringVar(&jaegerEsResourceLimitMemory, "es-memory-limit", "2Gi", "elasticsearch node container memory limit")
	installJaegerCmd.Flags().StringSliceVar(&jaegerEsOptions, "es-option", []string{}, "comma separated elasticsearch options. key1=value1,key2=value2")
	installJaegerCmd.Flags().BoolVar(&jaegerForceRecreate, "force", false, "uninstall the jaeger if exists and install")
}

var installJaegerCmd = &cobra.Command{
	Use:   "jaeger",
	Short: "installs jaeger",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, jaegersName []string) {
		// get kubernetes client
		k8sClient := openshiftClient.GetKubernetesClient()

		// create jaeger CR
		jaegerCR := &jaegerv1.Jaeger{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jaegersName[0],
				Namespace: jaegerNamespace,
			},
			Spec: jaegerv1.JaegerSpec{
				Strategy: jaegerv1.DeploymentStrategy(jaegerDeploymentStrategy),
			},
		}

		// update es storage details
		if jaegerUseEsStorage {
			// es storage options
			esOptions := make(map[string]interface{})
			for _, esRawOption := range jaegerEsOptions {
				esOption := strings.SplitN(esRawOption, "=", 2)
				if len(esOption) != 2 {
					zap.L().Error("invalid es option", zap.String("esOption", esRawOption))
					continue
				}
				esOptions[esOption[0]] = esOption[1]
			}

			// update elasticsearch storage details
			jaegerCR.Spec.Storage = jaegerv1.JaegerStorageSpec{
				Type:    jaegerv1.JaegerESStorage,
				Options: jaegerv1.NewOptions(esOptions),
				Elasticsearch: jaegerv1.ElasticsearchSpec{
					NodeCount: int32(jaegerEsNodeCount),
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse(jaegerEsResourceLimitMemory),
							corev1.ResourceCPU:    resource.MustParse(jaegerEsResourceLimitCPU),
						},
					},
				},
			}
		}

		// uninstall jaeger if recreate enabled
		if jaegerForceRecreate {
			err := jaegerAPI.Delete(k8sClient, jaegerCR)
			if err != nil {
				zap.L().Error("error on uninstalling jaeger", zap.Any("jaeger", jaegerCR.GetName()), zap.Error(err))
				return
			}
		}

		// create namespace if not available
		err := nsAPI.CreateIfNotAvailable(k8sClient, jaegerCR.GetNamespace())
		if err != nil {
			zap.L().Fatal("error on creating namespace", zap.String("namespace", jaegerCR.GetNamespace()), zap.Error(err))
			return
		}

		// install jaeger
		tc := types.TimeoutConfig{}
		tc.UpdateDefaults()
		tc.ExpectedSuccessCount = 2
		zap.L().Debug("installing an jaeger", zap.String("name", jaegerCR.GetName()), zap.String("namespace", jaegerCR.GetNamespace()))
		err = jaegerAPI.CreateAndWait(k8sClient, jaegerCR, tc)
		if err != nil {
			zap.L().Error("error on installing jaeger", zap.String("name", jaegerCR.GetName()), zap.String("namespace", jaegerCR.GetNamespace()), zap.Error(err))
			return
		}
		zap.L().Info("installed an jaeger", zap.String("name", jaegerCR.GetName()), zap.String("namespace", jaegerCR.GetNamespace()))
	},
}
