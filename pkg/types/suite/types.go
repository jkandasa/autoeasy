package types

import (
	variableTY "github.com/jkandasa/autoeasy/pkg/types/variable"
)

type SuiteConfig struct {
	Name      string               `yaml:"name"`
	Disabled  bool                 `yaml:"disabled"`
	Default   Default              `yaml:"default"`
	Variables variableTY.Variables `yaml:"variables"`
	Actions   []Action             `yaml:"actions"`
	FileName  string               `yaml:"-"`
}

type Default struct {
	TemplateName   string   `yaml:"template_name"`
	VariablesNames []string `yaml:"variables_names"`
}

type Action struct {
	ActionName string               `yaml:"action_name"`
	Template   string               `yaml:"template"`
	Variables  variableTY.Variables `yaml:"variables"`
}
