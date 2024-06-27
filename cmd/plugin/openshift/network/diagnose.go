package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	diagnoseRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/diagnose"
	openshiftRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"
	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	namespaceAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/namespace"
	nodeAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/node"
	podAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/pod"
	portForwardAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/port_forward"
	routeAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/route"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	osroutev1 "github.com/openshift/api/route/v1"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PodAccessModePortForward = "port-forward"
	PodAccessModeRoute       = "route"
)

type diagnoseNetwork struct {
	image             string
	namespace         string
	alwaysPullImage   bool
	podAccessMode     string
	measureThroughput bool
	measureLatency    bool
	quickRun          bool
	k8sClient         client.Client
	k8sRestConfig     *rest.Config
}

var (
	diagnoseImage     string
	diagnoseNamespace string
	alwaysPullImage   bool
	podAccessMode     string
	measureThroughput bool
	measureLatency    bool
	quickRun          bool
)

func init() {
	diagnoseRootCmd.AddCommand(diagnoseNetworkCmd)
	diagnoseNetworkCmd.Flags().StringVar(&diagnoseImage, "image", openshiftRootCmd.DefaultDiagnoseImage, "diagnose container image, will be used on the pod")
	diagnoseNetworkCmd.Flags().BoolVar(&alwaysPullImage, "always-pull-image", true, "always pulls the image, while setting the diagnose pods")
	diagnoseNetworkCmd.Flags().StringVar(&diagnoseNamespace, "namespace", openshiftRootCmd.DefaultDiagnoseNamespace, "name of the namespace to be used to deploy resources required for diagnose tests")
	diagnoseNetworkCmd.Flags().StringVar(&podAccessMode, "pod-access-mode", "port-forward", "how to access the remote pod REST API, options: port-forward, route")
	diagnoseNetworkCmd.Flags().BoolVar(&measureThroughput, "throughput", false, "measures network throughput")
	diagnoseNetworkCmd.Flags().BoolVar(&measureLatency, "latency", true, "measures network latency")
	diagnoseNetworkCmd.Flags().BoolVar(&quickRun, "quick-run", true, "performs quick run on a single node, if false performs the tests on all the nodes")
}

// example:
// autoeasy openshift diagnose network --quick-run
// implementation steps:
// get node details
// setup iperf3/ping container(pods) on each nodes
// get pod IPs, node1-pod: 10.0.0.21, node2-pod: 10.1.0.22
// check the latency and throughput
//   - node1 to node2 (access node1 via port forward)
//   - node2 to node1 (access node2 via port forward)
// print all the details in a readable format

var diagnoseNetworkCmd = &cobra.Command{
	Use:   "network",
	Short: "Performs network latency and throughput tests on different node and reports the status",
	Example: `  # simple
  autoeasy openshift diagnose network --latency --throughput --quick-run

  # detailed run, performs the tests on all the nodes
  autoeasy openshift diagnose network --latency --throughput --quick-run=false
  
  # runs only throughput test
  autoeasy openshift diagnose network --latency=false --throughput

  # runs only latency test
  autoeasy openshift diagnose network --latency --throughput=false

  # with debug logs
  autoeasy openshift diagnose network --latency --throughput --log-level=debug`,
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()

		// get kubernetes client
		k8sClientCfg := openshiftClient.GetK8SClientConfig()
		k8sClient, err := k8sClientCfg.NewClient()
		if err != nil {
			zap.L().Error("error on setup k8s client", zap.Error(err))
			rootCmd.ExitWithError()
		}

		dn := diagnoseNetwork{
			image:             diagnoseImage,
			namespace:         diagnoseNamespace,
			alwaysPullImage:   alwaysPullImage,
			measureThroughput: measureThroughput,
			measureLatency:    measureLatency,
			quickRun:          quickRun,
			k8sClient:         k8sClient,
			k8sRestConfig:     k8sClientCfg.GetRestConfig(),
		}
		if !dn.measureLatency && !dn.measureThroughput {
			zap.L().Error("'latency'and 'throughput', both can not be disabled")
			rootCmd.ExitWithError()
		}

		// update pod access mode
		dn.podAccessMode = podAccessMode
		if dn.podAccessMode == "" {
			dn.podAccessMode = PodAccessModePortForward
		}

		defer func() {
			dn.cleanup()
			fmt.Fprintf(os.Stdout, "Overall time taken:%s\nDone\n", time.Since(startTime))
		}()

		nodesDiagnoseData, err := dn.executeDiagnose()
		if err != nil {
			zap.L().Error("error on setup pod", zap.Error(err))
			rootCmd.ExitWithError()
		}

		zap.L().Debug("diagnose data", zap.Any("data", nodesDiagnoseData))

		dn.printLatency(nodesDiagnoseData)

		dn.printThroughput(nodesDiagnoseData)

	},
}

// resource cleanup actions
func (dn *diagnoseNetwork) cleanup() {
	fmt.Fprintf(os.Stdout, "\nCleanup: removing resources\n")

	// delete pods
	err := namespaceAPI.DeleteAndWait(dn.k8sClient, &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: dn.namespace,
		},
	})
	if err != nil {
		zap.L().Error("error on running cleanup", zap.Error(err))
		rootCmd.ExitWithError()
	}
}

func (dn *diagnoseNetwork) executeDiagnose() ([]DiagnoseNodeData, error) {
	fmt.Fprintf(os.Stdout, "\nCollecting nodes detail\n\n")
	// get node details
	nodeList, err := nodeAPI.List(dn.k8sClient, []client.ListOption{})
	if err != nil {
		zap.L().Error("error on getting nodes list", zap.Error(err))
		return nil, err
	}

	nodes := nodeList.Items

	// create namespace
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: dn.namespace,
		},
	}
	err = namespaceAPI.Create(dn.k8sClient, ns)
	if err != nil {
		zap.L().Error("error on creating a namespace", zap.String("namespace", ns.Name), zap.Error(err))
		return nil, err
	}

	// collect nodes
	nodesData := []DiagnoseNodeData{}
	for index, node := range nodes {
		nodeData := DiagnoseNodeData{
			Index:          index,
			NodeName:       node.Name,
			NodeNameSimple: dn.getSimpleNodeName(node.Name),
			Roles:          []string{},
		}

		// update roles
		// role labels
		// node-role.kubernetes.io/worker: ""
		// node-role.kubernetes.io/control-plane: ""
		// node-role.kubernetes.io/master: ""

		nodeLabels := node.Labels
		if nodeLabels == nil {
			nodeLabels = map[string]string{}
		}

		// is worker
		if _, ok := nodeLabels["node-role.kubernetes.io/worker"]; ok {
			nodeData.Roles = append(nodeData.Roles, "worker")
		}

		// is control-plane
		if _, ok := nodeLabels["node-role.kubernetes.io/control-plane"]; ok {
			nodeData.Roles = append(nodeData.Roles, "control-plane")
		}

		// is master
		if _, ok := nodeLabels["node-role.kubernetes.io/master"]; ok {
			nodeData.Roles = append(nodeData.Roles, "master")
		}

		zap.L().Debug("node details", zap.String("name", nodeData.NodeName), zap.Strings("roles", nodeData.Roles))

		nodesData = append(nodesData, nodeData)
	}

	// print nodes
	headers := []printer.Header{
		{Title: "nodes", ValuePath: "nodeName"},
		{Title: "roles", ValuePath: "roles"},
	}

	rows := make([]interface{}, len(nodesData))
	for index, nodeData := range nodesData {
		rows[index] = nodeData
	}

	printer.Print(os.Stdout, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)

	if len(nodeList.Items) == 1 {
		return nil, fmt.Errorf("diagnose tests can not be executed on a single node cluster")
	}

	// setup pods
	podNames := []string{}
	// deploy pods
	fmt.Fprintf(os.Stdout, "\nCreating pods\n")
	for _, node := range nodesData {
		// create pods
		pod := dn.getPodTemplate(node.NodeName, node.NodeNameSimple)
		err := podAPI.Create(dn.k8sClient, pod)
		if err != nil {
			zap.L().Error("error on creating a pod", zap.String("node", node.NodeName), zap.Error(err))
			return nodesData, err
		}
		nodesData[node.Index].PodName = pod.Name
		podNames = append(podNames, pod.Name)
	}

	// wait for pods to be up and running
	fmt.Fprintf(os.Stdout, "Waiting for the pods availability\n")
	err = podAPI.WaitForPods(dn.k8sClient, podNames, ns.Name, openshiftTY.TimeoutConfig{
		ScanInterval:         time.Second * 2,
		Timeout:              time.Second * 45,
		ExpectedSuccessCount: 1,
	})
	if err != nil {
		zap.L().Error("error on waiting pods to be in running state", zap.Strings("pods", podNames), zap.Error(err))
		return nodesData, err
	}

	// list all the pods
	podList, err := podAPI.List(dn.k8sClient, []client.ListOption{client.InNamespace(ns.Name)})
	if err != nil {
		zap.L().Error("error on listing pods", zap.Error(err))
		return nodesData, err
	}

	// update pod, host ip
	for _, pod := range podList.Items {
		for _, node := range nodesData {
			if node.PodName == pod.Name {
				nodesData[node.Index].PodIP = pod.Status.PodIP
				nodesData[node.Index].HostIP = pod.Status.HostIP
			}
		}
	}

	if dn.podAccessMode == PodAccessModePortForward {
		fmt.Fprintf(os.Stdout, "Enabling Port Forward to access remote pods\n\n")
		err := dn.setupPortForward(nodesData)
		if err != nil {
			return nodesData, err
		}
	} else if dn.podAccessMode == PodAccessModeRoute {
		fmt.Fprintf(os.Stdout, "Creating routes to access remote pods\n\n")
		err := dn.setupRoute(nodesData)
		if err != nil {
			return nodesData, err
		}
	} else {
		zap.L().Error("unsupported pod access mode", zap.String("pod-access-mode", string(dn.podAccessMode)))
		return nodesData, nil
	}

	// execute diagnose on remote pod
	err = dn.executeDiagnoseOnRemote(nodesData)
	return nodesData, err
}

func (dn *diagnoseNetwork) executeDiagnoseOnRemote(nodesData []DiagnoseNodeData) error {
	queryParameters := []string{
		fmt.Sprintf("iperf3_enabled=%v", dn.measureThroughput),
		"iperf3_options=--json,--time=5",
		"ping_enabled=true",
		"ping_count=4",
		"ping_interval=1s",
	}

	for index, node := range nodesData {
		if dn.quickRun && index > 0 {
			break
		}
		targetHosts := strings.Join(dn.getTargetHosts(node.Index, nodesData), ",")
		// do curl
		fmt.Fprintf(os.Stdout, "Performing network tests on the node: %s ", node.NodeName)
		startTime := time.Now()
		targetUrl := fmt.Sprintf("%s/api/diagnose/network?hosts=%s&%s", node.PodHttpUrl, targetHosts, strings.Join(queryParameters, "&"))
		zap.L().Debug("running diagnose tests", zap.String("node", node.NodeName), zap.String("targetIPs", targetHosts), zap.String("remoteUrl", targetUrl))
		resp, err := http.Get(targetUrl)
		if err != nil {
			zap.L().Error("error on getting response", zap.Error(err))
			return err
		}
		fmt.Fprintf(os.Stdout, "[timeTaken: %s]\n", time.Since(startTime).String())
		if resp.StatusCode != http.StatusOK {
			zap.L().Error("invalid status response received", zap.Int("statusCode", resp.StatusCode), zap.String("status", resp.Status), zap.String("targetUrl", targetUrl))
			return err
		}

		defer resp.Body.Close()
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			zap.L().Error("error on getting body", zap.Error(err))
			return err
		}

		diagnoseResponse := []NetworkDiagnoseResponse{}

		err = json.Unmarshal(responseData, &diagnoseResponse)
		if err != nil {
			zap.L().Error("error on unmarshal diagnose data", zap.Error(err))
			return err
		}
		nodesData[node.Index].DiagnoseResponse = diagnoseResponse

		// close the port forward
		if nodesData[node.Index].PortForwardCloseFunc != nil {
			nodesData[node.Index].PortForwardCloseFunc()
			nodesData[node.Index].PortForwardCloseFunc = nil
		}
	}
	return nil
}

func (dn *diagnoseNetwork) getSimpleNodeName(nodeName string) string {
	return strings.SplitN(nodeName, ".", 2)[0]
}

func (dn *diagnoseNetwork) getTargetHosts(sourceIndex int, nodesData []DiagnoseNodeData) []string {
	targetHosts := []string{}
	for _, node := range nodesData {
		if node.Index != sourceIndex {
			targetHosts = append(targetHosts, node.PodIP)
		}
	}
	return targetHosts
}

// pod ip, node name mapping
func (dn *diagnoseNetwork) getPodIPNodeMap(nodesData []DiagnoseNodeData) map[string]string {
	nodeMap := map[string]string{}
	for _, node := range nodesData {
		nodeMap[node.PodIP] = node.NodeNameSimple
	}
	return nodeMap
}

// setup port forward to access pods REST API
func (dn *diagnoseNetwork) setupPortForward(nodesData []DiagnoseNodeData) error {
	for index, node := range nodesData {
		localHostPort := uint32(42000 + index)
		portForwardCfg := openshiftTY.PortForwardRequest{
			Namespace: dn.namespace,
			Pod:       node.PodName,
			Addresses: []string{"127.0.0.1"},
			Ports:     []string{fmt.Sprintf("%d:8080", localHostPort)},
		}
		zap.L().Debug("starting port forward", zap.Any("node", node))
		closeFunc, err := portForwardAPI.PortForward(dn.k8sRestConfig, portForwardCfg)
		if err != nil {
			zap.L().Error("error on setup port forward", zap.Any("node", node), zap.Error(err))
			return err
		}
		nodesData[index].PortForwardCloseFunc = closeFunc
		nodesData[index].LocalHostPort = localHostPort
		nodesData[index].PodHttpUrl = fmt.Sprintf("http://127.0.0.1:%d", localHostPort)
	}
	return nil
}

// setup route to access pods REST API
func (dn *diagnoseNetwork) setupRoute(nodesData []DiagnoseNodeData) error {
	for _, node := range nodesData {
		zap.L().Debug("creating route", zap.Any("node", node))

		// create services
		service := dn.getServiceTemplate(node.NodeNameSimple)
		err := dn.k8sClient.Create(context.Background(), service)
		if err != nil {
			zap.L().Error("error on creating a service", zap.String("node", node.NodeName), zap.Error(err))
			return err
		}

		route := dn.getRouteTemplate(node.NodeNameSimple)
		err = routeAPI.Create(dn.k8sClient, route)
		if err != nil {
			zap.L().Error("error on creating route", zap.Any("node", node), zap.Error(err))
			return err
		}
		updateRoute, err := routeAPI.Get(dn.k8sClient, route.Name, route.Namespace)
		if err != nil {
			zap.L().Error("error on getting a route", zap.Any("node", node), zap.Error(err))
			return err
		}
		if len(updateRoute.Status.Ingress) > 0 {
			nodesData[node.Index].PodHttpUrl = fmt.Sprintf("http://%s", updateRoute.Status.Ingress[0].Host)
		}
	}
	return nil
}

// templates

// pod template
func (dn *diagnoseNetwork) getPodTemplate(actualNodeName, nodeName string) *corev1.Pod {
	imagePullPolicy := corev1.PullIfNotPresent
	if dn.alwaysPullImage {
		imagePullPolicy = corev1.PullAlways
	}
	pod := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      nodeName,
			Namespace: dn.namespace,
			Labels: map[string]string{
				"app":       "network-diagnose",
				"node-name": nodeName,
				"tool":      "autoeasy",
			},
		},
		Spec: corev1.PodSpec{
			NodeName: actualNodeName,
			Containers: []corev1.Container{
				{
					Name:            "diagnose-nw",
					Image:           dn.image,
					ImagePullPolicy: imagePullPolicy,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							Protocol:      "TCP",
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}

	return pod
}

// service template
func (dn *diagnoseNetwork) getServiceTemplate(nodeName string) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      nodeName,
			Namespace: dn.namespace,
			Labels: map[string]string{
				"app":       "network-diagnose",
				"node-name": nodeName,
				"tool":      "autoeasy",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"node-name": nodeName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}

	return service
}

// route template
func (dn *diagnoseNetwork) getRouteTemplate(nodeName string) *osroutev1.Route {
	route := &osroutev1.Route{
		ObjectMeta: v1.ObjectMeta{
			Name:      nodeName,
			Namespace: dn.namespace,
			Annotations: map[string]string{
				"haproxy.router.openshift.io/timeout": "5m",
			},
			Labels: map[string]string{
				"tool": "autoeasy",
			},
		},
		Spec: osroutev1.RouteSpec{
			Subdomain: nodeName,
			To: osroutev1.RouteTargetReference{
				Kind: "Service",
				Name: nodeName,
			},
			Port: &osroutev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
		},
	}

	return route
}

// print the results

// prints the network latency test results
func (dn *diagnoseNetwork) printLatency(nodesData []DiagnoseNodeData) {
	if !dn.measureLatency {
		return
	}

	podNodeMap := dn.getPodIPNodeMap(nodesData)
	nodeNames := []string{}
	for _, name := range podNodeMap {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	type LatencyData struct {
		Node   string
		AvgRtt time.Duration
	}

	pingDataAll := map[string]map[string]LatencyData{}
	for _, diagnoseData := range nodesData {
		sourceNodeName := diagnoseData.NodeNameSimple
		targetPingData := map[string]LatencyData{}
		for _, diagnoseResp := range diagnoseData.DiagnoseResponse {
			if diagnoseResp.Ping == nil {
				continue
			}
			pingTargetNode := dn.getSimpleNodeName(podNodeMap[diagnoseResp.Ping.Hostname])

			pingResponse := diagnoseResp.Ping.Response
			if !diagnoseResp.Ping.IsSuccess {
				zap.L().Error("error on latency test", zap.String("node", diagnoseData.NodeName), zap.String("error", diagnoseResp.Ping.ErrorMessage))
				continue
			}
			pingStatistics := PingStatistics{}
			err := json.Unmarshal([]byte(pingResponse), &pingStatistics)
			if err != nil {
				zap.L().Error("error on converting json string to PingResponse", zap.String("data", pingResponse), zap.Error(err))
				continue
			}
			targetPingData[pingTargetNode] = LatencyData{
				Node:   pingTargetNode,
				AvgRtt: pingStatistics.AvgRtt,
			}

		}
		pingDataAll[sourceNodeName] = targetPingData
	}

	headers := []printer.Header{
		{Title: "node", ValuePath: "sourceNode"},
	}

	type RowData struct {
		SourceNode string      `json:"sourceNode"`
		Data       interface{} `json:"data"`
	}

	for _, nodeName := range nodeNames {
		selectedNode := fmt.Sprintf("data.%s-avg-rtt", nodeName)
		headers = append(headers, printer.Header{
			Title:     nodeName,
			ValuePath: selectedNode,
		})
	}

	rows := []interface{}{}
	for index, nodeName := range nodeNames {
		if dn.quickRun && index > 0 {
			break
		}

		data := map[string]interface{}{}
		diagnoseData := pingDataAll[nodeName]
		for _, targetNode := range nodeNames {
			if targetNode == nodeName {
				continue
			}
			avgRttValue := "-"
			if diagnoseData[targetNode].AvgRtt != 0 {
				avgRttValue = diagnoseData[targetNode].AvgRtt.String()
			}
			data[fmt.Sprintf("%s-avg-rtt", targetNode)] = avgRttValue
		}
		row := RowData{
			SourceNode: strings.ToUpper(nodeName),
			Data:       data,
		}
		rows = append(rows, row)
	}

	zap.L().Debug("row data", zap.Any("data", rows))

	fmt.Fprintf(os.Stdout, "\nNETWORK LATENCY(AVG RTT)\n")
	printer.Print(os.Stdout, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
}

// prints the network throughput test results
func (dn *diagnoseNetwork) printThroughput(nodesData []DiagnoseNodeData) {
	if !dn.measureThroughput {
		return
	}

	podNodeMap := dn.getPodIPNodeMap(nodesData)
	nodeNames := []string{}
	for _, name := range podNodeMap {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	type ThroughputData struct {
		Node            string
		BitsPerSecondTx float64
		BitsPerSecondRx float64
		BitsPerSecond   string
	}

	throughputDataAll := map[string]map[string]ThroughputData{}
	for _, diagnoseData := range nodesData {
		sourceNodeName := diagnoseData.NodeNameSimple
		targetIPerf3Data := map[string]ThroughputData{}
		for _, diagnoseResp := range diagnoseData.DiagnoseResponse {
			if diagnoseResp.IPerf3 == nil {
				continue
			}
			iperf3TargetNode := podNodeMap[diagnoseResp.IPerf3.Hostname]

			if !diagnoseResp.IPerf3.IsSuccess {
				zap.L().Error("error on throughput test", zap.String("node", diagnoseData.NodeName), zap.String("error", diagnoseResp.IPerf3.ErrorMessage))
				continue
			}

			iperf3Response := diagnoseResp.IPerf3.Response
			statistics := IPerf3Statistics{}
			err := json.Unmarshal([]byte(iperf3Response), &statistics)
			if err != nil {
				zap.L().Error("error on converting json string to IPerf3 response", zap.String("data", iperf3Response), zap.Error(err))
				continue
			}

			if len(statistics.End.Streams) == 0 {
				continue
			}

			stream := statistics.End.Streams[0]

			targetIPerf3Data[iperf3TargetNode] = ThroughputData{
				Node:            iperf3TargetNode,
				BitsPerSecondTx: stream.Sender.BitsPerSecond,
				BitsPerSecondRx: stream.Receiver.BitsPerSecond,
				BitsPerSecond:   fmt.Sprintf("%.2f/%.2f", stream.Sender.BitsPerSecond/1000000000, stream.Receiver.BitsPerSecond/1000000000),
			}

		}
		throughputDataAll[sourceNodeName] = targetIPerf3Data
	}

	headers := []printer.Header{
		{Title: "node", ValuePath: "sourceNode"},
	}

	type RowData struct {
		SourceNode string      `json:"sourceNode"`
		Data       interface{} `json:"data"`
	}

	rows := []interface{}{}

	for _, nodeName := range nodeNames {
		selectedNode := fmt.Sprintf("data.%s-bits-per-second", nodeName)
		headers = append(headers, printer.Header{
			Title:     nodeName,
			ValuePath: selectedNode,
		})
	}

	for index, nodeName := range nodeNames {
		if dn.quickRun && index > 0 {
			break
		}

		data := map[string]interface{}{}
		diagnoseData := throughputDataAll[nodeName]
		for _, targetNode := range nodeNames {
			if targetNode == nodeName {
				continue
			}
			data[fmt.Sprintf("%s-bits-per-second", targetNode)] = diagnoseData[targetNode].BitsPerSecond
		}
		row := RowData{
			SourceNode: strings.ToUpper(nodeName),
			Data:       data,
		}
		rows = append(rows, row)
	}

	zap.L().Debug("row data", zap.Any("data", rows))

	fmt.Fprintf(os.Stdout, "\nNETWORK THROUGHPUT (Gbit/s: Tx/Rx)\n")
	printer.Print(os.Stdout, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
}
