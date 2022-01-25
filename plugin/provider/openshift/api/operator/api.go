package api

import (
	"github.com/jkandasa/autoeasy/pkg/utils"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	csvAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/cluster_service_version"
	deploymentAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/deployment"
	subscriptionAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/subscription"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	mcUtils "github.com/mycontroller-org/server/v2/pkg/utils"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Uninstall removes the Subscription and ClusterServiceVersion
func Uninstall(cfg map[string]interface{}) error {
	subscription := &corsosv1alpha1.Subscription{}
	err := mcUtils.MapToStruct(mcUtils.TagNameJSON, cfg, subscription)
	if err != nil {
		return err
	}
	opts := []client.ListOption{
		client.InNamespace(""),
	}

	subscriptionList, err := subscriptionAPI.List(opts)
	if err != nil {
		return err
	}

	// get the match and remove subscription and csv
	for _, rxSub := range subscriptionList.Items {
		if subscription.Name == rxSub.Name {
			installedCSV := rxSub.Status.InstalledCSV
			removableCSVs := []string{installedCSV}
			if installedCSV != rxSub.Status.CurrentCSV {
				removableCSVs = append(removableCSVs, rxSub.Status.CurrentCSV)
			}

			// remove csv
			csvList, err := csvAPI.List(opts)
			if err != nil {
				return err
			}
			for _, csv := range csvList.Items {
				if _, remove := mcUtils.FindItem(removableCSVs, csv.Name); remove {
					err = csvAPI.Delete(&csv)
					if err != nil {
						zap.L().Error("error on csv deletion", zap.String("name", csv.Name), zap.String("namespace", csv.Namespace), zap.Error(err))
						return err
					}
				}
			}

			// remove subscription
			err = subscriptionAPI.Delete(&rxSub)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Install(cfg map[string]interface{}, tc openshiftTY.TimeoutConfig) error {
	// create subscription
	err := subscriptionAPI.CreateWithMap(cfg)
	if err != nil {
		return err
	}

	// get updated subscription
	objectMeta, err := utils.GetObjectMeta(cfg)
	if err != nil {
		return err
	}
	subscription, err := subscriptionAPI.Get(objectMeta.Name, objectMeta.Namespace)
	if err != nil {
		return err
	}

	// get deployments
	deployments := []string{}
	executeFunc := func() (bool, error) {
		_deployments, err := getDeployments(subscription.Name, subscription.Namespace)
		if err != nil {
			return false, err
		}

		deployments = _deployments
		return len(deployments) > 0, nil
	}
	err = funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, tc.Timeout, tc.ScanInterval, tc.ExpectedSuccessCount)
	if err != nil {
		return err
	}

	return deploymentAPI.WaitForDeployments(deployments, subscription.Namespace, tc)
}

func getDeployments(subscriptionName, namespace string) ([]string, error) {
	deployments := []string{}
	zap.L().Debug("operator deployment details not available. getting deployments details", zap.String("subscriptionName", subscriptionName), zap.String("namespace", namespace))
	subscription, err := subscriptionAPI.Get(subscriptionName, namespace)
	if err != nil {
		return nil, err
	}
	installedCSV := subscription.Status.InstalledCSV
	if installedCSV != "" { // get csv details
		opts := []client.ListOption{
			client.InNamespace(""),
		}
		csvList, err := csvAPI.List(opts)
		if err != nil {
			return nil, err
		}
		for _, csv := range csvList.Items {
			if csv.Name == installedCSV {
				for _, deployment := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
					if _, available := mcUtils.FindItem(deployments, deployment.Name); !available {
						deployments = append(deployments, deployment.Name)
					}
				}
			}
		}
		if len(deployments) > 0 {
			zap.L().Debug("operator deployments detail", zap.String("subscriptionName", subscriptionName), zap.String("namespace", namespace), zap.Any("deployments", deployments))
		}
	}
	return deployments, nil
}
