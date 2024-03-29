package root

import (
	"os"
	"path/filepath"

	suiteStore "github.com/jkandasa/autoeasy/pkg/execute/suite"
	templateStore "github.com/jkandasa/autoeasy/pkg/execute/template"
	variableStore "github.com/jkandasa/autoeasy/pkg/execute/variable"
	providerSVC "github.com/jkandasa/autoeasy/pkg/service/provider"
	"github.com/jkandasa/autoeasy/pkg/types"
	templateUtils "github.com/jkandasa/autoeasy/pkg/utils/template"
	"github.com/jkandasa/autoeasy/pkg/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	resourceDir  string
	pluginConfig string
)

const (
	templateDir  = "templates"
	variablesDir = "variables"
	suitesDir    = "suites"
)

func init() {
	rootCmd.AddCommand(executeCmd)

	executeCmd.Flags().StringVar(&resourceDir, "resource-dir", "./resources", "resources directory")
	executeCmd.Flags().StringVar(&pluginConfig, "plugin-config", "./plugin.yaml", "plugin config file")
}

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "executes the provided tasks",
	Example: `  # simple
  autoeasy execute

	# with custom location
  autoeasy execute --resource-dir=/tmp/resources --plugin-config=/tmp/plugin.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// print this tool details
		zap.L().Debug("this tool information", zap.Any("version", version.Get()))
		zap.L().Info("user input", zap.String("resource-dir", resourceDir), zap.String("plugin-config", pluginConfig))

		// load providers
		bytes, err := os.ReadFile(pluginConfig)
		if err != nil {
			zap.L().Error("error on reading plugin config file", zap.String("plugin-config", pluginConfig), zap.Error(err))
			ExitWithError()
		}
		// update environment variables in plugin file
		updatedPluginConfig, err := templateUtils.Execute(string(bytes), nil)
		if err != nil {
			zap.L().Error("error on updating plugin environment variables", zap.Error(err))
			ExitWithError()
		}

		pluginData := &types.PluginFile{}
		err = yaml.Unmarshal([]byte(updatedPluginConfig), pluginData)
		if err != nil {
			zap.L().Error("error on unmarshal plugin config file", zap.String("plugin-config", pluginConfig), zap.Error(err))
			ExitWithError()
		}
		err = providerSVC.Start(pluginData.Provider)
		if err != nil {
			zap.L().Error("error on loading a provider", zap.Error(err))
			ExitWithError()
		}

		// load templates
		templateDirPath := filepath.Join(resourceDir, templateDir)
		err = templateStore.LoadTemplates(templateDirPath)
		if err != nil {
			zap.L().Error("error on loading template files", zap.String("templateDir", templateDirPath), zap.Error(err))
			ExitWithError()
		}

		// load variables
		variablesDirPath := filepath.Join(resourceDir, variablesDir)
		err = variableStore.LoadVariables(variablesDirPath)
		if err != nil {
			zap.L().Error("error on loading variable files", zap.String("variablesDir", variablesDirPath), zap.Error(err))
			ExitWithError()
		}

		// load executions
		suitesDirPath := filepath.Join(resourceDir, suitesDir)
		err = suiteStore.Load(suitesDirPath)
		if err != nil {
			zap.L().Error("error on loading suites files", zap.String("suitesDir", suitesDirPath), zap.Error(err))
			ExitWithError()
		}

		// execute tasks
		err = suiteStore.Execute()
		if err != nil {
			zap.L().Error("error on execution", zap.Error(err))
			ExitWithError()
		}
	},
}
