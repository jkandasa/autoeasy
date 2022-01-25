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

func List(opts []client.ListOption) (*corsosv1alpha1.SubscriptionList, error) {
	subscriptions := &corsosv1alpha1.SubscriptionList{}
	err := store.K8SClient.List(context.Background(), subscriptions, opts...)
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func Get(name, namespace string) (*corsosv1alpha1.Subscription, error) {
	subscription := &corsosv1alpha1.Subscription{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := store.K8SClient.Get(context.Background(), namespacedName, subscription)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func Delete(subscription *corsosv1alpha1.Subscription) error {
	return utils.IgnoreNotFoundError(store.K8SClient.Delete(context.Background(), subscription))
}

func DeleteOfAll(subscription *corsosv1alpha1.Subscription, opts []client.DeleteAllOfOption) error {
	if subscription == nil {
		subscription = &corsosv1alpha1.Subscription{}
	}
	return store.K8SClient.DeleteAllOf(context.Background(), subscription, opts...)
}

func CreateWithMap(cfg map[string]interface{}) error {
	subscription := &corsosv1alpha1.Subscription{}
	err := formatterUtils.JsonMapToStruct(cfg, subscription)
	if err != nil {
		return err
	}
	return store.K8SClient.Create(context.Background(), subscription)
}

func Create(subscription *corsosv1alpha1.Subscription) error {
	return store.K8SClient.Create(context.Background(), subscription)
}
