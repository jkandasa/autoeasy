package api

import (
	"context"

	osroutev1 "github.com/openshift/api/route/v1"

	"github.com/jkandasa/autoeasy/pkg/utils"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/store"
	mcUtils "github.com/mycontroller-org/server/v2/pkg/utils"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(opts []client.ListOption) (*osroutev1.RouteList, error) {
	routeList := &osroutev1.RouteList{}
	err := store.K8SClient.List(context.Background(), routeList, opts...)
	if err != nil {
		return nil, err
	}
	return routeList, nil
}

func Get(name string) (*osroutev1.Route, error) {
	route := &osroutev1.Route{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: "",
	}
	err := store.K8SClient.Get(context.Background(), namespacedName, route)
	if err != nil {
		return nil, err
	}
	return route, nil
}

func Delete(route *osroutev1.Route) error {
	return utils.IgnoreNotFoundError(store.K8SClient.Delete(context.Background(), route))
}

func DeleteOfAll(route *osroutev1.Route, opts []client.DeleteAllOfOption) error {
	if route == nil {
		route = &osroutev1.Route{}
	}
	return store.K8SClient.DeleteAllOf(context.Background(), route, opts...)
}

func Create(cfg map[string]interface{}) error {
	route := &osroutev1.Route{}
	err := mcUtils.MapToStruct(mcUtils.TagNameJSON, cfg, route)
	if err != nil {
		return err
	}
	return store.K8SClient.Create(context.Background(), route)
}

func CreateIfNotAvailable(name string) error {
	route := &osroutev1.Route{
		ObjectMeta: v1.ObjectMeta{Name: name},
	}

	list, err := List([]client.ListOption{})
	if err != nil {
		return err
	}
	for _, rxNamespace := range list.Items {
		if route.Name == rxNamespace.Name {
			return nil
		}
	}

	return store.K8SClient.Create(context.Background(), route)
}
