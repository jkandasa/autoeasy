package task

import (
	"github.com/jkandasa/autoeasy/pkg/utils"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	icspAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/image_content_source_policy"
	nodeAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/node"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"github.com/openshift/api/operator/v1alpha1"
	"go.uber.org/zap"
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
	icspList, err := icspAPI.List(k8sClient, opts)
	if err != nil {
		zap.L().Fatal("error on getting imageContentSourcePolicy list", zap.Error(err))
	}

	if cfg.Function == openshiftTY.FuncRemoveAll {
		return delete(k8sClient, cfg, icspList.Items)
	} else if cfg.Function == openshiftTY.FuncRemoveAll || cfg.Function == openshiftTY.FuncKeepOnly {
		deletionList := make([]v1alpha1.ImageContentSourcePolicy, 0)

		suppliedItems := utils.ToStringSlice(cfg.Data)

		isRemove := cfg.Function == openshiftTY.FuncRemove

		for _, icsp := range icspList.Items {
			if isRemove { // remove
				if utils.ContainsString(suppliedItems, icsp.Name) {
					deletionList = append(deletionList, icsp)
				}
			} else { // keep only
				if !utils.ContainsString(suppliedItems, icsp.Name) {
					deletionList = append(deletionList, icsp)
				}
			}
		}

		return delete(k8sClient, cfg, deletionList)
	}
	return nil

}

func delete(k8sClient client.Client, cfg *openshiftTY.ProviderConfig, items []v1alpha1.ImageContentSourcePolicy) error {
	if len(items) == 0 {
		return nil
	}
	for _, icsp := range items {
		err := icspAPI.Delete(k8sClient, &icsp)
		if err != nil {
			return err
		}
		zap.L().Debug("deleted a ImageContentSourcePolicy", zap.String("name", icsp.Name))
	}
	return waitForNodes(k8sClient, cfg)
}

func add(k8sClient client.Client, task *openshiftTY.ProviderConfig) error {
	if len(task.Data) == 0 {
		// TODO: report error
		return nil
	}

	for _, cfgRaw := range task.Data {
		icspCfg, ok := cfgRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := []client.ListOption{
			client.InNamespace(""),
		}
		icspList, err := icspAPI.List(k8sClient, opts)
		if err != nil {
			zap.L().Fatal("error on getting imageContentSourcePolicy list", zap.Error(err))
		}

		metadata, err := utils.GetObjectMeta(icspCfg)
		if err != nil {
			zap.L().Fatal("error on getting object meta", zap.Any("metadata", metadata), zap.Error(err))
		}
		found := false
		for _, icsp := range icspList.Items {
			if icsp.ObjectMeta.Name == metadata.Name {
				zap.L().Debug("imageContentSourcePolicy exists", zap.String("name", metadata.Name))
				found = true
				if task.Config.Recreate {
					zap.L().Debug("imageContentSourcePolicy recreate enabled", zap.String("name", metadata.Name))
					err = icspAPI.Delete(k8sClient, &icsp)
					if err != nil {
						return err
					}
					found = false
				}
				break
			}
		}
		if !found {
			err = icspAPI.CreateWithMap(k8sClient, icspCfg)
			if err != nil {
				zap.L().Fatal("error on creating imageContentSourcePolicy", zap.String("name", metadata.Name), zap.Error(err))
			}
			zap.L().Info("imageContentSourcePolicy created", zap.String("name", metadata.Name))

		}
	}

	return waitForNodes(k8sClient, task)
}

func waitForNodes(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	executeFunc := func() (bool, error) {
		return isNodesReady(k8sClient)
	}
	tc := cfg.Config.TimeoutConfig
	return funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, tc.Timeout, tc.ScanInterval, tc.ExpectedSuccessCount)
}

func isNodesReady(k8sClient client.Client) (bool, error) {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	nodeList, err := nodeAPI.List(k8sClient, opts)
	unavailable := []string{}
	if err == nil {
		for _, node := range nodeList.Items {
			if node.Spec.Unschedulable {
				unavailable = append(unavailable, node.Name)
			}
		}
	}
	if len(unavailable) == 0 {
		zap.L().Debug("nodes are ready", zap.Any("unavailableNodes", unavailable))
		return true, nil
	}
	zap.L().Debug("waiting for nodes are getting ready", zap.Any("unavailableNodes", unavailable))
	return false, nil
}
