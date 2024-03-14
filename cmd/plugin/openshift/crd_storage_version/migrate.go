package storageversion

import (
	"context"

	openshiftMigrateCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/migrate"
	crdAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/custom_resource_definition"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	apixclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
)

var (
	skipResourceMigration bool
)

func init() {
	openshiftMigrateCmd.AddCommand(migrateStorageVersionCmd)

	migrateStorageVersionCmd.PersistentFlags().BoolVar(&skipResourceMigration, "skip-resources", false, "skips resources/CRs migration")

}

var migrateStorageVersionCmd = &cobra.Command{
	Use:   "crd-storage-version",
	Short: "migrates CRDs storage version",
	Example: `  # supply a single crd group
  autoeasy openshift migrate crd-storage-version tasks.tekton.dev
  
  # supply multiple crd groups
  autoeasy openshift migrate crd-storage-version tasks.tekton.dev taskruns.tekton.dev

  # skips resources (CRs) migration
  autoeasy openshift migrate crd-storage-version tasks.tekton.dev --skip-resources`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, crdGroups []string) {
		// get kubernetes restConfig
		restConfig := openshiftClient.GetK8SClientConfig().GetRestConfig()

		logger := zap.L().Sugar()

		migrator := crdAPI.NewStorageVersionMigrator(
			dynamic.NewForConfigOrDie(restConfig),
			apixclient.NewForConfigOrDie(restConfig),
			logger,
		)

		migrator.MigrateCrdGroups(context.TODO(), crdGroups, skipResourceMigration)
	},
}
