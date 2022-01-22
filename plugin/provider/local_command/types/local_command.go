package types

import "time"

const (
	DefaultErrorDir      = "./logs/local_command"
	DefaultErrorFilename = "errors.txt"
	DefaultTimeout       = time.Minute * 2
)

// ProviderConfig struct
type ProviderConfig struct {
	Timeout time.Duration `yaml:"timeout"`
	Error   Error         `yaml:"error"`
}

type Error struct {
	Record   bool   `yaml:"record"`
	Dir      string `yaml:"dir"`
	Filename string `yaml:"filename"`
}

func (pc *ProviderConfig) UpdateDefaults() {
	if pc.Timeout <= 0 {
		pc.Timeout = DefaultTimeout
	}
	if pc.Error.Dir == "" {
		pc.Error.Dir = DefaultErrorDir
	}
	if pc.Error.Filename == "" {
		pc.Error.Filename = DefaultErrorFilename
	}
}

// InputConfig struct
type InputConfig struct {
	Data []Command `yaml:"data"`
}

type Command struct {
	Command string            `yaml:"command"`
	Script  string            `yaml:"script"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
	Timeout time.Duration     `yaml:"timeout"`
	Output  Output            `yaml:"output"`
}

type Output struct {
	Dir      string `yaml:"dir"`
	Filename string `yaml:"filename"`
	Append   bool   `yaml:"append"`
}
