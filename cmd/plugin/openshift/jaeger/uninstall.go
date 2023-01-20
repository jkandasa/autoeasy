package jaeger

import (
	jaegerv1 "github.com/jaegertracing/jaeger-operator/apis/v1"
	openshiftUninstallCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/uninstall"
	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	jaegerAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/jaeger"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	openshiftUninstallCmd.AddCommand(uninstallJaegerCmd)
	uninstallJaegerCmd.Flags().StringVar(&jaegerNamespace, "namespace", "jaeger-ns", "namespace of the jaeger")
}

var uninstallJaegerCmd = &cobra.Command{
	Use:   "jaeger",
	Short: "uninstalls jaeger",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, jaegersList []string) {
		// get kubernetes client
		k8sClient := openshiftClient.GetKubernetesClient()

		// uninstall jaegers
		for _, jaegerName := range jaegersList {
			// create jaeger CR
			jaegerCR := &jaegerv1.Jaeger{
				ObjectMeta: metav1.ObjectMeta{
					Name:      jaegerName,
					Namespace: jaegerNamespace,
				},
			}
			zap.L().Debug("uninstalling jaeger", zap.String("name", jaegerCR.GetName()), zap.String("namespace", jaegerCR.GetNamespace()))
			err := jaegerAPI.Delete(k8sClient, jaegerCR)
			if err != nil {
				zap.L().Error("error on uninstalling jaeger", zap.String("name", jaegerCR.GetName()), zap.String("namespace", jaegerCR.GetNamespace()), zap.Error(err))
				rootCmd.ExitWithError()
			}
			zap.L().Info("uninstalled jaeger", zap.String("name", jaegerCR.GetName()), zap.String("namespace", jaegerCR.GetNamespace()))
		}
	},
}
