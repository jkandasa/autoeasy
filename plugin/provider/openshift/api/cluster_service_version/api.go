package api

import (
	"context"
	"fmt"

	"github.com/jkandasa/autoeasy/pkg/utils"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/store"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	mcUtils "github.com/mycontroller-org/server/v2/pkg/utils"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(opts []client.ListOption) (*corsosv1alpha1.ClusterServiceVersionList, error) {
	csvList := &corsosv1alpha1.ClusterServiceVersionList{}
	err := store.K8SClient.List(context.Background(), csvList, opts...)
	if err != nil {
		return nil, err
	}
	return csvList, nil
}

func Get(name, namespace string) (*corsosv1alpha1.ClusterServiceVersion, error) {
	csv := &corsosv1alpha1.ClusterServiceVersion{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := store.K8SClient.Get(context.Background(), namespacedName, csv)
	if err != nil {
		return nil, err
	}
	return csv, nil
}

func Delete(csv *corsosv1alpha1.ClusterServiceVersion) error {
	return utils.IgnoreNotFoundError(store.K8SClient.Delete(context.Background(), csv))
}

func DeleteOfAll(csv *corsosv1alpha1.ClusterServiceVersion, opts []client.DeleteAllOfOption) error {
	if csv == nil {
		csv = &corsosv1alpha1.ClusterServiceVersion{}
	}
	return store.K8SClient.DeleteAllOf(context.Background(), csv, opts...)
}

func Create(cfg map[string]interface{}) error {
	csv := &corsosv1alpha1.ClusterServiceVersion{}
	err := mcUtils.MapToStruct(mcUtils.TagNameJSON, cfg, csv)
	if err != nil {
		return err
	}
	return store.K8SClient.Create(context.Background(), csv)
}

func Info() ([]openshiftTY.Info, error) {
	opts := []client.ListOption{
		// check only rom default namespace, otherwise we will duplicate in all namespaces
		client.InNamespace("default"),
	}

	csvList, err := List(opts)
	if err != nil {
		return nil, err
	}

	fmt.Println("items count:", len(csvList.Items))

	csvs := make([]openshiftTY.Info, 0)
	for _, csv := range csvList.Items {
		images := []string{}
		for _, img := range csv.Spec.RelatedImages {
			images = append(images, img.Image)
		}
		csvs = append(csvs, openshiftTY.Info{
			Name:        csv.Name,
			DisplayName: csv.Spec.DisplayName,
			Version:     csv.Spec.Version,
			Phase:       csv.Status.Phase,
			Images:      images,
		})
	}

	return csvs, nil
}
