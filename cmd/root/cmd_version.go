package root

import (
	"fmt"

	"github.com/jkandasa/autoeasy/pkg/json"
	"github.com/jkandasa/autoeasy/pkg/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the tool version details",
	Run: func(cmd *cobra.Command, args []string) {
		version, err := json.MarshalToString(version.Get())
		if err != nil {
			zap.L().Error("error on getting version details", zap.Error(err))
			ExitWithError()
		}
		fmt.Println(version)
	},
}
