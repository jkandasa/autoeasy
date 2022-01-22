package cluster

import (
	nodeAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/node"
	"github.com/jkandasa/autoeasy/plugin/provider/openshift/store"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func PrintClusterInfo() {
	// get server version
	serverVersion, err := store.K8SClientSet.ServerVersion()
	if err != nil {
		zap.L().Error("error on getting server version", zap.Error(err))
	} else {
		serAddr := store.K8SClientSet.RESTClient().Get().URL()
		zap.L().Info("server version",
			zap.Any("server", serAddr.String()),
			zap.String("buildDate", serverVersion.BuildDate),
			zap.String("compiler", serverVersion.Compiler),
			zap.String("gitCommit", serverVersion.GitCommit),
			zap.String("gitTreeState", serverVersion.GitTreeState),
			zap.String("gitVersion", serverVersion.GitVersion),
			zap.String("goVersion", serverVersion.GoVersion),
			zap.String("major", serverVersion.Major),
			zap.String("minor", serverVersion.Minor),
			zap.String("platform", serverVersion.Platform),
			zap.String("kubeVersion", serverVersion.String()),
		)
	}

	opts := []client.ListOption{}
	nodesList, err := nodeAPI.List(opts)
	nodes := make([]openshiftTY.Node, 0)

	if err != nil {
		zap.L().Error("error on getting node details", zap.Error(err))
	} else {
		masterCount := 0
		workerCount := 0
		for _, node := range nodesList.Items {
			// update node type
			isMaster := false
			if _, found := node.Labels["node-role.kubernetes.io/worker"]; found {
				workerCount++
				isMaster = false
			} else if _, found = node.Labels["node-role.kubernetes.io/master"]; found {
				masterCount++
				isMaster = true
			}

			nodes = append(nodes, openshiftTY.Node{
				Name:     node.Name,
				Labels:   node.Labels,
				IsMaster: isMaster,
				Capacity: node.Status.Capacity,
				NodeInfo: node.Status.NodeInfo,
			})
		}
		zap.L().Info("cluster details", zap.Int("numberOfNodes", len(nodesList.Items)), zap.Int("numberOfMaster", masterCount), zap.Int("numberOfWorker", workerCount))

		// print node details
		for _, node := range nodes {
			zap.L().Debug("node details", zap.Any("node", node))
		}

		// TODO: get OCP version
	}
}
