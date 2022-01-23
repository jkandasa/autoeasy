package types

const (
	// OnFailure
	OnFailureExit     = "exit"
	OnFailureRepeat   = "repeat"
	OnFailureContinue = "continue"
)

type RawTemplate struct {
	Name      string `yaml:"name"`
	FileName  string `yaml:"-"`
	RawString string `yaml:"-"`
}

type Template struct {
	Name  string `yaml:"name"`
	Tasks []Task `yaml:"tasks"`
}

type Task struct {
	Name      string                 `yaml:"name"`
	Template  string                 `yaml:"template"`
	OnFailure string                 `yaml:"on_failure"`
	Provider  string                 `yaml:"provider"`
	Input     map[string]interface{} `yaml:"input"`
}
