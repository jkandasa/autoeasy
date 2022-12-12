package k8s

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	iostreamTY "github.com/jkandasa/autoeasy/pkg/types/iostream"
	iostreamUtils "github.com/jkandasa/autoeasy/pkg/utils/iostream"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardRequest struct {
	Namespace string                `json:"namespace"`
	Pod       string                `json:"pod"`
	Addresses []string              `json:"addresses"`
	Ports     []string              `json:"ports"`
	Streams   *iostreamTY.IOStreams `json:"-"`
	stopCh    <-chan struct{}       `json:"-"`
	readyCh   chan struct{}         `json:"-"`
}

func (k8s *K8SClient) PortForward(req PortForwardRequest) (func(), error) {
	if k8s.restConfig == nil {
		return nil, errors.New("seems not logged into the cluster")
	}

	if req.Namespace == "" || req.Pod == "" {
		return nil, fmt.Errorf("namespace or pod can not be empty. namespace:%s, pod:%s", req.Namespace, req.Pod)
	}

	// stopCh control the port forwarding lifecycle. When it gets closed the port forward will terminate
	stopCh := make(chan struct{}, 1)
	// readyCh communicate when the port forward is ready to get traffic
	readyCh := make(chan struct{})

	req.stopCh = stopCh
	req.readyCh = readyCh

	// load defaults
	if req.Streams == nil {
		req.Streams = iostreamUtils.GetLogWriter()
	}

	if len(req.Addresses) == 0 {
		req.Addresses = []string{"127.0.0.1"}
	}
	if len(req.Ports) == 0 {
		req.Addresses = []string{"8080:8080"} // localPort:targetPort
	}

	go func() {
		err := portForwardToPod(req, k8s.restConfig)
		if err != nil {
			zap.L().Error("error on port forward", zap.Any("config", req), zap.Error(err))
		}
	}()

	select {
	case <-readyCh:
		zap.L().Info("port forward ready", zap.Any("config", req))
		break

	case <-time.After(10 * time.Second):
		zap.L().Error("port forward reached timeout", zap.Any("config", req))
		return nil, errors.New("port forward reached timeout")
	}

	closeFunc := func() {
		close(stopCh)
	}

	return closeFunc, nil
}

func portForwardToPod(req PortForwardRequest, restCfg *rest.Config) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		req.Namespace, req.Pod)
	hostIP := strings.TrimLeft(restCfg.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(restCfg)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.NewOnAddresses(dialer, req.Addresses, req.Ports, req.stopCh, req.readyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}
