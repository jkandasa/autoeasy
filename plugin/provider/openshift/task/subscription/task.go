package task

import (
	"github.com/jkandasa/autoeasy/pkg/utils"
	operatorAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/operator"
	subscriptionAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/subscription"
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

	}

	return nil, nil
}

func performDelete(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	subscriptionList, err := subscriptionAPI.List(k8sClient, opts)
	if err != nil {
		zap.L().Fatal("error on getting Subscription list", zap.Error(err))
		return err
	}

	if cfg.Function == openshiftTY.FuncRemoveAll {
		return delete(k8sClient, cfg, subscriptionList.Items)
	} else if cfg.Function == openshiftTY.FuncRemove || cfg.Function == openshiftTY.FuncKeepOnly {
		deletionList := make([]corsosv1alpha1.Subscription, 0)

		suppliedItems := utils.ToNamespacedNameSlice(cfg.Data)

		isRemove := cfg.Function == openshiftTY.FuncRemove

		for _, subscription := range subscriptionList.Items {
			if isRemove { // remove
				if utils.ContainsNamespacedName(suppliedItems, subscription.ObjectMeta) {
					deletionList = append(deletionList, subscription)
				}
			} else { // keep only
				if !utils.ContainsNamespacedName(suppliedItems, subscription.ObjectMeta) {
					deletionList = append(deletionList, subscription)
				}
			}
		}

		return delete(k8sClient, cfg, deletionList)
	}
	return nil

}

func delete(k8sClient client.Client, cfg *openshiftTY.ProviderConfig, items []corsosv1alpha1.Subscription) error {
	if len(items) == 0 {
		return nil
	}
	for _, cs := range items {
		err := subscriptionAPI.Delete(k8sClient, &cs)
		if err != nil {
			return err
		}
		zap.L().Debug("deleted a Subscription", zap.String("name", cs.Name), zap.String("namespace", cs.Name))
	}
	return nil
}

func add(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	if len(cfg.Data) == 0 {
		// TODO: report error
		return nil
	}

	for _, cfgRaw := range cfg.Data {
		subscriptionCfg, ok := cfgRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := []client.ListOption{
			client.InNamespace(""),
		}
		csList, err := subscriptionAPI.List(k8sClient, opts)
		if err != nil {
			zap.L().Fatal("error on getting Subscription list", zap.Error(err))
		}

		metadata, err := utils.GetObjectMeta(subscriptionCfg)
		if err != nil {
			zap.L().Fatal("error on getting object meta", zap.Any("metadata", metadata), zap.Error(err))
		}
		found := false
		for _, sub := range csList.Items {
			if sub.ObjectMeta.Name == metadata.Name && sub.Namespace == metadata.Namespace {
				zap.L().Debug("Subscription exists", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
				found = true
				if cfg.Config.Recreate {
					zap.L().Debug("Subscription recreate enabled", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
					err = operatorAPI.Uninstall(k8sClient, subscriptionCfg)
					if err != nil {
						return err
					}
					zap.L().Debug("Subscription removed", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
					found = false
				}
				break
			}
		}
		if !found {
			err = operatorAPI.Install(k8sClient, subscriptionCfg, cfg.Config.TimeoutConfig)
			if err != nil {
				zap.L().Fatal("error on adding a Subscription", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace), zap.Error(err))
				return err
			}
			zap.L().Debug("Subscription created", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))

		}
	}

	return nil
}
