package get

import (
	openshiftRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"

	"github.com/spf13/cobra"
)

func init() {
	openshiftRootCmd.AddCommand(openshiftUninstallCmd)
}

var openshiftUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstalls a resource",
}

func AddCommand(cmds ...*cobra.Command) {
	openshiftUninstallCmd.AddCommand(cmds...)
}
