package icsp

import (
	"time"

	openshiftCreateCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/create"
	icspAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/image_content_source_policy"
	nodeAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/node"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	openshiftTY "github.com/jkandasa/autoeasy/plugin/provider/openshift/types"
	"github.com/openshift/api/operator/v1alpha1"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	waitForNodeReady    bool
	createSource        string
	createMirrors       []string
	createForceRecreate bool

	nodeReadyTimeout = openshiftTY.TimeoutConfig{
		Timeout:              time.Minute * 10,
		ScanInterval:         time.Second * 20,
		ExpectedSuccessCount: 5,
	}
)

func init() {
	openshiftCreateCmd.AddCommand(createIcspCmd)
	createIcspCmd.Flags().StringVar(&createSource, "source", "registry.redhat.io", "source registry")
	createIcspCmd.Flags().StringSliceVar(&createMirrors, "mirror", []string{}, "comma separated mirror registries. registry1,registry2")
	createIcspCmd.Flags().BoolVar(&waitForNodeReady, "wait-for-node-ready", false, "waits until the nodes are ready to schedule")
	createIcspCmd.Flags().BoolVar(&createForceRecreate, "force", false, "deletes the icsp if exists and creates")
}

var createIcspCmd = &cobra.Command{
	Use:     "icsp",
	Short:   "installs ImageContentSourcePolicy",
	Aliases: []string{"imagecontentsourcepolicy"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, icspName []string) {

		if len(createMirrors) == 0 {
			zap.L().Error("mirror registries can not be empty")
			return
		}

		// get kubernetes client
		k8sClient := openshiftClient.GetKubernetesClient()

		// deletes icsp if recreate enabled
		if createForceRecreate {
			err := deleteIcsp(k8sClient, icspName, false)
			if err != nil {
				zap.L().Error("error on deleting an ImageContentSourcePolicy", zap.Any("name", icspName[0]), zap.Error(err))
				return
			}
		}

		// create icsp
		icsp := v1alpha1.ImageContentSourcePolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: icspName[0],
			},
			Spec: v1alpha1.ImageContentSourcePolicySpec{
				RepositoryDigestMirrors: []v1alpha1.RepositoryDigestMirrors{
					{
						Source:  createSource,
						Mirrors: createMirrors,
					},
				},
			},
		}

		err := icspAPI.Create(k8sClient, &icsp)
		if err != nil {
			zap.L().Error("error on creating an ImageContentSourcePolicy", zap.String("name", icspName[0]), zap.Error(err))
			return
		}

		zap.L().Info("ImageContentSourcePolicy created", zap.String("name", icspName[0]))
		if waitForNodeReady {
			zap.L().Info("wait for node ready enabled")
			err = nodeAPI.WaitForNodesReady(k8sClient, nodeReadyTimeout)
			if err != nil {
				zap.L().Error("error on waiting to node ready state", zap.Error(err))
			} else {
				zap.L().Info("nodes are available to schedule")
			}
		}
	},
}
