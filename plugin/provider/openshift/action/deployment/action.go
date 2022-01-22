package action

import (
	"fmt"

	"github.com/jkandasa/autoeasy/pkg/utils"
	deploymentAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/deployment"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Run(cfg *openshiftTY.ProviderConfig) error {
	switch cfg.Function {
	case openshiftTY.FuncAdd:
		if len(cfg.Data) == 0 {
			return fmt.Errorf("no data supplied. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
		}
		return add(cfg)

	case openshiftTY.FuncKeepOnly, openshiftTY.FuncRemove:
		if len(cfg.Data) == 0 {
			return fmt.Errorf("no data supplied. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
		}
		fallthrough
	case openshiftTY.FuncRemoveAll:
		return performDelete(cfg)

	case openshiftTY.FuncWaitForReady:
		if len(cfg.Data) == 0 {
			return fmt.Errorf("no data supplied. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
		}
		return waitForReady(cfg)

	}

	return fmt.Errorf("unknown function. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
}

func waitForReady(cfg *openshiftTY.ProviderConfig) error {
	// get deployments detail
	suppliedItems := utils.ToNamespacedNameSlice(cfg.Data)

	// group by namespace
	items := map[string][]string{}
	for _, deployment := range suppliedItems {
		deployments, ok := items[deployment.Namespace]
		if !ok {
			deployments = make([]string, 0)
		}
		deployments = append(deployments, deployment.Name)
		items[deployment.Namespace] = deployments
	}

	// verify status
	for namespace, deployments := range items {
		err := deploymentAPI.WaitForDeployments(deployments, namespace, cfg.Config.TimeoutConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func performDelete(cfg *openshiftTY.ProviderConfig) error {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	deploymentList, err := deploymentAPI.List(opts)
	if err != nil {
		zap.L().Fatal("error on getting Deployment list", zap.Error(err))
	}

	if cfg.Function == openshiftTY.FuncRemoveAll {
		return delete(cfg, deploymentList.Items)
	} else if cfg.Function == openshiftTY.FuncRemove || cfg.Function == openshiftTY.FuncKeepOnly {
		deletionList := make([]appsv1.Deployment, 0)

		suppliedItems := utils.ToNamespacedNameSlice(cfg.Data)

		isRemove := cfg.Function == openshiftTY.FuncRemove

		for _, deployment := range deploymentList.Items {
			if isRemove { // remove
				if utils.ContainsNamespacedName(suppliedItems, deployment.ObjectMeta) {
					deletionList = append(deletionList, deployment)
				}
			} else { // keep only
				if utils.ContainsNamespacedName(suppliedItems, deployment.ObjectMeta) {
					deletionList = append(deletionList, deployment)
				}
			}
		}

		return delete(cfg, deletionList)
	}
	return nil

}

func delete(cfg *openshiftTY.ProviderConfig, items []appsv1.Deployment) error {
	if len(items) == 0 {
		return nil
	}
	for _, deployment := range items {
		err := deploymentAPI.Delete(&deployment)
		if err != nil {
			return err
		}
		zap.L().Debug("deleted a Deployment", zap.String("name", deployment.Name), zap.String("namespace", deployment.Name))
	}
	return nil
}

func add(cfg *openshiftTY.ProviderConfig) error {
	for _, cfgRaw := range cfg.Data {
		deploymentCfg, ok := cfgRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := []client.ListOption{
			client.InNamespace(""),
		}
		deploymentList, err := deploymentAPI.List(opts)
		if err != nil {
			zap.L().Fatal("error on getting Deployment list", zap.Error(err))
		}

		metadata, err := utils.GetObjectMeta(deploymentCfg)
		if err != nil {
			zap.L().Fatal("error on getting object meta", zap.Any("metadata", metadata), zap.Error(err))
		}
		found := false
		for _, deployment := range deploymentList.Items {
			if deployment.Name == metadata.Name && deployment.Namespace == metadata.Namespace {
				zap.L().Debug("Deployment exists", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
				found = true
				if cfg.Config.Recreate {
					zap.L().Debug("Deployment recreate enabled", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
					err = deploymentAPI.Delete(&deployment)
					if err != nil {
						return err
					}
					found = false
				}
				break
			}
		}
		if !found {
			err = deploymentAPI.Create(deploymentCfg)
			if err != nil {
				zap.L().Fatal("error on creating Deployment", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace), zap.Error(err))
			}
			zap.L().Info("Deployment created", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
			err = deploymentAPI.WaitForDeployments([]string{metadata.Name}, metadata.Namespace, cfg.Config.TimeoutConfig)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
