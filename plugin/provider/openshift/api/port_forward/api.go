package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	iostreamUtils "github.com/jkandasa/autoeasy/pkg/utils/iostream"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func PortForward(restConfig *rest.Config, portForwardConfig openshiftTY.PortForwardRequest) (func(), error) {
	if restConfig == nil {
		return nil, errors.New("cluster rest config can not be empty")
	}

	if portForwardConfig.Namespace == "" || portForwardConfig.Pod == "" {
		return nil, fmt.Errorf("namespace or pod can not be empty. namespace:%s, pod:%s", portForwardConfig.Namespace, portForwardConfig.Pod)
	}

	// stopCh control the port forwarding lifecycle. When it gets closed the port forward will terminate
	stopCh := make(chan struct{}, 1)
	// readyCh communicate when the port forward is ready to get traffic
	readyCh := make(chan struct{})

	// load defaults
	if portForwardConfig.Streams == nil {
		portForwardConfig.Streams = iostreamUtils.GetLogWriter()
	}

	if len(portForwardConfig.Addresses) == 0 {
		portForwardConfig.Addresses = []string{"127.0.0.1"}
	}
	if len(portForwardConfig.Ports) == 0 {
		portForwardConfig.Addresses = []string{"8080:8080"} // localPort:targetPort
	}

	go func() {
		err := portForwardToPod(restConfig, portForwardConfig, stopCh, readyCh)
		if err != nil {
			zap.L().Error("error on port forward", zap.Any("config", portForwardConfig), zap.Error(err))
		}
	}()

	select {
	case <-readyCh:
		zap.L().Info("port forward ready", zap.Any("config", portForwardConfig))
		break

	case <-time.After(10 * time.Second):
		zap.L().Error("port forward reached timeout", zap.Any("config", portForwardConfig))
		return nil, errors.New("port forward reached timeout")
	}

	closeFunc := func() {
		close(stopCh)
	}

	return closeFunc, nil
}

func portForwardToPod(restCfg *rest.Config, pfConfig openshiftTY.PortForwardRequest, stopCh <-chan struct{}, readyCh chan struct{}) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		pfConfig.Namespace, pfConfig.Pod)
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
