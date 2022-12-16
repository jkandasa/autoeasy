package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	iostreamUtils "github.com/jkandasa/autoeasy/pkg/utils/iostream"
	deploymentAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/deployment"
	k8s "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func PortForward(restConfig *rest.Config, pfCfg openshiftTY.PortForwardRequest) (func(), error) {
	if restConfig == nil {
		return nil, errors.New("cluster rest config can not be empty")
	}

	if pfCfg.Namespace == "" {
		return nil, fmt.Errorf("namespace can not be empty. namespace:%s, pod:%s", pfCfg.Namespace, pfCfg.Pod)
	}

	if pfCfg.Pod == "" && pfCfg.Deployment == "" {
		return nil, errors.New("both pod and deployment can not be empty")
	}

	// get pod name
	if pfCfg.Pod == "" && pfCfg.Deployment != "" {
		k8sClient, err := k8s.NewClientFromRestConfig(restConfig)
		if err != nil {
			zap.L().Error("error on getting k8s client", zap.Error(err))
			return nil, err
		}
		pods, err := deploymentAPI.ListRunningPods(k8sClient, pfCfg.Deployment, pfCfg.Namespace)
		if err != nil {
			zap.L().Error("error on getting deployment", zap.Any("config", pfCfg), zap.Error(err))
			return nil, err
		}

		if len(pods) == 0 {
			return nil, fmt.Errorf("unable to get a running pod. deployment:%s, namespace:%s", pfCfg.Deployment, pfCfg.Namespace)
		}

		// update pod details
		_pod := pods[0]
		pfCfg.Pod = _pod.GetName()
		zap.L().Debug("selected a pod", zap.String("name", _pod.GetName()), zap.String("namespace", _pod.GetNamespace()))
	}

	// stopCh control the port forwarding lifecycle. When it gets closed the port forward will terminate
	stopCh := make(chan struct{}, 1)
	// readyCh communicate when the port forward is ready to get traffic
	readyCh := make(chan struct{})

	// load defaults
	if pfCfg.Streams == nil {
		pfCfg.Streams = iostreamUtils.GetLogWriter()
	}

	if len(pfCfg.Addresses) == 0 {
		pfCfg.Addresses = []string{"127.0.0.1"}
	}
	if len(pfCfg.Ports) == 0 {
		pfCfg.Addresses = []string{"8080:8080"} // localPort:targetPort
	}

	go func() {
		err := portForwardToPodOrDeployment(restConfig, pfCfg, stopCh, readyCh)
		if err != nil {
			zap.L().Error("error on port forward", zap.Any("config", pfCfg), zap.Error(err))
		}
	}()

	select {
	case <-readyCh:
		zap.L().Info("port forward ready", zap.Any("config", pfCfg))
		break

	case <-time.After(10 * time.Second):
		zap.L().Error("port forward reached timeout", zap.Any("config", pfCfg))
		return nil, errors.New("port forward reached timeout")
	}

	closeFunc := func() {
		close(stopCh)
	}

	return closeFunc, nil
}

func portForwardToPodOrDeployment(restCfg *rest.Config, pfConfig openshiftTY.PortForwardRequest, stopCh <-chan struct{}, readyCh chan struct{}) error {
	if pfConfig.Pod == "" {
		return errors.New("pod name can not be empty")
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", pfConfig.Namespace, pfConfig.Pod)
	hostIP := strings.TrimLeft(restCfg.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(restCfg)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.NewOnAddresses(dialer, pfConfig.Addresses, pfConfig.Ports, stopCh, readyCh, pfConfig.Streams.Out, pfConfig.Streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}
