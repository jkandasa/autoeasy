package catalogsource

import (
	openshiftUninstallCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/uninstall"
	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	csAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/catalog_source"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	openshiftUninstallCmd.AddCommand(uninstallCsCmd)
}

var uninstallCsCmd = &cobra.Command{
	Use:   "catalog-source",
	Short: "removes catalog source",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, catalogSourceList []string) {
		// get kubernetes client
		k8sClient := openshiftClient.GetKubernetesClient()

		uninstallCatalogSources(k8sClient, catalogSourceList)
	},
}

func uninstallCatalogSources(k8sClient client.Client, catalogSourceList []string) {
	csList, err := csAPI.List(k8sClient, []client.ListOption{})
	if err != nil {
		zap.L().Error("error on getting catalog source list", zap.Error(err))
		rootCmd.ExitWithError()
	}
	for _, catalogSourceName := range catalogSourceList {
		found := false
		for _, _cs := range csList.Items {
			if _cs.Name == catalogSourceName {
				found = true
				err = csAPI.Delete(k8sClient, &_cs)
				if err != nil {
					zap.L().Error("error on deleting catalog source", zap.String("name", _cs.GetName()), zap.String("namespace", _cs.GetName()), zap.Error(err))
					continue
				}
				zap.L().Info("uninstalled a catalog source", zap.String("name", _cs.GetName()), zap.String("namespace", _cs.GetName()), zap.String("image", _cs.Spec.Image))
			}
			if found {
				break
			}
		}
	}
}
