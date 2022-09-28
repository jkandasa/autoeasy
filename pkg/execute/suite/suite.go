package execute

import (
	"fmt"
	"time"

	templateStore "github.com/jkandasa/autoeasy/pkg/execute/template"
	variableStore "github.com/jkandasa/autoeasy/pkg/execute/variable"
	fileTY "github.com/jkandasa/autoeasy/pkg/types/file"
	suiteTY "github.com/jkandasa/autoeasy/pkg/types/suite"
	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	variableTY "github.com/jkandasa/autoeasy/pkg/types/variable"
	fileUtils "github.com/jkandasa/autoeasy/pkg/utils/file"
	templateUtils "github.com/jkandasa/autoeasy/pkg/utils/template"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var suiteConfigs = make([]suiteTY.SuiteConfig, 0)

func Load(dir string) error {
	if !fileUtils.IsDirExists(dir) {
		return fmt.Errorf("suites config directory not found. dir:%s", dir)
	}

	files, err := fileUtils.ListFiles(dir)
	if err != nil {
		zap.L().Error("error on listing files", zap.String("dir", dir), zap.Error(err))
		return err
	}

	// load execution configs
	for _, file := range files {
		if file.IsDir {
			continue
		}
		data, err := fileUtils.ReadFile(dir, file.Name)
		if err != nil {
			zap.L().Error("error on reading a file", zap.String("filename", file.FullPath), zap.Error(err))
			return err
		}

		// update variables in suite file
		cfgPre := suiteTY.SuiteConfigPre{}
		// load environment variables
		tplPre, err := templateUtils.Execute(string(data), map[string]interface{}{})
		if err != nil {
			zap.L().Error("error on executing template", zap.String("filename", file.FullPath), zap.Error(err))
			return err
		}
		err = yaml.Unmarshal([]byte(tplPre), &cfgPre)
		if err != nil {
			zap.L().Error("error on yaml unmarshal", zap.String("filename", file.FullPath), zap.Error(err))
			return err
		}

		// load all variables
		suiteDefaultVars, err := getSuiteDefaultVariables(cfgPre.Default.VariablesName)
		if err != nil {
			zap.L().Error("error on loading default variables list", zap.String("filename", file.FullPath), zap.Error(err))
			return err
		}

		// merge all the variables
		vars, err := updateVariables(suiteDefaultVars, cfgPre.Variables)
		if err != nil {
			zap.L().Error("error on merging variables", zap.String("filename", file.FullPath), zap.Error(err))
			return err
		}

		if len(cfgPre.Matrix) > 0 { // switching to matrix mode
			for matrixIndex := range cfgPre.Matrix {
				matrix := cfgPre.Matrix[matrixIndex]
				if matrix.Disabled {
					zap.L().Debug("matrix disabled", zap.String("suiteName", cfgPre.Name), zap.Int("matrixIndex", matrixIndex), zap.String("matrixDescription", matrix.Description))
					continue
				}

				updateVars, err := updateVariables(vars, matrix.Variables)
				if err != nil {
					zap.L().Error("error on merging variables", zap.String("filename", file.FullPath), zap.Int("matrixIndex", matrixIndex), zap.String("matrixDescription", matrix.Description), zap.Error(err))
					return err
				}
				// include it on the list
				suiteCfg, err := getSuite(data, updateVars, &file)
				if err != nil {
					return err
				}
				if suiteCfg != nil {
					suiteCfg.SelectedMatrix = &matrix
					suiteConfigs = append(suiteConfigs, *suiteCfg)
				}
			}
		} else {
			suiteCfg, err := getSuite(data, vars, &file)
			if err != nil {
				return err
			}
			if suiteCfg != nil {
				suiteConfigs = append(suiteConfigs, *suiteCfg)
			}
		}
	}

	return nil
}

func getSuite(suiteBytes []byte, vars variableTY.Variables, file *fileTY.File) (*suiteTY.SuiteConfig, error) {
	// load template with defined variables
	cfg := suiteTY.SuiteConfig{}
	suitePost, err := templateUtils.Execute(string(suiteBytes), vars)
	if err != nil {
		zap.L().Error("error on post executing template with defined variables", zap.String("filename", file.FullPath), zap.Error(err))
		return nil, err
	}

	err = yaml.Unmarshal([]byte(suitePost), &cfg)
	if err != nil {
		zap.L().Error("error on post yaml unmarshal", zap.String("filename", file.FullPath), zap.Error(err))
		return nil, err
	}

	// include filename
	cfg.FileName = file.FullPath
	cfg.RawData = string(suiteBytes)

	// if disabled, do not include
	if cfg.Disabled {
		zap.L().Info("suite disabled", zap.String("filename", cfg.FileName), zap.String("name", cfg.Name))
		return nil, nil
	}

	return &cfg, nil
}

func Execute() error {
	for _, cfg := range suiteConfigs {
		zap.L().Info("about to execute a suite", zap.String("name", cfg.Name), zap.String("filename", cfg.FileName), zap.Int("numbeOfTask", len(cfg.Tasks)))
		startTime := time.Now()
		err := runSuite(&cfg)
		if err != nil {
			return err
		}
		zap.L().Info("suite execution completed", zap.String("name", cfg.Name), zap.String("filename", cfg.FileName), zap.String("timeTaken", time.Since(startTime).String()))
	}
	return nil
}

func runSuite(suiteCfg *suiteTY.SuiteConfig) error {
	for taskIndex, task := range suiteCfg.Tasks {
		// update template
		if task.Template == "" {
			task.Template = suiteCfg.Default.TemplateName
		}

		if task.Disabled {
			zap.L().Info("task disabled", zap.String("taskName", task.Name), zap.String("description", task.Description), zap.String("template", task.Template))
			continue
		}

		zap.L().Info("about to execute a task", zap.String("taskName", task.Name), zap.String("description", task.Description), zap.String("template", task.Template))
		startTime := time.Now()
		err := runTask(suiteCfg, &task, taskIndex)
		if err != nil {
			return err
		}
		zap.L().Info("task execution completed", zap.String("taskName", task.Name), zap.String("description", task.Description), zap.String("template", task.Template), zap.String("timeTaken", time.Since(startTime).String()))
	}
	return nil
}

// steps to execute task
// 1. get template raw bytes
// 2. get template variables
// 3. get variables from suite defaults
// 4. get local variables
// 5. merge all the variables
// 6. execute template with available variables
func runTask(suiteCfg *suiteTY.SuiteConfig, task *suiteTY.Task, taskIndex int) error {
	// get template
	rawTemplate, err := templateStore.Get(task.Template)
	if err != nil {
		zap.L().Error("error on getting template", zap.String("template", task.Template), zap.Error(err))
		return err
	}

	// get suite variables
	suiteDefaultVars, err := getSuiteDefaultVariables(suiteCfg.Default.VariablesName)
	if err != nil {
		zap.L().Error("error on getting suite default variables", zap.String("suiteFilename", suiteCfg.FileName), zap.Error(err))
		return err
	}

	// merge all the variables

	// include matrix variables, if available
	var matrixVars variableTY.Variables
	if suiteCfg.SelectedMatrix != nil {
		zap.L().Info("suite with matrix configuration", zap.String("suiteName", suiteCfg.Name), zap.String("matrixDescription", suiteCfg.SelectedMatrix.Description))
		matrixVars = suiteCfg.SelectedMatrix.Variables
	}

	vars, err := updateVariables(rawTemplate.Variables, suiteDefaultVars, suiteCfg.Variables, matrixVars)
	if err != nil {
		zap.L().Error("error on merging variables", zap.String("suiteFilename", suiteCfg.FileName), zap.Error(err))
		return err
	}

	// update task variables
	suiteData, err := templateUtils.Execute(suiteCfg.RawData, vars)
	if err != nil {
		zap.L().Error("error on applying post template", zap.String("suiteFilename", suiteCfg.FileName), zap.Error(err))
		return err
	}
	suiteCfgUpdated := suiteTY.SuiteConfig{}
	err = yaml.Unmarshal([]byte(suiteData), &suiteCfgUpdated)
	if err != nil {
		zap.L().Error("error on post yaml unmarshal", zap.String("suiteFilename", suiteCfg.FileName), zap.Error(err))
		return err
	}

	// get task variables
	taskVars := suiteCfgUpdated.Tasks[taskIndex].Variables

	// merge with task variables
	vars, err = updateVariables(vars, taskVars)
	if err != nil {
		zap.L().Error("error on merge variables, stage: final", zap.String("suiteFilename", suiteCfg.FileName), zap.String("taskName", task.Name), zap.String("taskDescription", task.Description), zap.Error(err))
		return err
	}

	// load template with defined variables
	updatedData, err := templateUtils.Execute(rawTemplate.RawString, vars)
	if err != nil {
		zap.L().Error("error on loading task template, stage: final", zap.String("suiteFilename", suiteCfg.FileName), zap.String("taskName", task.Name), zap.String("taskDescription", task.Description), zap.String("templateFile", rawTemplate.FileName), zap.Error(err))
		return err
	}

	// convert to actual template
	tpl := templateTY.Template{}
	err = yaml.Unmarshal([]byte(updatedData), &tpl)
	if err != nil {
		zap.L().Error("error on yaml unmarshal", zap.String("suiteFilename", suiteCfg.FileName), zap.String("taskName", task.Name), zap.String("taskDescription", task.Description), zap.String("templateFile", rawTemplate.FileName), zap.Error(err))
		return err
	}

	// get tplTask from the template
	var tplTask *templateTY.Task
	for _, a := range tpl.Tasks {
		if a.Name == task.Name {
			tplTask = &a
			break
		}
	}
	if tplTask == nil {
		return fmt.Errorf("task not available in the template. templateName:%s, taskName:%s, taskDescription:%s", task.Template, task.Name, task.Description)
	}

	// update description from suite, if available
	if task.Description != "" {
		tplTask.Description = task.Description
	}

	// execute task
	err = run(tplTask)
	return err
}

func getSuiteDefaultVariables(variablesNameList []string) (variableTY.Variables, error) {
	variables := variableTY.Variables{}
	for _, varName := range variablesNameList {
		varCfg, err := variableStore.Get(varName)
		if err != nil {
			zap.L().Error("error on getting a variables", zap.String("name", varName), zap.Error(err))
			return nil, err
		}

		// update templates in variables (supports only for environment variables, store variables)
		updatedData, err := templateUtils.Execute(varCfg.RawData, nil)
		if err != nil {
			zap.L().Error("error on template on variables", zap.String("name", varName), zap.Error(err))
			return nil, err
		}

		suiteCfgPre := suiteTY.SuiteConfigPre{}
		err = yaml.Unmarshal([]byte(updatedData), &suiteCfgPre)
		if err != nil {
			zap.L().Error("error on yaml unmarshal on variables", zap.String("name", varName), zap.Error(err))
			return nil, err
		}
		variables.Merge(suiteCfgPre.Variables)
	}
	return variables, nil
}

// merges all the variables, in the given order
func updateVariables(variablesSlice ...variableTY.Variables) (variableTY.Variables, error) {
	variables := variableTY.Variables{}
	for _, vars := range variablesSlice {
		variables.Merge(vars)
	}
	return variables, nil
}
