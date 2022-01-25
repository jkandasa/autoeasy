package api

import (
	"context"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/store"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(opts []client.ListOption) (*corsosv1alpha1.CatalogSourceList, error) {
	catalogs := &corsosv1alpha1.CatalogSourceList{}
	err := store.K8SClient.List(context.Background(), catalogs, opts...)
	if err != nil {
		return nil, err
	}
	return catalogs, nil
}

func Get(name, namespace string) (*corsosv1alpha1.CatalogSource, error) {
	catalog := &corsosv1alpha1.CatalogSource{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := store.K8SClient.Get(context.Background(), namespacedName, catalog)
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

func Delete(catalogSource *corsosv1alpha1.CatalogSource) error {
	return utils.IgnoreNotFoundError(store.K8SClient.Delete(context.Background(), catalogSource))

}

func DeleteOfAll(catalogSource *corsosv1alpha1.CatalogSource, opts []client.DeleteAllOfOption) error {
	if catalogSource == nil {
		catalogSource = &corsosv1alpha1.CatalogSource{}
	}
	return store.K8SClient.DeleteAllOf(context.Background(), catalogSource, opts...)
}

func CreateWithMap(cfg map[string]interface{}) error {
	catalogSource := &corsosv1alpha1.CatalogSource{}
	err := formatterUtils.JsonMapToStruct(cfg, catalogSource)
	if err != nil {
		return err
	}
	return store.K8SClient.Create(context.Background(), catalogSource)
}

func Create(catalogSource *corsosv1alpha1.CatalogSource) error {
	return store.K8SClient.Create(context.Background(), catalogSource)
}
