package api

import (
	"context"

	jaegerv1 "github.com/jaegertracing/jaeger-operator/apis/v1"
	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"

	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*jaegerv1.JaegerList, error) {
	jaegerList := &jaegerv1.JaegerList{}
	err := k8sClient.List(context.Background(), jaegerList, opts...)
	if err != nil {
		return nil, err
	}
	return jaegerList, nil
}

func Get(k8sClient client.Client, name, namespace string) (*jaegerv1.Jaeger, error) {
	jaeger := &jaegerv1.Jaeger{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(context.Background(), namespacedName, jaeger)
	if err != nil {
		return nil, err
	}
	return jaeger, nil
}

func Delete(k8sClient client.Client, jaeger *jaegerv1.Jaeger) error {
	return utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), jaeger))
}

func DeleteOfAll(k8sClient client.Client, jaeger *jaegerv1.Jaeger, opts []client.DeleteAllOfOption) error {
	if jaeger == nil {
		jaeger = &jaegerv1.Jaeger{}
	}
	return k8sClient.DeleteAllOf(context.Background(), jaeger, opts...)
}

func Create(k8sClient client.Client, jaeger *jaegerv1.Jaeger) error {
	return k8sClient.Create(context.Background(), jaeger)
}

func CreateAndWait(k8sClient client.Client, jaeger *jaegerv1.Jaeger, timeoutConfig openshiftTY.TimeoutConfig) error {
	err := k8sClient.Create(context.Background(), jaeger)
	if err != nil {
		return err
	}
	executeFunc := func() (bool, error) {
		return isRunning(k8sClient, jaeger.Name, jaeger.Namespace)
	}

	return funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, timeoutConfig.Timeout, timeoutConfig.ScanInterval, timeoutConfig.ExpectedSuccessCount)
}

func CreateWithMap(k8sClient client.Client, cfg map[string]interface{}) error {
	jaeger := &jaegerv1.Jaeger{}
	err := formatterUtils.JsonMapToStruct(cfg, jaeger)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return k8sClient.Create(context.Background(), jaeger)
}

func CreateWithMapAndWait(k8sClient client.Client, cfg map[string]interface{}) error {
	jaeger := &jaegerv1.Jaeger{}
	err := formatterUtils.JsonMapToStruct(cfg, jaeger)
	if err != nil {
		return err
	}
	err = k8sClient.Create(context.Background(), jaeger)
	if err != nil {
		return err
	}
	executeFunc := func() (bool, error) {
		return isRunning(k8sClient, jaeger.Name, jaeger.Namespace)
	}
	return funcUtils.ExecuteWithDefaultTimeoutAndContinuesSuccessCount(executeFunc)
}

func isRunning(k8sClient client.Client, name, namespace string) (bool, error) {
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}

	jaegerList, err := List(k8sClient, opts)
	if err != nil {
		return false, err
	}
	for _, rxJaeger := range jaegerList.Items {
		if name == rxJaeger.Name {
			if rxJaeger.Status.Phase == jaegerv1.JaegerPhaseRunning {
				zap.L().Debug("jaeger is running", zap.String("name", name), zap.String("namespace", namespace))
				return true, nil
			}
			zap.L().Debug("waiting for the jaeger status running", zap.String("name", name), zap.String("namespace", namespace), zap.String("status", string(rxJaeger.Status.Phase)))
			break
		}
	}
	return false, nil
}
