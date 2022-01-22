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
		cfg := suiteTY.SuiteConfig{}

		tpl, err := templateUtils.Execute(string(data), map[string]interface{}{})
		if err != nil {
			return err
		}

		err = yaml.Unmarshal([]byte(tpl), &cfg)
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
		zap.L().Info("about to execute a suite", zap.String("name", cfg.Name), zap.String("filename", cfg.FileName), zap.Int("numbeOfAction", len(cfg.Actions)))
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
	for _, execution := range exeCfg.Actions {
		// update template
		if execution.Template == "" {
			execution.Template = exeCfg.Default.TemplateName
		}

		zap.L().Info("about to execute an action", zap.String("actionName", execution.ActionName), zap.String("template", execution.Template))
		startTime := time.Now()
		err := runAction(exeCfg, &execution)
		if err != nil {
			return err
		}
		zap.L().Info("action execution completed", zap.String("actionName", execution.ActionName), zap.String("template", execution.Template), zap.String("timeTaken", time.Since(startTime).String()))
	}
	return nil
}

func runAction(exeCfg *suiteTY.SuiteConfig, execution *suiteTY.Action) error {
	// get template variables
	rawTemplate, err := templateStore.Get(execution.Template)
	if err != nil {
		return err
	}

	// update variables
	vars, err := updateVariables(exeCfg.Default.VariablesNames, exeCfg.Variables, execution.Variables)
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

	// get action from the template
	var action *templateTY.Action
	for _, a := range tpl.Actions {
		if a.Name == execution.ActionName {
			action = &a
			break
		}
	}
	if action == nil {
		return fmt.Errorf("action not available in the template. templateName:%s, actionName:%s", execution.Template, execution.ActionName)
	}

	// execute action
	err = run(action)
	return err
}

func updateVariables(variableNames []string, exeVariables, localVariables variableTY.Variables) (*variableTY.Variables, error) {
	variables := variableTY.Variables{}
	for _, varName := range variableNames {
		varCfg, err := variableStore.Get(varName)
		if err != nil {
			return nil, err
		}
		// execute as template
		// convert to string
		data, err := yaml.Marshal(varCfg.Variables)
		if err != nil {
			return nil, err
		}
		updatedData, err := templateUtils.Execute(string(data), nil)
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
