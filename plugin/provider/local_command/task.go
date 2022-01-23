package local_command

import (
	"fmt"
	"time"

	goCmd "github.com/go-cmd/cmd"
	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	utils "github.com/jkandasa/autoeasy/pkg/utils"
	commandUtils "github.com/jkandasa/autoeasy/pkg/utils/command"
	fileUtils "github.com/jkandasa/autoeasy/pkg/utils/file"
	formatterUtils "github.com/jkandasa/autoeasy/pkg/utils/formatter"
	commandTY "github.com/jkandasa/autoeasy/plugin/provider/local_command/types"
	"go.uber.org/zap"
)

const (
	defaultStatusUpdateDuration = time.Second * 10
	defaultDir                  = "./logs/command_output"
	generatedScriptDir          = "./generated_scripts"
)

func (lc *LocalCommand) run(task *templateTY.Task) (interface{}, error) {
	cfg := commandTY.InputConfig{}

	err := formatterUtils.YamlInterfaceToStruct(task.Input, &cfg)
	if err != nil {
		return nil, err
	}

	for _, data := range cfg.Data {
		err := lc.executeCmd(task, data)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (lc *LocalCommand) executeCmd(task *templateTY.Task, data interface{}) error {
	cmd := commandTY.Command{}
	err := formatterUtils.YamlInterfaceToStruct(data, &cmd)
	if err != nil {
		return err
	}
	zap.L().Debug("executing a local command", zap.String("taskName", task.Name), zap.String("command", cmd.Command))

	if cmd.Timeout <= 0 {
		cmd.Timeout = lc.Config.Timeout
	}

	// setup script if defined
	if cmd.Script != "" {
		scriptname := fmt.Sprintf("%s.sh", utils.RandIDWithLength(20))
		err = fileUtils.WriteFile(generatedScriptDir, scriptname, []byte(cmd.Script))
		if err != nil {
			return err
		}
		scriptfile := fmt.Sprintf("%s/%s", generatedScriptDir, scriptname)
		// on exit remove the file
		defer func() {
			err := fileUtils.RemoveFileOrEmptyDir(scriptfile)
			if err != nil {
				zap.L().Error("error on deleting generated script file", zap.String("scriptFile", scriptfile), zap.Error(err))
			}
		}()

		cmd.Command = "sh"
		cmd.Args = []string{scriptfile}
	}

	// execute command
	command := commandUtils.Command{
		Name:                 task.Name,
		Command:              cmd.Command,
		Args:                 cmd.Args,
		Env:                  getEnv(cmd.Env),
		Timeout:              cmd.Timeout,
		StatusUpdateDuration: defaultStatusUpdateDuration,
	}
	status := goCmd.Status{}
	result := ""

	ExitFn := func(rxResult string, rxStatus goCmd.Status) {
		result = rxResult
		status = rxStatus
	}

	command.ExitFn = ExitFn
	err = command.StartAndWait()
	if err != nil {
		return err
	}

	if result != commandUtils.ExitTypeNormal {
		return fmt.Errorf("command complated with '%s' state", result)
	}

	errorOutput := ""
	for _, line := range status.Stderr {
		errorOutput += fmt.Sprintln(line)
	}

	// print error log if enabled
	if lc.Config.Error.Record && errorOutput != "" {
		// command details
		cmdDetails := fmt.Sprintf(`
-----------------COMMAND------------------
cmd:%s, task:%s, template:%s
-----------------BEGINING-----------------
%s
-------------------END--------------------


`,
			cmd.Command, task.Name, task.Template, errorOutput)
		err = fileUtils.AppendFile(lc.Config.Error.Dir, lc.Config.Error.Filename, []byte(cmdDetails))
		if err != nil {
			return nil
		}
	}

	// print the output to the file, if enabled
	if cmd.Output.Filename != "" {
		output := ""
		for _, line := range status.Stdout {
			output += fmt.Sprintln(line)
		}

		if cmd.Output.Dir == "" {
			cmd.Output.Dir = defaultDir
		}
		if errorOutput != "" {
			filename := fmt.Sprintf("%s_err", cmd.Output.Filename)
			if cmd.Output.Append {
				err = fileUtils.AppendFile(cmd.Output.Dir, filename, []byte(errorOutput))
				if err != nil {
					return nil
				}
			} else {
				err = fileUtils.WriteFile(cmd.Output.Dir, filename, []byte(errorOutput))
				if err != nil {
					return nil
				}
			}
		}

		if output != "" {
			if cmd.Output.Append {
				return fileUtils.AppendFile(cmd.Output.Dir, cmd.Output.Filename, []byte(output))
			} else {
				return fileUtils.WriteFile(cmd.Output.Dir, cmd.Output.Filename, []byte(output))
			}
		}
	}
	return nil
}

func getEnv(env map[string]string) []string {
	envs := make([]string, 0)
	for k, v := range env {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}
