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
	for _, action := range exeCfg.Actions {
		// update template
		if action.Template == "" {
			action.Template = exeCfg.Default.TemplateName
		}

		if action.Disabled {
			zap.L().Info("action disabled", zap.String("actionName", action.Name), zap.String("template", action.Template))
			continue
		}

		zap.L().Info("about to execute an action", zap.String("actionName", action.Name), zap.String("template", action.Template))
		startTime := time.Now()
		err := runAction(exeCfg, &action)
		if err != nil {
			return err
		}
		zap.L().Info("action execution completed", zap.String("actionName", action.Name), zap.String("template", action.Template), zap.String("timeTaken", time.Since(startTime).String()))
	}
	return nil
}

func runAction(exeCfg *suiteTY.SuiteConfig, action *suiteTY.Action) error {
	// get template variables
	rawTemplate, err := templateStore.Get(action.Template)
	if err != nil {
		return err
	}

	// update variables
	vars, err := updateVariables(exeCfg.Default.VariablesName, exeCfg.Variables, action.Variables)
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

	// get tplAction from the template
	var tplAction *templateTY.Action
	for _, a := range tpl.Actions {
		if a.Name == action.Name {
			tplAction = &a
			break
		}
	}
	if tplAction == nil {
		return fmt.Errorf("action not available in the template. templateName:%s, actionName:%s", action.Template, action.Name)
	}

	// execute action
	err = run(tplAction)
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
