package api

import (
	"context"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	podAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/pod"

	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*appsv1.DeploymentList, error) {
	deploymentList := &appsv1.DeploymentList{}
	err := k8sClient.List(context.Background(), deploymentList, opts...)
	if err != nil {
		return nil, err
	}
	return deploymentList, nil
}

func Get(k8sClient client.Client, name, namespace string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(context.Background(), namespacedName, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func ListPods(k8sClient client.Client, deploymentName, namespace string) (*corev1.PodList, error) {
	deployment, err := Get(k8sClient, deploymentName, namespace)
	if err != nil {
		return nil, err
	}
	return podAPI.List(k8sClient, []client.ListOption{client.MatchingLabels(deployment.Spec.Selector.MatchLabels)})
}

func ListRunningPods(k8sClient client.Client, deploymentName, namespace string) ([]corev1.Pod, error) {
	podsList, err := ListPods(k8sClient, deploymentName, namespace)
	if err != nil {
		return nil, err
	}
	pods := make([]corev1.Pod, 0)
	for _, pod := range podsList.Items {
		if pod.Status.Phase == corev1.PodRunning {
			pods = append(pods, pod)
		}
	}
	return pods, nil
}

func Delete(k8sClient client.Client, deployment *appsv1.Deployment) error {
	return utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), deployment))
}

func DeleteOfAll(k8sClient client.Client, deployment *appsv1.Deployment, opts []client.DeleteAllOfOption) error {
	if deployment == nil {
		deployment = &appsv1.Deployment{}
	}
	return k8sClient.DeleteAllOf(context.Background(), deployment, opts...)
}

func Create(k8sClient client.Client, deployment *appsv1.Deployment) error {
	return k8sClient.Create(context.Background(), deployment)
}

func CreateWithMap(k8sClient client.Client, cfg map[string]interface{}) error {
	deployment := &appsv1.Deployment{}
	err := formatterUtils.JsonMapToStruct(cfg, deployment)
	if err != nil {
		return err
	}
	return k8sClient.Create(context.Background(), deployment)
}

// wait for deployment
func WaitForDeployments(k8sClient client.Client, deployments []string, namespace string, tc openshiftTY.TimeoutConfig) error {
	executeFunc := func() (bool, error) {
		return isDeployed(k8sClient, deployments, namespace)
	}
	return funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, tc.Timeout, tc.ScanInterval, tc.ExpectedSuccessCount)
}

func isDeployed(k8sClient client.Client, deployments []string, namespace string) (bool, error) {
	// check deployment status
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}
	deploymentList, err := List(k8sClient, opts)
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
