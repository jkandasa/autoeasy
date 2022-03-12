package api

import (
	"context"
	"fmt"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*corsosv1alpha1.ClusterServiceVersionList, error) {
	csvList := &corsosv1alpha1.ClusterServiceVersionList{}
	err := k8sClient.List(context.Background(), csvList, opts...)
	if err != nil {
		return nil, err
	}
	return csvList, nil
}

func Get(k8sClient client.Client, name, namespace string) (*corsosv1alpha1.ClusterServiceVersion, error) {
	csv := &corsosv1alpha1.ClusterServiceVersion{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(context.Background(), namespacedName, csv)
	if err != nil {
		return nil, err
	}
	return csv, nil
}

func Delete(k8sClient client.Client, csv *corsosv1alpha1.ClusterServiceVersion) error {
	return utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), csv))
}

func DeleteOfAll(k8sClient client.Client, csv *corsosv1alpha1.ClusterServiceVersion, opts []client.DeleteAllOfOption) error {
	if csv == nil {
		csv = &corsosv1alpha1.ClusterServiceVersion{}
	}
	return k8sClient.DeleteAllOf(context.Background(), csv, opts...)
}

func Create(k8sClient client.Client, csv *corsosv1alpha1.ClusterServiceVersion) error {
	return k8sClient.Create(context.Background(), csv)
}

func CreateWithMap(k8sClient client.Client, cfg map[string]interface{}) error {
	csv := &corsosv1alpha1.ClusterServiceVersion{}
	err := formatterUtils.JsonMapToStruct(cfg, csv)
	if err != nil {
		return err
	}
	return k8sClient.Create(context.Background(), csv)
}

func Info(k8sClient client.Client) ([]openshiftTY.Info, error) {
	opts := []client.ListOption{
		// check only rom default namespace, otherwise we will duplicate in all namespaces
		client.InNamespace("default"),
	}

	csvList, err := List(k8sClient, opts)
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
