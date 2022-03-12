package task

import (
	"github.com/jkandasa/autoeasy/pkg/utils"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	nsAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/namespace"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Run(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) (interface{}, error) {
	switch cfg.Function {
	case openshiftTY.FuncAdd:
		return nil, add(k8sClient, cfg)

	case openshiftTY.FuncKeepOnly, openshiftTY.FuncRemove, openshiftTY.FuncRemoveAll:
		return nil, performDelete(k8sClient, cfg)

	}

	return nil, nil
}

func performDelete(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	nsList, err := nsAPI.List(k8sClient, opts)
	if err != nil {
		zap.L().Fatal("error on getting Namespace list", zap.Error(err))
	}

	if cfg.Function == openshiftTY.FuncRemoveAll {
		return delete(k8sClient, cfg, nsList.Items)
	} else if cfg.Function == openshiftTY.FuncRemoveAll || cfg.Function == openshiftTY.FuncKeepOnly {
		deletionList := make([]corev1.Namespace, 0)

		suppliedItems := utils.ToStringSlice(cfg.Data)

		isRemove := cfg.Function == openshiftTY.FuncRemove

		for _, ns := range nsList.Items {
			if isRemove { // remove
				if utils.ContainsString(suppliedItems, ns.Name) {
					deletionList = append(deletionList, ns)
				}
			} else { // keep only
				if !utils.ContainsString(suppliedItems, ns.Name) {
					deletionList = append(deletionList, ns)
				}
			}
		}

		return delete(k8sClient, cfg, deletionList)
	}
	return nil

}

func delete(k8sClient client.Client, cfg *openshiftTY.ProviderConfig, items []corev1.Namespace) error {
	if len(items) == 0 {
		return nil
	}
	for _, ns := range items {
		err := nsAPI.Delete(k8sClient, &ns)
		if err != nil {
			return err
		}
		zap.L().Debug("deleted a Namespace", zap.String("namespace", ns.Name))
		err = waitForDeletion(k8sClient, cfg, ns.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func add(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	if len(cfg.Data) == 0 {
		// TODO: report error
		return nil
	}

	for _, cfgRaw := range cfg.Data {
		nsCfg, ok := cfgRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := []client.ListOption{
			client.InNamespace(""),
		}
		nsList, err := nsAPI.List(k8sClient, opts)
		if err != nil {
			zap.L().Fatal("error on getting Namespace list", zap.Error(err))
		}

		metadata, err := utils.GetObjectMeta(nsCfg)
		if err != nil {
			zap.L().Fatal("error on getting object meta", zap.Any("metadata", metadata), zap.Error(err))
		}
		found := false
		for _, ns := range nsList.Items {
			if ns.ObjectMeta.Name == metadata.Name {
				zap.L().Debug("Namespace exists", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
				found = true
				if cfg.Config.Recreate {
					zap.L().Debug("Namespace recreate enabled", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
					err = delete(k8sClient, cfg, []corev1.Namespace{ns})
					if err != nil {
						return err
					}
					found = false
				}
				break
			}
		}
		if !found {
			err = nsAPI.CreateWithMap(k8sClient, nsCfg)
			if err != nil {
				zap.L().Fatal("error on creating Namespace", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace), zap.Error(err))
			}
			zap.L().Info("Namespace created", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
		}
	}

	return nil
}

func waitForDeletion(k8sClient client.Client, cfg *openshiftTY.ProviderConfig, namespace string) error {
	executeFunc := func() (bool, error) {
		return isDeleted(k8sClient, namespace)
	}
	tc := cfg.Config.TimeoutConfig
	return funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, tc.Timeout, tc.ScanInterval, tc.ExpectedSuccessCount)
}

func isDeleted(k8sClient client.Client, namespace string) (bool, error) {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	nsList, err := nsAPI.List(k8sClient, opts)
	if err == nil {
		for _, cs := range nsList.Items {
			if cs.Name == namespace {
				zap.L().Debug("waiting for Namespace to removed", zap.Any("namespace", namespace))
				return false, nil
			}
		}
	}

	return true, nil
}
