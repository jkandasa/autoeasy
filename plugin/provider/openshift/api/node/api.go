package api

import (
	"context"

	"github.com/jkandasa/autoeasy/plugin/provider/openshift/store"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(opts []client.ListOption) (*corev1.NodeList, error) {
	nodeList := &corev1.NodeList{}
	err := store.K8SClient.List(context.Background(), nodeList, opts...)
	if err != nil {
		return nil, err
	}
	return nodeList, nil
}

func Get(name, namespace string) (*corev1.Node, error) {
	node := &corev1.Node{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := store.K8SClient.Get(context.Background(), namespacedName, node)
	if err != nil {
		return nil, err
	}
	return node, nil
}
