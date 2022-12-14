package catalogsource

import (
	"os"

	openshiftGetCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/get"
	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	csAPI "github.com/jkandasa/autoeasy/plugin/provider/openshift/api/catalog_source"
	openshiftClient "github.com/jkandasa/autoeasy/plugin/provider/openshift/client"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	openshiftGetCmd.AddCommand(getCsCmd)
}

var getCsCmd = &cobra.Command{
	Use:   "catalog-source",
	Short: "lists catalog source",
	Run: func(cmd *cobra.Command, catalogSourceList []string) {
		// get kubernetes client
		k8sClient := openshiftClient.GetKubernetesClient()

		csList, err := csAPI.List(k8sClient, []client.ListOption{})
		if err != nil {
			zap.L().Fatal("error on getting catalog source list", zap.Error(err))
		}

		headers := []printer.Header{
			{Title: "namespace", ValuePath: "objectMeta.namespace"},
			{Title: "name", ValuePath: "objectMeta.name"},
			{Title: "display name", ValuePath: "spec.displayName"},
			{Title: "source type", ValuePath: "spec.sourceType"},
			{Title: "image", ValuePath: "spec.image"},
		}

		rows := make([]interface{}, len(csList.Items))
		for index, cs := range csList.Items {
			rows[index] = cs
		}

		printer.Print(os.Stdout, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
	},
}
