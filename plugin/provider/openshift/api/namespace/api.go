package api

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*corev1.NamespaceList, error) {
	namespaceList := &corev1.NamespaceList{}
	err := k8sClient.List(context.Background(), namespaceList, opts...)
	if err != nil {
		return nil, err
	}
	return namespaceList, nil
}

func Get(k8sClient client.Client, name string) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: "",
	}
	err := k8sClient.Get(context.Background(), namespacedName, namespace)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

func Delete(k8sClient client.Client, namespace *corev1.Namespace) error {
	return utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), namespace))
}

func DeleteOfAll(k8sClient client.Client, namespace *corev1.Namespace, opts []client.DeleteAllOfOption) error {
	if namespace == nil {
		namespace = &corev1.Namespace{}
	}
	return k8sClient.DeleteAllOf(context.Background(), namespace, opts...)
}

func Create(k8sClient client.Client, namespace *corev1.Namespace) error {
	return k8sClient.Create(context.Background(), namespace)
}

func CreateWithMap(k8sClient client.Client, cfg map[string]interface{}) error {
	namespace := &corev1.Namespace{}
	err := formatterUtils.JsonMapToStruct(cfg, namespace)
	if err != nil {
		return err
	}
	return k8sClient.Create(context.Background(), namespace)
}

func CreateIfNotAvailable(k8sClient client.Client, name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{Name: name},
	}

	list, err := List(k8sClient, []client.ListOption{})
	if err != nil {
		return err
	}
	for _, rxNamespace := range list.Items {
		if namespace.Name == rxNamespace.Name {
			return nil
		}
	}

	return k8sClient.Create(context.Background(), namespace)
}
