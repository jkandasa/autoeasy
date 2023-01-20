package operator

import (
	"strings"

	openshiftInstallCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/install"
	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	nsAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/namespace"
	operatorAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/operator"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	installOperatorCatalogSource string
	installOperatorNamespace     string
	installOperatorChannel       string
	installOperatorEnvironments  []string
	installOperatorForceRecreate bool
)

func init() {
	openshiftInstallCmd.AddCommand(installOperatorCmd)
	installOperatorCmd.Flags().StringVar(&installOperatorCatalogSource, "catalog-source", "redhat-operators", "catalog source name")
	installOperatorCmd.Flags().StringVar(&installOperatorNamespace, "namespace", "openshift-operators", "namespace of the operator")
	installOperatorCmd.Flags().StringVar(&installOperatorChannel, "channel", "stable", "channel name of the operator")
	installOperatorCmd.Flags().StringSliceVar(&installOperatorEnvironments, "environment", []string{}, "comma separated environment variables. key1=value1,key2=value2")
	installOperatorCmd.Flags().BoolVar(&installOperatorForceRecreate, "force", false, "uninstall the operator if exists and install")
}

var installOperatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "installs operator",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, operatorsList []string) {
		// get kubernetes client
		k8sClient := openshiftClient.GetKubernetesClient()

		// uninstall operators if recreate enabled
		if installOperatorForceRecreate {
			err := uninstallOperator(k8sClient, operatorsList)
			if err != nil {
				zap.L().Error("error on uninstalling operators", zap.Any("operators", operatorsList), zap.Error(err))
				rootCmd.ExitWithError()
			}
		}

		// create namespace if not available
		err := nsAPI.CreateIfNotAvailable(k8sClient, installOperatorNamespace)
		if err != nil {
			zap.L().Fatal("error on creating namespace", zap.String("namespace", installOperatorNamespace), zap.Error(err))
			rootCmd.ExitWithError()
		}

		// parse environments
		_environments := []corev1.EnvVar{}
		for _, _rawEnv := range installOperatorEnvironments {
			_envs := strings.SplitN(_rawEnv, "=", 2)
			if len(_envs) != 2 {
				zap.L().Error("invalid environment variable", zap.String("env", _rawEnv))
				continue
			}
			_environments = append(_environments, corev1.EnvVar{
				Name:  strings.ToUpper(_envs[0]),
				Value: _envs[1],
			})
		}

		// install operators
		for _, _operator := range operatorsList {
			subscription := corsosv1alpha1.Subscription{
				ObjectMeta: metav1.ObjectMeta{
					Name:      _operator,
					Namespace: installOperatorNamespace,
				},
				Spec: &corsosv1alpha1.SubscriptionSpec{
					CatalogSource:          installOperatorCatalogSource,
					CatalogSourceNamespace: "openshift-marketplace",
					Package:                _operator,
					Channel:                installOperatorChannel,
					Config: &corsosv1alpha1.SubscriptionConfig{
						Env: _environments,
					},
				},
			}

			tc := types.TimeoutConfig{}
			tc.UpdateDefaults()
			tc.ExpectedSuccessCount = 2
			zap.L().Debug("installing an operator", zap.String("operator", _operator))
			err = operatorAPI.Install(k8sClient, &subscription, tc)
			if err != nil {
				zap.L().Error("error on installing an operator", zap.String("name", _operator), zap.Error(err))
				rootCmd.ExitWithError()
			}
			zap.L().Info("installed an operator", zap.String("operator", subscription.GetName()), zap.String("namespace", subscription.GetNamespace()))
		}
	},
}
