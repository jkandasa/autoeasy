package execute

import (
	"fmt"

	providerSVC "github.com/jkandasa/autoeasy/pkg/service/provider"
	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	"go.uber.org/zap"
)

func run(task *templateTY.Task) error {
	providerName := task.Provider

	// get provider instance
	provider := providerSVC.GetProvider(providerName)
	if provider == nil {
		return fmt.Errorf("provider not available. providerName:[%s]", providerName)
	}

	// execute task
	err := provider.Execute(task)
	if err != nil {
		zap.L().Error("error on a task", zap.String("taskName", task.Name), zap.String("template", task.Template), zap.Error(err))
		switch task.OnFailure {
		case templateTY.OnFailureContinue:
			return nil

		case templateTY.OnFailureExit:
			return err

		case templateTY.OnFailureRepeat:
			err = provider.Execute(task)
			return err
		}
	}

	return nil
}
