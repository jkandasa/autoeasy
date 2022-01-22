package execute

import (
	"fmt"

	providerSVC "github.com/jkandasa/autoeasy/pkg/service/provider"
	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	"go.uber.org/zap"
)

func run(action *templateTY.Action) error {
	providerName := action.Provider

	// get provider instance
	provider := providerSVC.GetProvider(providerName)
	if provider == nil {
		return fmt.Errorf("provider not available. providerName:[%s]", providerName)
	}

	// execute action
	err := provider.Execute(action)
	if err != nil {
		zap.L().Error("error on a action", zap.String("actionName", action.Name), zap.String("template", action.Template), zap.Error(err))
		switch action.OnFailure {
		case templateTY.OnFailureContinue:
			return nil

		case templateTY.OnFailureExit:
			return err

		case templateTY.OnFailureRepeat:
			err = provider.Execute(action)
			return err
		}
	}

	return nil
}
