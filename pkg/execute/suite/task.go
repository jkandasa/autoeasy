package execute

import (
	"fmt"

	dataRepoSVC "github.com/jkandasa/autoeasy/pkg/service/data_repository"
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
	data, err := provider.Execute(task)
	if err != nil {
		zap.L().Error("error on a task", zap.String("taskName", task.Name), zap.String("template", task.Template), zap.Error(err))
		switch task.OnFailure {
		case templateTY.OnFailureContinue:
			return nil

		case templateTY.OnFailureExit:
			return err

		case templateTY.OnFailureRepeat:
			data, err = provider.Execute(task)
			if err != nil {
				return err
			}
		}
	}

	if len(task.Store) > 0 {
		for _, store := range task.Store {
			if data == nil {
				dataRepoSVC.Add(store.Key, "")
			} else {
				dataRepoSVC.AddWithStore(store, data)
			}
		}
	}
	return nil
}
