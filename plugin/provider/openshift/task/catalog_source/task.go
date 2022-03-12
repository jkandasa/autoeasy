package task

import (
	"fmt"
	"strings"

	"github.com/jkandasa/autoeasy/pkg/utils"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	csAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/catalog_source"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Run(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) (interface{}, error) {
	switch cfg.Function {
	case openshiftTY.FuncAdd:
		return nil, add(k8sClient, cfg)

	case openshiftTY.FuncKeepOnly, openshiftTY.FuncRemove, openshiftTY.FuncRemoveAll:
		return nil, performDelete(k8sClient, cfg)

	default:
		return nil, fmt.Errorf("invalid function. kind:%s, function:%s", cfg.Kind, cfg.Function)
	}

}

func performDelete(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	csList, err := csAPI.List(k8sClient, opts)
	if err != nil {
		zap.L().Fatal("error on getting CatalogSource list", zap.Error(err))
	}

	if cfg.Function == openshiftTY.FuncRemoveAll {
		return delete(k8sClient, cfg, csList.Items)
	} else if cfg.Function == openshiftTY.FuncRemoveAll || cfg.Function == openshiftTY.FuncKeepOnly {
		deletionList := make([]corsosv1alpha1.CatalogSource, 0)

		suppliedItems := utils.ToStringSlice(cfg.Data)

		isRemove := cfg.Function == openshiftTY.FuncRemove

		for _, cs := range csList.Items {
			if isRemove { // remove
				if utils.ContainsString(suppliedItems, cs.Name) {
					deletionList = append(deletionList, cs)
				}
			} else { // keep only
				if !utils.ContainsString(suppliedItems, cs.Name) {
					deletionList = append(deletionList, cs)
				}
			}
		}

		return delete(k8sClient, cfg, deletionList)
	}
	return nil

}

func delete(k8sClient client.Client, cfg *openshiftTY.ProviderConfig, items []corsosv1alpha1.CatalogSource) error {
	if len(items) == 0 {
		return nil
	}
	for _, cs := range items {
		err := csAPI.Delete(k8sClient, &cs)
		if err != nil {
			return err
		}
		zap.L().Debug("deleted a CatalogSource", zap.String("name", cs.Name), zap.String("namespace", cs.Name))
	}
	return nil
}

func add(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	if len(cfg.Data) == 0 {
		// TODO: report error
		return nil
	}

	for _, cfgRaw := range cfg.Data {
		icspCfg, ok := cfgRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := []client.ListOption{
			client.InNamespace(""),
		}
		csList, err := csAPI.List(k8sClient, opts)
		if err != nil {
			zap.L().Fatal("error on getting CatalogSource list", zap.Error(err))
		}

		metadata, err := utils.GetObjectMeta(icspCfg)
		if err != nil {
			zap.L().Fatal("error on getting object meta", zap.Any("metadata", metadata), zap.Error(err))
		}
		found := false
		for _, icsp := range csList.Items {
			if icsp.ObjectMeta.Name == metadata.Name {
				zap.L().Debug("CatalogSource exists", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
				found = true
				if cfg.Config.Recreate {
					zap.L().Debug("CatalogSource recreate enabled", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
					err = csAPI.Delete(k8sClient, &icsp)
					if err != nil {
						return err
					}
					found = false
				}
				break
			}
		}
		if !found {
			err = csAPI.CreateWithMap(k8sClient, icspCfg)
			if err != nil {
				zap.L().Fatal("error on creating CatalogSource", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace), zap.Error(err))
			}
			zap.L().Info("CatalogSource created", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
			err = waitForCatalogSource(k8sClient, cfg, metadata.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func waitForCatalogSource(k8sClient client.Client, cfg *openshiftTY.ProviderConfig, name string) error {
	executeFunc := func() (bool, error) {
		return isReady(k8sClient, name)
	}
	tc := cfg.Config.TimeoutConfig
	return funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, tc.Timeout, tc.ScanInterval, tc.ExpectedSuccessCount)
}

func isReady(k8sClient client.Client, name string) (bool, error) {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	csList, err := csAPI.List(k8sClient, opts)
	if err == nil {
		for _, cs := range csList.Items {
			if cs.Name == name {
				if cs.Status.GRPCConnectionState != nil && strings.ToLower(cs.Status.GRPCConnectionState.LastObservedState) == "ready" {
					zap.L().Debug("CatalogSource is ready", zap.Any("name", name), zap.Any("namespace", cs.Namespace))
					return true, nil
				}
			}
		}
	} else {
		return false, err
	}

	zap.L().Debug("waiting for CatalogSource is getting ready", zap.Any("name", name))
	return false, nil
}
