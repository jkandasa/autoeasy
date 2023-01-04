package get

import (
	openshiftRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"

	"github.com/spf13/cobra"
)

func init() {
	openshiftRootCmd.AddCommand(openshiftInstallCmd)
}

var openshiftInstallCmd = &cobra.Command{
	Use:   "delete",
	Short: "deletes a resource",
}

func AddCommand(cmds ...*cobra.Command) {
	openshiftInstallCmd.AddCommand(cmds...)
}
