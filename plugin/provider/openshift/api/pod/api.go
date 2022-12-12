package api

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/jkandasa/autoeasy/pkg/utils"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"

	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, opts []client.ListOption) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	err := k8sClient.List(context.Background(), podList, opts...)
	if err != nil {
		return nil, err
	}
	return podList, nil
}

func Get(k8sClient client.Client, name, namespace string) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(context.Background(), namespacedName, pod)
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func Delete(k8sClient client.Client, pod *corev1.Pod) error {
	return utils.IgnoreNotFoundError(k8sClient.Delete(context.Background(), pod))
}

func DeleteOfAll(k8sClient client.Client, pod *corev1.Pod, opts []client.DeleteAllOfOption) error {
	if pod == nil {
		pod = &corev1.Pod{}
	}
	return k8sClient.DeleteAllOf(context.Background(), pod, opts...)
}

func Create(k8sClient client.Client, pod *corev1.Pod) error {
	return k8sClient.Create(context.Background(), pod)
}

func CreateAndWait(k8sClient client.Client, pod *corev1.Pod) error {
	err := k8sClient.Create(context.Background(), pod)
	if err != nil {
		return err
	}

	executeFunc := func() (bool, error) {
		return isRunning(k8sClient, []string{pod.Name}, pod.Namespace)
	}
	return funcUtils.ExecuteWithDefaultTimeoutAndContinuesSuccessCount(executeFunc)
}

func CreateWithMap(k8sClient client.Client, cfg map[string]interface{}) error {
	pod := &corev1.Pod{}
	err := formatterUtils.JsonMapToStruct(cfg, pod)
	if err != nil {
		return err
	}
	return k8sClient.Create(context.Background(), pod)
}

// wait for pods
func WaitForPods(k8sClient client.Client, pods []string, namespace string, tc openshiftTY.TimeoutConfig) error {
	executeFunc := func() (bool, error) {
		return isRunning(k8sClient, pods, namespace)
	}
	return funcUtils.ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, tc.Timeout, tc.ScanInterval, tc.ExpectedSuccessCount)
}

func isRunning(k8sClient client.Client, pods []string, namespace string) (bool, error) {
	// check pods status
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}
	podList, err := List(k8sClient, opts)
	if err != nil {
		return false, err
	}
	notReadyList := []string{}
	for _, name := range pods {
		found := false
		for _, pod := range podList.Items {
			if pod.Namespace == namespace && pod.Name == name {
				found = true
				for _, _status := range pod.Status.ContainerStatuses {
					if !_status.Ready {
						notReadyList = append(notReadyList, fmt.Sprintf("%s:%s", pod.Name, _status.Name))
					}
				}
			}
		}
		if !found {
			notReadyList = append(notReadyList, name)
		}
	}

	if len(notReadyList) == 0 { // all pods are running
		zap.L().Debug("pods are running", zap.String("namespace", namespace), zap.Any("pods", pods))
		return true, nil
	}
	zap.L().Debug("waiting for pods into running state", zap.String("namespace", namespace), zap.Any("podsNotReady", notReadyList))
	return false, nil
}
