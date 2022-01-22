package variable

import (
	"fmt"

	variableTY "github.com/jkandasa/autoeasy/pkg/types/variable"
	fileUtils "github.com/jkandasa/autoeasy/pkg/utils/file"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func LoadVariables(dir string) error {
	if !fileUtils.IsDirExists(dir) {
		return fmt.Errorf("variables directory not found. dir:%s", dir)
	}

	files, err := fileUtils.ListFiles(dir)
	if err != nil {
		return err
	}

	// load variables
	for _, file := range files {
		if file.IsDir {
			continue
		}
		zap.L().Debug("loading variable", zap.String("filename", file.Name))
		data, err := fileUtils.ReadFile(dir, file.Name)
		if err != nil {
			return err
		}
		varCfg := &variableTY.VariableConfigPre{}
		err = yaml.Unmarshal(data, varCfg)
		if err != nil {
			return err
		}
		// include filename and raw data
		varCfg.FileName = file.FullPath
		varCfg.RawData = string(data)
		err = add(varCfg)
		if err != nil {
			return err
		}
	}

	return nil
}
