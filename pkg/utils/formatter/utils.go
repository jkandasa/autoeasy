package utils

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

const (
	TagNameYaml = "yaml"
	TagNameJSON = "json"
	TagNameNone = ""
)

// YamlInterfaceToStruct converts map to struct
func YamlInterfaceToStruct(in interface{}, out interface{}) error {
	bytes, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, out)
}

// mapToStructDecodeHookFunc will be called on MapToStruct
func mapToStructDecodeHookFunc(fromType reflect.Type, toType reflect.Type, value interface{}) (interface{}, error) {
	switch toType {
	case reflect.TypeOf(time.Time{}):
		value, err := time.Parse(time.RFC3339Nano, value.(string))
		if err != nil {
			return nil, err
		}
		return value, nil

	}
	return value, nil
}

// MapToStruct converts map to struct
func MapToStruct(tagName string, in map[string]interface{}, out interface{}) error {
	if tagName == "" {
		return mapstructure.Decode(in, out)
	}
	cfg := &mapstructure.DecoderConfig{
		TagName:          tagName,
		WeaklyTypedInput: true,
		DecodeHook:       mapToStructDecodeHookFunc,
		Result:           out,
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return err
	}
	return decoder.Decode(in)
}

// JsonMapToStruct converts map to strict
func JsonMapToStruct(in map[string]interface{}, out interface{}) error {
	return MapToStruct(TagNameJSON, in, out)
}
