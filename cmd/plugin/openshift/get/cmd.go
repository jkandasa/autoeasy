package get

import (
	openshiftRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"

	"github.com/spf13/cobra"
)

func init() {
	openshiftRootCmd.AddCommand(openshiftGetCmd)
}

var openshiftGetCmd = &cobra.Command{
	Use:   "get",
	Short: "displays resource details",
}

func AddCommand(cmds ...*cobra.Command) {
	openshiftGetCmd.AddCommand(cmds...)
}
