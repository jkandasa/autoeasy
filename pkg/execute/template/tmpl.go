package template

import (
	"fmt"

	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	fileUtils "github.com/jkandasa/autoeasy/pkg/utils/file"
	templateUtils "github.com/jkandasa/autoeasy/pkg/utils/template"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func LoadTemplates(dir string) error {
	if !fileUtils.IsDirExists(dir) {
		return fmt.Errorf("template directory not found. dir:%s", dir)
	}

	files, err := fileUtils.ListFiles(dir)
	if err != nil {
		zap.L().Error("error on geting files", zap.String("dir", dir), zap.Error(err))
		return err
	}

	// load templates
	for _, file := range files {
		if file.IsDir {
			continue
		}
		data, err := fileUtils.ReadFile(dir, file.Name)
		if err != nil {
			zap.L().Error("error on reading a file", zap.String("filename", file.FullPath), zap.Error(err))
			return err
		}

		// get variables and update
		// execute template and get updated variables
		templateVariableRaw, err := templateUtils.Execute(string(data), nil)
		if err != nil {
			zap.L().Error("error on applying template", zap.String("filename", file.FullPath), zap.Error(err))
			return err
		}

		// get template variables
		tplPre := templateTY.TemplatePre{}
		err = yaml.Unmarshal([]byte(templateVariableRaw), &tplPre)
		if err != nil {
			zap.L().Error("error on yaml unmarshal", zap.String("filename", file.FullPath), zap.Error(err))
			return err
		}

		tmpl := &templateTY.RawTemplate{
			Name:      file.Name,
			FileName:  file.FullPath,
			RawString: string(data),
			Variables: tplPre.Variables,
		}

		err = add(tmpl)
		if err != nil {
			zap.L().Error("error on adding into store", zap.String("filename", file.FullPath), zap.Error(err))
			return err
		}
	}

	return nil
}
