package api

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/store"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(opts []client.ListOption) (*corev1.NamespaceList, error) {
	namespaceList := &corev1.NamespaceList{}
	err := store.K8SClient.List(context.Background(), namespaceList, opts...)
	if err != nil {
		return nil, err
	}
	return namespaceList, nil
}

func Get(name string) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: "",
	}
	err := store.K8SClient.Get(context.Background(), namespacedName, namespace)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

func Delete(namespace *corev1.Namespace) error {
	return utils.IgnoreNotFoundError(store.K8SClient.Delete(context.Background(), namespace))
}

func DeleteOfAll(namespace *corev1.Namespace, opts []client.DeleteAllOfOption) error {
	if namespace == nil {
		namespace = &corev1.Namespace{}
	}
	return store.K8SClient.DeleteAllOf(context.Background(), namespace, opts...)
}

func Create(namespace *corev1.Namespace) error {
	return store.K8SClient.Create(context.Background(), namespace)
}

func CreateWithMap(cfg map[string]interface{}) error {
	namespace := &corev1.Namespace{}
	err := formatterUtils.JsonMapToStruct(cfg, namespace)
	if err != nil {
		return err
	}
	return store.K8SClient.Create(context.Background(), namespace)
}

func CreateIfNotAvailable(name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{Name: name},
	}

	list, err := List([]client.ListOption{})
	if err != nil {
		return err
	}
	for _, rxNamespace := range list.Items {
		if namespace.Name == rxNamespace.Name {
			return nil
		}
	}

	return store.K8SClient.Create(context.Background(), namespace)
}
