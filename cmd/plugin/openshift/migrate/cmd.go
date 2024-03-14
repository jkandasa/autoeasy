package get

import (
	openshiftRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"

	"github.com/spf13/cobra"
)

func init() {
	openshiftRootCmd.AddCommand(openshiftMigrateCmd)
}

var openshiftMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "performs migration",
}

func AddCommand(cmds ...*cobra.Command) {
	openshiftMigrateCmd.AddCommand(cmds...)
}
