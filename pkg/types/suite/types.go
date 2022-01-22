package types

import (
	variableTY "github.com/jkandasa/autoeasy/pkg/types/variable"
)

type SuiteConfigPre struct {
	Name      string               `yaml:"name"`
	Disabled  bool                 `yaml:"disabled"`
	Default   Default              `yaml:"default"`
	Variables variableTY.Variables `yaml:"variables"`
}

type SuiteConfig struct {
	Name      string               `yaml:"name"`
	Disabled  bool                 `yaml:"disabled"`
	Default   Default              `yaml:"default"`
	Variables variableTY.Variables `yaml:"variables"`
	Actions   []Action             `yaml:"actions"`
	FileName  string               `yaml:"-"`
}

type Default struct {
	TemplateName  string   `yaml:"template_name"`
	VariablesName []string `yaml:"variables_name"`
}

type Action struct {
	Name      string               `yaml:"name"`
	Template  string               `yaml:"template"`
	Variables variableTY.Variables `yaml:"variables"`
	Disabled  bool                 `yaml:"disabled"`
}
