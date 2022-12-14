package sub

import (
	"os"

	loggerSVC "github.com/jkandasa/autoeasy/pkg/service/logger"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	"github.com/spf13/cobra"
)

var (
	HideHeader       bool
	Pretty           bool
	OutputFormat     string
	logLevel         string
	enableStacktrace bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&OutputFormat, "output", "o", printer.OutputConsole, "output format. options: yaml, json, console, wide")
	rootCmd.PersistentFlags().BoolVar(&HideHeader, "hide-header", false, "hides the header on the console output")
	rootCmd.PersistentFlags().BoolVar(&Pretty, "pretty", false, "JSON pretty print")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "warn", "log level. options: debug, info, warn, error, fatal")
	rootCmd.PersistentFlags().BoolVar(&enableStacktrace, "stacktrace", false, "enables stacktrace of an error")
}

var rootCmd = &cobra.Command{
	Use:   "autoeasy",
	Short: "autoeasy tool can perform automation tasks from yaml files also support command line operations",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// load logger
		loggerSVC.Load(logLevel, enableStacktrace)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func AddCommand(cmds ...*cobra.Command) {
	rootCmd.AddCommand(cmds...)
}
