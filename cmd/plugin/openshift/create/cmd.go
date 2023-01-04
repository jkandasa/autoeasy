package get

import (
	openshiftRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"

	"github.com/spf13/cobra"
)

func init() {
	openshiftRootCmd.AddCommand(openshiftInstallCmd)
}

var openshiftInstallCmd = &cobra.Command{
	Use:   "create",
	Short: "creates a resource",
}

func AddCommand(cmds ...*cobra.Command) {
	openshiftInstallCmd.AddCommand(cmds...)
}
