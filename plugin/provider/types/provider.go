package provider

import (
	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
)

// Plugin interface details for operation
type Plugin interface {
	Name() string
	Start() error
	Close() error
	Execute(task *templateTY.Task) error
}
