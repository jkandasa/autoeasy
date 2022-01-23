package execute

import (
	"fmt"
	"time"

	templateStore "github.com/jkandasa/autoeasy/pkg/execute/template"
	variableStore "github.com/jkandasa/autoeasy/pkg/execute/variable"
	suiteTY "github.com/jkandasa/autoeasy/pkg/types/suite"
	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	variableTY "github.com/jkandasa/autoeasy/pkg/types/variable"
	fileUtils "github.com/jkandasa/autoeasy/pkg/utils/file"
	templateUtils "github.com/jkandasa/autoeasy/pkg/utils/template"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var executionConfigs = make([]suiteTY.SuiteConfig, 0)

func Load(dir string) error {
	if !fileUtils.IsDirExists(dir) {
		return fmt.Errorf("suites config directory not found. dir:%s", dir)
	}

	files, err := fileUtils.ListFiles(dir)
	if err != nil {
		return err
	}

	// load execution configs
	for _, file := range files {
		if file.IsDir {
			continue
		}
		data, err := fileUtils.ReadFile(dir, file.Name)
		if err != nil {
			return err
		}

		// update variables in suite file
		cfgPre := suiteTY.SuiteConfigPre{}
		// load environment variables
		tplPre, err := templateUtils.Execute(string(data), map[string]interface{}{})
		if err != nil {
			return err
		}
		err = yaml.Unmarshal([]byte(tplPre), &cfgPre)
		if err != nil {
			return err
		}

		// load all variables
		vars, err := updateVariables(cfgPre.Default.VariablesName, cfgPre.Variables, nil)
		if err != nil {
			return err
		}

		// load template with defined variables
		cfg := suiteTY.SuiteConfig{}
		tplPost, err := templateUtils.Execute(string(data), *vars)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal([]byte(tplPost), &cfg)
		if err != nil {
			return err
		}

		// include filename
		cfg.FileName = file.FullPath

		// if disabled, do not include
		if cfg.Disabled {
			zap.L().Info("suite disabled", zap.String("filename", cfg.FileName))
			continue
		}

		executionConfigs = append(executionConfigs, cfg)
	}

	return nil
}

func Execute() error {
	for _, cfg := range executionConfigs {
		zap.L().Info("about to execute a suite", zap.String("name", cfg.Name), zap.String("filename", cfg.FileName), zap.Int("numbeOfTask", len(cfg.Tasks)))
		startTime := time.Now()
		err := runExecution(&cfg)
		if err != nil {
			return err
		}
		zap.L().Info("suite execution completed", zap.String("name", cfg.Name), zap.String("filename", cfg.FileName), zap.String("timeTaken", time.Since(startTime).String()))
	}
	return nil
}

func runExecution(exeCfg *suiteTY.SuiteConfig) error {
	// TODO: load default variables
	for _, task := range exeCfg.Tasks {
		// update template
		if task.Template == "" {
			task.Template = exeCfg.Default.TemplateName
		}

		if task.Disabled {
			zap.L().Info("task disabled", zap.String("taskName", task.Name), zap.String("template", task.Template))
			continue
		}

		zap.L().Info("about to execute a task", zap.String("taskName", task.Name), zap.String("template", task.Template))
		startTime := time.Now()
		err := runTask(exeCfg, &task)
		if err != nil {
			return err
		}
		zap.L().Info("task execution completed", zap.String("taskName", task.Name), zap.String("template", task.Template), zap.String("timeTaken", time.Since(startTime).String()))
	}
	return nil
}

func runTask(exeCfg *suiteTY.SuiteConfig, task *suiteTY.Task) error {
	// get template variables
	rawTemplate, err := templateStore.Get(task.Template)
	if err != nil {
		return err
	}

	// update variables
	vars, err := updateVariables(exeCfg.Default.VariablesName, exeCfg.Variables, task.Variables)
	if err != nil {
		return err
	}

	// load template with defined variables
	updatedData, err := templateUtils.Execute(rawTemplate.RawString, *vars)
	if err != nil {
		return err
	}

	// convert to actual template
	tpl := templateTY.Template{}
	err = yaml.Unmarshal([]byte(updatedData), &tpl)
	if err != nil {
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
		return fmt.Errorf("task not available in the template. templateName:%s, taskName:%s", task.Template, task.Name)
	}

	// execute task
	err = run(tplTask)
	return err
}

func updateVariables(variableNames []string, exeVariables, localVariables variableTY.Variables) (*variableTY.Variables, error) {
	variables := variableTY.Variables{}
	for _, varName := range variableNames {
		varCfg, err := variableStore.Get(varName)
		if err != nil {
			return nil, err
		}

		// update templates in variables (supports only for environment variables)
		updatedData, err := templateUtils.Execute(varCfg.RawData, nil)
		if err != nil {
			return nil, err
		}

		vars := variableTY.Variables{}
		err = yaml.Unmarshal([]byte(updatedData), &vars)
		if err != nil {
			return nil, err
		}
		variables.Merge(vars)
	}

	// merge executable vars
	variables.Merge(exeVariables)
	// merge localvars
	variables.Merge(localVariables)

	return &variables, nil
}
