package types

type Variables map[string]interface{}

func (v Variables) Merge(variables Variables) {
	for key, val := range variables {
		v[key] = val
	}
}

func (v Variables) Get(key string) interface{} {
	value, found := v[key]
	if found {
		return value
	}
	return nil
}

type VariableConfig struct {
	Name      string    `yaml:"name"`
	Variables Variables `yaml:"variables"`
	FileName  string    `yaml:"-"`
}
