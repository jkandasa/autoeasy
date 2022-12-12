package k8s

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardRequest struct {
	Pod       v1.Pod          `json:"pod"`
	Addresses []string        `json:"addresses"`
	Ports     []string        `json:"ports"`
	Streams   *IOStreams      `json:"-"`
	StopCh    <-chan struct{} `json:"-"`
	ReadyCh   chan struct{}   `json:"-"`
}

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

func (k8s *K8SClient) PortForward(req PortForwardRequest) (func(), error) {
	if k8s.restConfig == nil {
		return nil, errors.New("seems not logged in")
	}

	// stopCh control the port forwarding lifecycle. When it gets closed the port forward will terminate
	stopCh := make(chan struct{}, 1)
	// readyCh communicate when the port forward is ready to get traffic
	readyCh := make(chan struct{})

	// load defaults
	if req.Streams == nil {
		req.Streams = &IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		}
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
		zap.L().Info("port forward ready")
		break

	case <-time.After(10 * time.Second):
		zap.L().Error("reached timeout")
		return nil, errors.New("reached timeout")
	}

	closeFunc := func() {
		close(stopCh)
	}

	return closeFunc, nil

}

func portForwardToPod(req PortForwardRequest, restCfg *rest.Config) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		req.Pod.Namespace, req.Pod.Name)
	hostIP := strings.TrimLeft(restCfg.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(restCfg)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.NewOnAddresses(dialer, req.Addresses, req.Ports, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}
