package api

import (
	"context"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/jkandasa/autoeasy/pkg/utils"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/store"
	mcUtils "github.com/mycontroller-org/server/v2/pkg/utils"

	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(opts []client.ListOption) (*appsv1.DeploymentList, error) {
	deploymentList := &appsv1.DeploymentList{}
	err := store.K8SClient.List(context.Background(), deploymentList, opts...)
	if err != nil {
		return nil, err
	}
	return deploymentList, nil
}

func Get(name, namespace string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := store.K8SClient.Get(context.Background(), namespacedName, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func Delete(deployment *appsv1.Deployment) error {
	return utils.IgnoreNotFoundError(store.K8SClient.Delete(context.Background(), deployment))
}

func DeleteOfAll(deployment *appsv1.Deployment, opts []client.DeleteAllOfOption) error {
	if deployment == nil {
		deployment = &appsv1.Deployment{}
	}
	return store.K8SClient.DeleteAllOf(context.Background(), deployment, opts...)
}

func Create(cfg map[string]interface{}) error {
	deployment := &appsv1.Deployment{}
	err := mcUtils.MapToStruct(mcUtils.TagNameJSON, cfg, deployment)
	if err != nil {
		return err
	}
	return store.K8SClient.Create(context.Background(), deployment)
}

// wait for deployment
func WaitForDeployments(deployments []string, namespace string, tc openshiftTY.TimeoutConfig) error {
	executeFunc := func() (bool, error) {
		return isDeployed(deployments, namespace)
	}
	return funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, tc.Timeout, tc.ScanInterval, tc.ExpectedSuccessCount)
}

func isDeployed(deployments []string, namespace string) (bool, error) {
	// check deployment status
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}
	deploymentList, err := List(opts)
	if err != nil {
		return false, err
	}
	notready := []string{}
	for _, name := range deployments {
		found := false
		for _, dep := range deploymentList.Items {
			if dep.Namespace == namespace && dep.Name == name {
				found = true
				if dep.Status.Replicas != dep.Status.ReadyReplicas {
					notready = append(notready, dep.Name)
				}
			}
		}
		if !found {
			notready = append(notready, name)
		}
	}

	if len(notready) == 0 { // all deployments success
		zap.L().Debug("deployments are running", zap.String("namespace", namespace), zap.Any("deployments", deployments))
		return true, nil
	}
	zap.L().Debug("waiting for deployment", zap.String("namespace", namespace), zap.Any("deployments", notready))
	return false, nil
}
