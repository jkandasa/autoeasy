package registry

import (
	nsAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/namespace"
	podAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/pod"
	portForwardAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/port_forward"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deployIndexImage(address string) (func(), string, error) {
	if address != "" {
		return nil, address, nil
	}

	// deploy index image
	k8sClientCfg := openshiftClient.GetK8SClientConfig()
	k8sClient, err := k8sClientCfg.NewClient()
	if err != nil {
		zap.L().Fatal("error on loading k8s client", zap.Error(err))
	}

	ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: IndexImageNamespace}}

	// delete old namespace
	err = nsAPI.DeleteAndWait(k8sClient, &ns)
	if err != nil {
		zap.L().Fatal("error on deleting namespace", zap.String("name", ns.Name), zap.Error(err))
	}

	// create namespace
	err = nsAPI.Create(k8sClient, &ns)
	if err != nil {
		zap.L().Fatal("error on creating namespace", zap.String("name", ns.Name), zap.Error(err))
	}

	indexImagePod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: IndexImagePodName, Namespace: ns.Name, Labels: map[string]string{"app": IndexImagePodName}},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  IndexImagePodName,
					Image: indexImage,
					Ports: []corev1.ContainerPort{
						{
							Name:          IndexImagePodName,
							ContainerPort: IndexImageContainerPort,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{Exec: &corev1.ExecAction{
							Command: []string{"grpc_health_probe", "-addr=:50051"},
						}},
					},
				},
			},
		},
	}

	// create pod and wait
	err = podAPI.CreateAndWait(k8sClient, &indexImagePod)
	if err != nil {
		zap.L().Fatal("error on creating pod", zap.String("name", indexImagePod.Name), zap.String("namespace", indexImagePod.Namespace), zap.Error(err))
	}

	// setup port-forward
	portForwardCfg := openshiftTY.PortForwardRequest{
		Namespace: IndexImageNamespace,
		Pod:       IndexImagePodName,
		Addresses: []string{"127.0.0.1"},
		Ports:     []string{"50051:50051"},
	}
	closeFunc, err := portForwardAPI.PortForward(k8sClientCfg.GetRestConfig(), portForwardCfg)
	if err != nil {
		return nil, "", err
	}

	return closeFunc, "127.0.0.1:50051", nil
}