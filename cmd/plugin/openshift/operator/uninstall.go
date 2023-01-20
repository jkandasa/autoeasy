package operator

import (
	openshiftUninstallCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/uninstall"
	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	operatorAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/operator"
	subscriptionAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/subscription"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	openshiftUninstallCmd.AddCommand(uninstallOperatorCmd)
}

var uninstallOperatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "uninstalls operator",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, operatorsList []string) {
		// get kubernetes client
		k8sClient := openshiftClient.GetKubernetesClient()

		err := uninstallOperator(k8sClient, operatorsList)
		if err != nil {
			zap.L().Error("error on uninstalling operator", zap.Any("operators", operatorsList), zap.Error(err))
			rootCmd.ExitWithError()
		}
	},
}

func uninstallOperator(k8sClient client.Client, operatorsList []string) error {
	subscriptionsList, err := subscriptionAPI.List(k8sClient, []client.ListOption{})
	if err != nil {
		zap.L().Error("error on getting subscriptions list", zap.Error(err))
		return err
	}

	// uninstall operators
	for _, _operator := range operatorsList {
		found := false
		for _, _subscription := range subscriptionsList.Items {
			if _subscription.Name == _operator {
				found = true
				zap.L().Debug("uninstalling an operator", zap.String("operator", _operator))
				err := operatorAPI.Uninstall(k8sClient, &_subscription)
				if err != nil {
					zap.L().Error("error on uninstalling an operator", zap.String("operator", _operator), zap.Error(err))
					continue
				}
				zap.L().Info("uninstalled an operator", zap.String("operator", _subscription.GetName()), zap.String("namespace", _subscription.GetNamespace()))
			}
		}
		if !found {
			zap.L().Info("operator not installed", zap.String("operator", _operator))
		}
	}
	return nil
}
