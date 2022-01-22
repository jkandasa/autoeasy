package sub

import (
	"fmt"
	"os"

	"github.com/jkandasa/autoeasy/pkg/json"
	loggerSVC "github.com/jkandasa/autoeasy/pkg/service/logger"
	"github.com/jkandasa/autoeasy/pkg/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the tool version details",
	Run: func(cmd *cobra.Command, args []string) {
		// load logger
		loggerSVC.Load(logLevel, enableStacktrace)

		version, err := json.MarshalToString(version.Get())
		if err != nil {
			zap.L().Error("error on getting version details", zap.Error(err))
			os.Exit(1)
		}
		fmt.Println(version)
	},
}
