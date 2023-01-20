package catalogsource

import (
	openshiftInstallCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/install"
	openshiftRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"
	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	csAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/catalog_source"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	installCsNamespace     string
	installCsDisplayName   string
	installCsImage         string
	installCsSourceType    string
	installCsForceRecreate bool
)

func init() {
	openshiftInstallCmd.AddCommand(installCsCmd)
	installCsCmd.Flags().StringVar(&installCsNamespace, "namespace", "openshift-marketplace", "namespace of the catalog source")
	installCsCmd.Flags().StringVar(&installCsDisplayName, "display-name", "", "display name")
	installCsCmd.Flags().StringVar(&installCsImage, "image", openshiftRootCmd.DefaultRegistryImage, "registry image")
	installCsCmd.Flags().StringVar(&installCsSourceType, "source-type", "grpc", "source type of the catalog source. options: grpc, internal, configmap")
	installCsCmd.Flags().BoolVar(&installCsForceRecreate, "force", false, "force recreate the catalog source, if exists")
}

var installCsCmd = &cobra.Command{
	Use:   "catalog-source",
	Short: "installs catalog source",
	Args:  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, catalogSourceNames []string) {
		catalogSourceName := catalogSourceNames[0]
		// get kubernetes client
		k8sClient := openshiftClient.GetKubernetesClient()

		if installCsForceRecreate {
			uninstallCatalogSources(k8sClient, catalogSourceNames)
		}

		// install a catalog source
		catalogSource := corsosv1alpha1.CatalogSource{
			ObjectMeta: v1.ObjectMeta{
				Name:      catalogSourceName,
				Namespace: installCsNamespace,
			},
			Spec: corsosv1alpha1.CatalogSourceSpec{
				SourceType:  corsosv1alpha1.SourceType(installCsSourceType),
				DisplayName: installCsDisplayName,
				Image:       installCsImage,
			},
		}

		err := csAPI.Create(k8sClient, &catalogSource)
		if err != nil {
			zap.L().Error("error on creating a catalog source", zap.String("name", catalogSource.GetName()), zap.String("namespace", catalogSource.GetName()), zap.String("image", catalogSource.Spec.Image), zap.Error(err))
			rootCmd.ExitWithError()
		}
		zap.L().Info("installed a catalog source", zap.String("name", catalogSource.GetName()), zap.String("namespace", catalogSource.GetName()), zap.String("image", catalogSource.Spec.Image))
	},
}
