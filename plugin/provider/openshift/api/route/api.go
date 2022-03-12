package api

import (
	"context"

	osroutev1 "github.com/openshift/api/route/v1"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*osroutev1.RouteList, error) {
	routeList := &osroutev1.RouteList{}
	err := k8sClient.List(context.Background(), routeList, opts...)
	if err != nil {
		return nil, err
	}
	return routeList, nil
}

func Get(k8sClient client.Client, name, namespace string) (*osroutev1.Route, error) {
	route := &osroutev1.Route{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(context.Background(), namespacedName, route)
	if err != nil {
		return nil, err
	}
	return route, nil
}

func Delete(k8sClient client.Client, route *osroutev1.Route) error {
	return utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), route))
}

func DeleteOfAll(k8sClient client.Client, route *osroutev1.Route, opts []client.DeleteAllOfOption) error {
	if route == nil {
		route = &osroutev1.Route{}
	}
	return k8sClient.DeleteAllOf(context.Background(), route, opts...)
}

func CreateWithMap(k8sClient client.Client, cfg map[string]interface{}) error {
	route := &osroutev1.Route{}
	err := formatterUtils.JsonMapToStruct(cfg, route)
	if err != nil {
		return err
	}
	return k8sClient.Create(context.Background(), route)
}

func Create(k8sClient client.Client, route *osroutev1.Route) error {
	return k8sClient.Create(context.Background(), route)
}
