package icsp

import (
	openshiftDeleteCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/delete"
	icspAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/image_content_source_policy"
	nodeAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/node"

	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	"github.com/openshift/api/operator/v1alpha1"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	deleteAll bool
)

func init() {
	openshiftDeleteCmd.AddCommand(deleteIcspCmd)
	deleteIcspCmd.Flags().BoolVar(&deleteAll, "delete-all", false, "deletes all the imageContentSourcePolicy")
	deleteIcspCmd.Flags().BoolVar(&waitForNodeReady, "wait-for-node-ready", false, "waits until the nodes are ready to schedule")

}

var deleteIcspCmd = &cobra.Command{
	Use:     "icsp",
	Short:   "deletes imageContentSourcePolicy",
	Aliases: []string{"imagecontentsourcepolicy"},
	Run: func(cmd *cobra.Command, icspNameList []string) {
		if !deleteAll && len(icspNameList) == 0 {
			zap.L().Error("at least a icsp name required or enable --delete-all flag")
			return
		}

		// get kubernetes client
		k8sClient := openshiftClient.GetKubernetesClient()

		if deleteAll {
			err := icspAPI.DeleteOfAll(k8sClient, &v1alpha1.ImageContentSourcePolicy{}, []client.DeleteAllOfOption{})
			if err != nil {
				zap.L().Error("error on deleting all ImageContentSourcePolicy", zap.Error(err))
				rootCmd.ExitWithError()
			}
		} else {
			err := deleteIcsp(k8sClient, icspNameList, false)
			if err != nil {
				zap.L().Error("error on deleting ImageContentSourcePolicy", zap.Any("names", icspNameList), zap.Error(err))
				rootCmd.ExitWithError()
			}
		}

		if waitForNodeReady {
			zap.L().Info("wait for node ready enabled")
			err := nodeAPI.WaitForNodesReady(k8sClient, nodeReadyTimeout)
			if err != nil {
				zap.L().Error("error on waiting to node ready state", zap.Error(err))
				rootCmd.ExitWithError()
			}
			zap.L().Info("nodes are available to schedule")
		}
	},
}

func deleteIcsp(k8sClient client.Client, icspNameList []string, waitForNodeReady bool) error {
	installedList, err := icspAPI.List(k8sClient, []client.ListOption{})
	if err != nil {
		zap.L().Fatal("error on getting list", zap.Error(err))
		return err
	}

	// delete icsp
	for _, icspName := range icspNameList {
		found := false
		for _, icspInstalled := range installedList.Items {
			if icspInstalled.Name == icspName {
				found = true
				zap.L().Debug("deleting an ImageContentSourcePolicy", zap.String("name", icspName))
				err := icspAPI.Delete(k8sClient, &icspInstalled)
				if err != nil {
					zap.L().Error("error on deleting an ImageContentSourcePolicy", zap.String("name", icspName), zap.Error(err))
					continue
				}
				zap.L().Info("deleted an ImageContentSourcePolicy", zap.String("name", icspInstalled.GetName()))
			}
		}
		if !found {
			zap.L().Info("ImageContentSourcePolicy not available", zap.String("name", icspName))
		} else if waitForNodeReady {
			zap.L().Info("wait for node ready enabled")
			err = nodeAPI.WaitForNodesReady(k8sClient, nodeReadyTimeout)
			if err != nil {
				zap.L().Error("error on waiting to node ready state", zap.Error(err))
			} else {
				zap.L().Info("nodes are available to schedule")
			}
		}
	}
	return nil
}
