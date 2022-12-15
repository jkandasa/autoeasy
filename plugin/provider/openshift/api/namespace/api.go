package api

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
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

func DeleteAndWait(k8sClient client.Client, namespace *corev1.Namespace) error {
	err := utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), namespace))
	if err != nil {
		return err
	}
	tc := openshiftTY.TimeoutConfig{}
	tc.UpdateDefaults()
	tc.ExpectedSuccessCount = 1
	return WaitForDeletion(k8sClient, []string{namespace.Name}, tc)
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

// wait for namespaces deletion
func WaitForDeletion(k8sClient client.Client, namespaces []string, tc openshiftTY.TimeoutConfig) error {
	executeFunc := func() (bool, error) {
		return isAbsent(k8sClient, namespaces)
	}
	return funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, tc.Timeout, tc.ScanInterval, tc.ExpectedSuccessCount)
}

func isAbsent(k8sClient client.Client, namespaces []string) (bool, error) {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	nsList, err := List(k8sClient, opts)
	if err != nil {
		return false, err
	}
	availableList := []string{}
	for _, _namespace := range namespaces {
		found := false
		for _, ns := range nsList.Items {
			if ns.Name == _namespace {
				found = true
			}
		}
		if found {
			availableList = append(availableList, _namespace)
		}
	}

	if len(availableList) == 0 {
		zap.L().Debug("namespaces are absent", zap.Any("namespaces", namespaces))
		return true, nil
	}
	zap.L().Debug("waiting for namespace to be removed", zap.Any("stillPresent", availableList))
	return false, nil
}
