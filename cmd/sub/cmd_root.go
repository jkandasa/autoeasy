package sub

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	enableStacktrace bool
	logLevel         string
)

func init() {
	RootCmd.PersistentFlags().BoolVar(&enableStacktrace, "enable-stacktrace", false, "enable error stacktrace")
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "console log level, options: debug, info, warn, error, fatal")
}

var RootCmd = &cobra.Command{
	Use:   "autoeasy",
	Short: "autoeasy is tool can be used to convert the manual works auto",
}

func Execute() {

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
