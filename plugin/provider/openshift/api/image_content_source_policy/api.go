package api

import (
	"context"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/store"
	osoperatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(opts []client.ListOption) (*osoperatorv1alpha1.ImageContentSourcePolicyList, error) {
	icspList := &osoperatorv1alpha1.ImageContentSourcePolicyList{}
	err := store.K8SClient.List(context.Background(), icspList, opts...)
	if err != nil {
		return nil, err
	}
	return icspList, nil
}

func Get(name, namespace string) (*osoperatorv1alpha1.ImageContentSourcePolicy, error) {
	icsp := &osoperatorv1alpha1.ImageContentSourcePolicy{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := store.K8SClient.Get(context.Background(), namespacedName, icsp)
	if err != nil {
		return nil, err
	}
	return icsp, nil
}

func Delete(icsp *osoperatorv1alpha1.ImageContentSourcePolicy) error {
	return utils.IgnoreNotFoundError(store.K8SClient.Delete(context.Background(), icsp))
}

func DeleteOfAll(icsp *osoperatorv1alpha1.ImageContentSourcePolicy, opts []client.DeleteAllOfOption) error {
	if icsp == nil {
		icsp = &osoperatorv1alpha1.ImageContentSourcePolicy{}
	}
	return store.K8SClient.DeleteAllOf(context.Background(), icsp, opts...)
}

func Create(icsp *osoperatorv1alpha1.ImageContentSourcePolicy) error {
	return store.K8SClient.Create(context.Background(), icsp)
}

func CreateWithMap(cfg map[string]interface{}) error {
	icsp := &osoperatorv1alpha1.ImageContentSourcePolicy{}
	err := formatterUtils.JsonMapToStruct(cfg, icsp)
	if err != nil {
		return err
	}
	return store.K8SClient.Create(context.Background(), icsp)
}
