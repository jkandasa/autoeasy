package diagnose

import (
	openshiftRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"

	"github.com/spf13/cobra"
)

func init() {
	openshiftRootCmd.AddCommand(openshiftDiagnoseCmd)
}

var openshiftDiagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "performs diagnose on OpenShift, see the sub commands",
}

func AddCommand(cmds ...*cobra.Command) {
	openshiftDiagnoseCmd.AddCommand(cmds...)
}
