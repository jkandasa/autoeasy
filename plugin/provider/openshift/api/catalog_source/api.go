package api

import (
	"context"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*corsosv1alpha1.CatalogSourceList, error) {
	catalogs := &corsosv1alpha1.CatalogSourceList{}
	err := k8sClient.List(context.Background(), catalogs, opts...)
	if err != nil {
		return nil, err
	}
	return catalogs, nil
}

func Get(k8sClient client.Client, name, namespace string) (*corsosv1alpha1.CatalogSource, error) {
	catalog := &corsosv1alpha1.CatalogSource{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(context.Background(), namespacedName, catalog)
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

func Delete(k8sClient client.Client, catalogSource *corsosv1alpha1.CatalogSource) error {
	return utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), catalogSource))

}

func DeleteOfAll(k8sClient client.Client, catalogSource *corsosv1alpha1.CatalogSource, opts []client.DeleteAllOfOption) error {
	if catalogSource == nil {
		catalogSource = &corsosv1alpha1.CatalogSource{}
	}
	return k8sClient.DeleteAllOf(context.Background(), catalogSource, opts...)
}

func CreateWithMap(k8sClient client.Client, cfg map[string]interface{}) error {
	catalogSource := &corsosv1alpha1.CatalogSource{}
	err := formatterUtils.JsonMapToStruct(cfg, catalogSource)
	if err != nil {
		return err
	}
	return k8sClient.Create(context.Background(), catalogSource)
}

func Create(k8sClient client.Client, catalogSource *corsosv1alpha1.CatalogSource) error {
	return k8sClient.Create(context.Background(), catalogSource)
}
