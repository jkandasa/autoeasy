package function

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"go.uber.org/zap"
)

const (
	DefaultTimeout              = time.Minute * 3
	DefaultInterval             = time.Second * 5
	DefaultExpectedSuccessCount = 4
)

func ExecuteWithDefaultTimeoutAndContinuesSuccessCount(executeFunc func() (bool, error)) error {
	return ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, DefaultTimeout, DefaultInterval, DefaultExpectedSuccessCount)
}

func ExecuteWithTimeout(executeFunc func() (bool, error), timeout time.Duration, interval time.Duration) error {
	return ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc, timeout, interval, 1)
}

func ExecuteWithTimeoutAndContinuesSuccessCount(executeFunc func() (bool, error), timeout time.Duration, interval time.Duration, expectedSuccessCount int) error {
	if executeFunc == nil {
		return errors.New("execute func can not be nil")
	}

	if expectedSuccessCount < 0 {
		expectedSuccessCount = 0
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(executeFunc).Pointer()).Name()
	zap.L().Debug("polling started", zap.Any("func", funcName), zap.String("timeout", timeout.String()), zap.String("scanInterval", interval.String()), zap.Int("expectedSuccessCount", expectedSuccessCount))
	startTime := time.Now()
	defer func() {
		zap.L().Debug("polling completed", zap.Any("func", funcName), zap.String("timeTaken", time.Since(startTime).String()))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	successCount := 0
	for {
		select {
		case <-ticker.C:
			success, err := executeFunc()
			if err != nil {
				return err
			} else if success {
				successCount++
				if successCount >= expectedSuccessCount {
					return nil
				}
			} else {
				successCount = 0
			}

		case <-ctx.Done():
			return fmt.Errorf("reached timeout: %s", timeout.String())
		}

	}
}
