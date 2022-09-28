package api

import (
	"context"

	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*corev1.NodeList, error) {
	nodeList := &corev1.NodeList{}
	err := k8sClient.List(context.Background(), nodeList, opts...)
	if err != nil {
		return nil, err
	}
	return nodeList, nil
}

func Get(k8sClient client.Client, name, namespace string) (*corev1.Node, error) {
	node := &corev1.Node{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(context.Background(), namespacedName, node)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func WaitForNodesReady(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	executeFunc := func() (bool, error) {
		return IsNodesReady(k8sClient)
	}
	tc := cfg.Config.TimeoutConfig
	return funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, tc.Timeout, tc.ScanInterval, tc.ExpectedSuccessCount)
}

func IsNodesReady(k8sClient client.Client) (bool, error) {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	nodeList, err := List(k8sClient, opts)
	unavailable := []string{}
	if err == nil {
		for _, node := range nodeList.Items {
			if node.Spec.Unschedulable {
				unavailable = append(unavailable, node.Name)
			}
		}
	}
	if len(unavailable) == 0 {
		zap.L().Debug("nodes are ready", zap.Any("unavailableNodes", unavailable))
		return true, nil
	}
	zap.L().Debug("waiting for nodes are getting ready", zap.Any("unavailableNodes", unavailable))
	return false, nil
}
