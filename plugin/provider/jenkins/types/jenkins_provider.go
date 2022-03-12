package jenkins_provider

import (
	"time"

	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
)

const (
	FunctionBuild   = "build"
	FunctionRebuild = "rebuild"
)

// Provider configuration
type ProviderConfig struct {
	Function string        `yaml:"function"`
	Config   TaskConfig    `yaml:"config"`
	Data     []interface{} `yaml:"data"`
}

type TaskConfig struct {
	WaitForCompletion bool          `yaml:"wait_for_completion"`
	RetryCount        int           `yaml:"retry_count"`
	Timeout           time.Duration `yaml:"timeout"`
}

// converts the data to build data
func (p *ProviderConfig) GetBuildData() ([]BuildData, error) {
	buildData := make([]BuildData, 0)
	err := formatterUtils.YamlInterfaceToStruct(p.Data, &buildData)
	if err != nil {
		return nil, err
	}
	return buildData, nil
}

// build data
type BuildData struct {
	JobName    string            `yaml:"job_name"`
	Limit      int               `yaml:"limit"`
	Parameters map[string]string `yaml:"parameters"`
}
