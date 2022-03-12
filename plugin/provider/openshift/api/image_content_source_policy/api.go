package api

import (
	"context"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	osoperatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*osoperatorv1alpha1.ImageContentSourcePolicyList, error) {
	icspList := &osoperatorv1alpha1.ImageContentSourcePolicyList{}
	err := k8sClient.List(context.Background(), icspList, opts...)
	if err != nil {
		return nil, err
	}
	return icspList, nil
}

func Get(k8sClient client.Client, name, namespace string) (*osoperatorv1alpha1.ImageContentSourcePolicy, error) {
	icsp := &osoperatorv1alpha1.ImageContentSourcePolicy{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(context.Background(), namespacedName, icsp)
	if err != nil {
		return nil, err
	}
	return icsp, nil
}

func Delete(k8sClient client.Client, icsp *osoperatorv1alpha1.ImageContentSourcePolicy) error {
	return utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), icsp))
}

func DeleteOfAll(k8sClient client.Client, icsp *osoperatorv1alpha1.ImageContentSourcePolicy, opts []client.DeleteAllOfOption) error {
	if icsp == nil {
		icsp = &osoperatorv1alpha1.ImageContentSourcePolicy{}
	}
	return k8sClient.DeleteAllOf(context.Background(), icsp, opts...)
}

func Create(k8sClient client.Client, icsp *osoperatorv1alpha1.ImageContentSourcePolicy) error {
	return k8sClient.Create(context.Background(), icsp)
}

func CreateWithMap(k8sClient client.Client, cfg map[string]interface{}) error {
	icsp := &osoperatorv1alpha1.ImageContentSourcePolicy{}
	err := formatterUtils.JsonMapToStruct(cfg, icsp)
	if err != nil {
		return err
	}
	return k8sClient.Create(context.Background(), icsp)
}
