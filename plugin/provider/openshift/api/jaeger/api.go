package api

import (
	"context"

	jaegerv1 "github.com/jaegertracing/jaeger-operator/apis/v1"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/store"
	"github.com/jkandasa/autoeasy/pkg/utils"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	mcUtils "github.com/mycontroller-org/server/v2/pkg/utils"

	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(opts []client.ListOption) (*jaegerv1.JaegerList, error) {
	jaegerList := &jaegerv1.JaegerList{}
	err := store.K8SClient.List(context.Background(), jaegerList, opts...)
	if err != nil {
		return nil, err
	}
	return jaegerList, nil
}

func Get(name, namespace string) (*jaegerv1.Jaeger, error) {
	jaeger := &jaegerv1.Jaeger{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := store.K8SClient.Get(context.Background(), namespacedName, jaeger)
	if err != nil {
		return nil, err
	}
	return jaeger, nil
}

func Delete(jaeger *jaegerv1.Jaeger) error {
	return utils.IgnoreNotFoundError(store.K8SClient.Delete(context.Background(), jaeger))
}

func DeleteOfAll(jaeger *jaegerv1.Jaeger, opts []client.DeleteAllOfOption) error {
	if jaeger == nil {
		jaeger = &jaegerv1.Jaeger{}
	}
	return store.K8SClient.DeleteAllOf(context.Background(), jaeger, opts...)
}

func Create(cfg map[string]interface{}) error {
	jaeger := &jaegerv1.Jaeger{}
	err := mcUtils.MapToStruct(mcUtils.TagNameJSON, cfg, jaeger)
	if err != nil {
		return err
	}
	return store.K8SClient.Create(context.Background(), jaeger)
}

func CreateAndWait(cfg map[string]interface{}) error {
	jaeger := &jaegerv1.Jaeger{}
	err := mcUtils.MapToStruct(mcUtils.TagNameJSON, cfg, jaeger)
	if err != nil {
		return err
	}
	err = store.K8SClient.Create(context.Background(), jaeger)
	if err != nil {
		return err
	}

	executeFunc := func() (bool, error) {
		return isRunning(jaeger.Name, jaeger.Namespace)
	}
	return funcUtils.ExecuteWithDefaultTimeoutAndContinuesSuccessCount(executeFunc)
}

func isRunning(name, namespace string) (bool, error) {
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}

	jaegerList, err := List(opts)
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
