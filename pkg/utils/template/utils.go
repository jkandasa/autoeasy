package utils

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
	"time"

	dataRepoSVC "github.com/jkandasa/autoeasy/pkg/service/data_repository"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func Execute(templateString string, variables map[string]interface{}) (string, error) {
	if variables == nil {
		variables = map[string]interface{}{}
	}
	t, err := template.New("custom-template").Funcs(getFuncMap()).Parse(templateString)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	err = t.Execute(&tpl, variables)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"now":   time.Now,
		"yaml":  toYaml,
		"env":   getEnv,
		"store": getValueFromRepo,
	}
}

func toYaml(v interface{}) string {
	a, err := yaml.Marshal(v)
	if err != nil {
		zap.L().Error("error on converting to yaml", zap.Any("input", v), zap.Error(err))
		return ""
	}
	return string(a)
}

func getEnv(v interface{}) string {
	return os.Getenv(fmt.Sprintf("%s", v))
}

func getValueFromRepo(v interface{}) interface{} {
	value := dataRepoSVC.Get(fmt.Sprintf("%s", v))
	if value != nil {
		return value
	}
	return ""
}
