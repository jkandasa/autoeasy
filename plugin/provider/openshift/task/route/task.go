package task

import (
	"errors"
	"fmt"

	"github.com/jkandasa/autoeasy/pkg/utils"
	routeAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/route"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	osroutev1 "github.com/openshift/api/route/v1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Run(cfg *openshiftTY.ProviderConfig) (interface{}, error) {
	switch cfg.Function {
	case openshiftTY.FuncAdd:
		if len(cfg.Data) == 0 {
			return nil, fmt.Errorf("no data supplied. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
		}
		return nil, add(cfg)

	case openshiftTY.FuncKeepOnly, openshiftTY.FuncRemove:
		if len(cfg.Data) == 0 {
			return nil, fmt.Errorf("no data supplied. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
		}
		fallthrough
	case openshiftTY.FuncRemoveAll:
		return nil, performDelete(cfg)

	case openshiftTY.FuncGet:
		return get(cfg)
	}

	return nil, fmt.Errorf("unknown function. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
}

func get(cfg *openshiftTY.ProviderConfig) (interface{}, error) {
	cfgRaw := cfg.Data[0]
	routeCfg, ok := cfgRaw.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid data format. expects slice of map[string]interface{}")
	}

	metadata, err := utils.GetObjectMeta(routeCfg)
	if err != nil {
		zap.L().Fatal("error on getting object meta", zap.Any("metadata", metadata), zap.Error(err))
	}
	return routeAPI.Get(metadata.Name, metadata.Namespace)
}

func performDelete(cfg *openshiftTY.ProviderConfig) error {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	routeList, err := routeAPI.List(opts)
	if err != nil {
		zap.L().Fatal("error on getting Route list", zap.Error(err))
	}

	if cfg.Function == openshiftTY.FuncRemoveAll {
		return delete(cfg, routeList.Items)
	} else if cfg.Function == openshiftTY.FuncRemove || cfg.Function == openshiftTY.FuncKeepOnly {
		deletionList := make([]osroutev1.Route, 0)

		suppliedItems := utils.ToNamespacedNameSlice(cfg.Data)

		isRemove := cfg.Function == openshiftTY.FuncRemove

		for _, route := range routeList.Items {
			if isRemove { // remove
				if utils.ContainsNamespacedName(suppliedItems, route.ObjectMeta) {
					deletionList = append(deletionList, route)
				}
			} else { // keep only
				if utils.ContainsNamespacedName(suppliedItems, route.ObjectMeta) {
					deletionList = append(deletionList, route)
				}
			}
		}

		return delete(cfg, deletionList)
	}
	return nil

}

func delete(cfg *openshiftTY.ProviderConfig, items []osroutev1.Route) error {
	if len(items) == 0 {
		return nil
	}
	for _, route := range items {
		err := routeAPI.Delete(&route)
		if err != nil {
			return err
		}
		zap.L().Debug("deleted a Route", zap.String("name", route.Name), zap.String("namespace", route.Name))
	}
	return nil
}

func add(cfg *openshiftTY.ProviderConfig) error {
	for _, cfgRaw := range cfg.Data {
		routeCfg, ok := cfgRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := []client.ListOption{
			client.InNamespace(""),
		}
		routeList, err := routeAPI.List(opts)
		if err != nil {
			zap.L().Fatal("error on getting Route list", zap.Error(err))
		}

		metadata, err := utils.GetObjectMeta(routeCfg)
		if err != nil {
			zap.L().Fatal("error on getting object meta", zap.Any("metadata", metadata), zap.Error(err))
		}
		found := false
		for _, route := range routeList.Items {
			if route.Name == metadata.Name && route.Namespace == metadata.Namespace {
				zap.L().Debug("Route exists", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
				found = true
				if cfg.Config.Recreate {
					zap.L().Debug("Route recreate enabled", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
					err = routeAPI.Delete(&route)
					if err != nil {
						return err
					}
					found = false
				}
				break
			}
		}
		if !found {
			err = routeAPI.CreateWithMap(routeCfg)
			if err != nil {
				zap.L().Fatal("error on creating Route", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace), zap.Error(err))
			}
			zap.L().Info("Route created", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
		}
	}

	return nil
}
