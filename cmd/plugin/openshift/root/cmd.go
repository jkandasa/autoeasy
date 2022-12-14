package openshift

import (
	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(openshiftCmd)
}

var openshiftCmd = &cobra.Command{
	Use:   "openshift",
	Short: "perform actions on openshift",
}

func AddCommand(cmds ...*cobra.Command) {
	openshiftCmd.AddCommand(cmds...)
}
