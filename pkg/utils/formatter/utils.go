package utils

import (
	"gopkg.in/yaml.v3"
)

// YamlInterfaceToStruct converts map to struct
func YamlInterfaceToStruct(in interface{}, out interface{}) error {
	bytes, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, out)
}
