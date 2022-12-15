package api

import (
	"context"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*corsosv1alpha1.SubscriptionList, error) {
	subscriptions := &corsosv1alpha1.SubscriptionList{}
	err := k8sClient.List(context.Background(), subscriptions, opts...)
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func Get(k8sClient client.Client, name, namespace string) (*corsosv1alpha1.Subscription, error) {
	subscription := &corsosv1alpha1.Subscription{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(context.Background(), namespacedName, subscription)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func Delete(k8sClient client.Client, subscription *corsosv1alpha1.Subscription) error {
	return utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), subscription))
}

func DeleteOfAll(k8sClient client.Client, subscription *corsosv1alpha1.Subscription, opts []client.DeleteAllOfOption) error {
	if subscription == nil {
		subscription = &corsosv1alpha1.Subscription{}
	}
	return k8sClient.DeleteAllOf(context.Background(), subscription, opts...)
}

func CreateWithMap(k8sClient client.Client, cfg map[string]interface{}) error {
	subscription := &corsosv1alpha1.Subscription{}
	err := formatterUtils.JsonMapToStruct(cfg, subscription)
	if err != nil {
		return err
	}
	return k8sClient.Create(context.Background(), subscription)
}

func Create(k8sClient client.Client, subscription *corsosv1alpha1.Subscription) error {
	return k8sClient.Create(context.Background(), subscription)
}

func Update(k8sClient client.Client, subscription *corsosv1alpha1.Subscription) error {
	return k8sClient.Update(context.Background(), subscription)
}
