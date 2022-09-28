package types

import (
	variableTY "github.com/jkandasa/autoeasy/pkg/types/variable"
)

type SuiteConfigPre struct {
	Name        string               `yaml:"name"`
	Description string               `yaml:"description"`
	Disabled    bool                 `yaml:"disabled"`
	Default     DefaultConfig        `yaml:"default"`
	Variables   variableTY.Variables `yaml:"variables"`
	Matrix      []MatrixConfig       `yaml:"matrix"`
}

type SuiteConfig struct {
	Name           string               `yaml:"name"`
	Description    string               `yaml:"description"`
	Disabled       bool                 `yaml:"disabled"`
	Default        DefaultConfig        `yaml:"default"`
	Variables      variableTY.Variables `yaml:"variables"`
	Matrix         []MatrixConfig       `yaml:"matrix"`
	Tasks          []Task               `yaml:"tasks"`
	SelectedMatrix *MatrixConfig        `yaml:"-"`
	FileName       string               `yaml:"-"`
	RawData        string               `yaml:"-"`
}

type DefaultConfig struct {
	TemplateName  string   `yaml:"template_name"`
	VariablesName []string `yaml:"variables_name"`
}

type Task struct {
	Description string               `yaml:"description"`
	Name        string               `yaml:"name"`
	Template    string               `yaml:"template"`
	Variables   variableTY.Variables `yaml:"variables"`
	Disabled    bool                 `yaml:"disabled"`
}

type MatrixConfig struct {
	Disabled    bool                 `yaml:"disabled"`
	Description string               `yaml:"description"`
	Variables   variableTY.Variables `yaml:"variables"`
}
