package api

import (
	"context"

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
