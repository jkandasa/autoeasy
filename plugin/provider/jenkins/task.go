package jenkins_provider

import (
	"fmt"

	jenkinsProviderTY "github.com/jkandasa/autoeasy/plugin/provider/jenkins/types"
)

// execute jenkins task
func (j *Jenkins) run(cfg *jenkinsProviderTY.ProviderConfig) (interface{}, error) {
	switch cfg.Function {
	case jenkinsProviderTY.FunctionBuild:
		return j.build(cfg)

	default:
		return nil, fmt.Errorf("invalid function:%s", cfg.Function)
	}
}
