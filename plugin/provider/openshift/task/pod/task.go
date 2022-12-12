package task

import (
	"fmt"

	"github.com/jkandasa/autoeasy/pkg/utils"
	podAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/pod"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Run(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) (interface{}, error) {
	switch cfg.Function {
	case openshiftTY.FuncAdd:
		if len(cfg.Data) == 0 {
			return nil, fmt.Errorf("no data supplied. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
		}
		return nil, add(k8sClient, cfg)

	case openshiftTY.FuncKeepOnly, openshiftTY.FuncRemove:
		if len(cfg.Data) == 0 {
			return nil, fmt.Errorf("no data supplied. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
		}
		fallthrough
	case openshiftTY.FuncRemoveAll:
		return nil, performDelete(k8sClient, cfg)

	case openshiftTY.FuncWaitForReady:
		if len(cfg.Data) == 0 {
			return nil, fmt.Errorf("no data supplied. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
		}
		return nil, waitForReady(k8sClient, cfg)

	}

	return nil, fmt.Errorf("unknown function. {kind:%s, function:%s}", cfg.Kind, cfg.Function)
}

func waitForReady(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	// get pods detail
	suppliedItems := utils.ToNamespacedNameSlice(cfg.Data)

	// group by namespace
	items := map[string][]string{}
	for _, pod := range suppliedItems {
		pods, ok := items[pod.Namespace]
		if !ok {
			pods = make([]string, 0)
		}
		pods = append(pods, pod.Name)
		items[pod.Namespace] = pods
	}

	// verify status
	for namespace, pods := range items {
		err := podAPI.WaitForPods(k8sClient, pods, namespace, cfg.Config.TimeoutConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func performDelete(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	podList, err := podAPI.List(k8sClient, opts)
	if err != nil {
		zap.L().Fatal("error on getting pod list", zap.Error(err))
	}

	if cfg.Function == openshiftTY.FuncRemoveAll {
		return delete(k8sClient, cfg, podList.Items)
	} else if cfg.Function == openshiftTY.FuncRemove || cfg.Function == openshiftTY.FuncKeepOnly {
		deletionList := make([]corev1.Pod, 0)

		suppliedItems := utils.ToNamespacedNameSlice(cfg.Data)

		isRemove := cfg.Function == openshiftTY.FuncRemove

		for _, pod := range podList.Items {
			if isRemove { // remove
				if utils.ContainsNamespacedName(suppliedItems, pod.ObjectMeta) {
					deletionList = append(deletionList, pod)
				}
			} else { // keep only
				if utils.ContainsNamespacedName(suppliedItems, pod.ObjectMeta) {
					deletionList = append(deletionList, pod)
				}
			}
		}

		return delete(k8sClient, cfg, deletionList)
	}
	return nil

}

func delete(k8sClient client.Client, cfg *openshiftTY.ProviderConfig, items []corev1.Pod) error {
	if len(items) == 0 {
		return nil
	}
	for _, pod := range items {
		err := podAPI.Delete(k8sClient, &pod)
		if err != nil {
			return err
		}
		zap.L().Debug("deleted a pod", zap.String("name", pod.Name), zap.String("namespace", pod.Name))
	}
	return nil
}

func add(k8sClient client.Client, cfg *openshiftTY.ProviderConfig) error {
	for _, cfgRaw := range cfg.Data {
		podCfg, ok := cfgRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := []client.ListOption{
			client.InNamespace(""),
		}
		podList, err := podAPI.List(k8sClient, opts)
		if err != nil {
			zap.L().Fatal("error on getting pod list", zap.Error(err))
		}

		metadata, err := utils.GetObjectMeta(podCfg)
		if err != nil {
			zap.L().Fatal("error on getting object meta", zap.Any("metadata", metadata), zap.Error(err))
		}
		found := false
		for _, pod := range podList.Items {
			if pod.Name == metadata.Name && pod.Namespace == metadata.Namespace {
				zap.L().Debug("pod exists", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
				found = true
				if cfg.Config.Recreate {
					zap.L().Debug("pod recreate enabled", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
					err = podAPI.Delete(k8sClient, &pod)
					if err != nil {
						return err
					}
					found = false
				}
				break
			}
		}
		if !found {
			err = podAPI.CreateWithMap(k8sClient, podCfg)
			if err != nil {
				zap.L().Fatal("error on creating pod", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace), zap.Error(err))
			}
			zap.L().Info("pod created", zap.String("name", metadata.Name), zap.String("namespace", metadata.Namespace))
			err = podAPI.WaitForPods(k8sClient, []string{metadata.Name}, metadata.Namespace, cfg.Config.TimeoutConfig)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
