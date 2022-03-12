package jenkins_provider

import (
	"errors"
	"strings"
	"time"

	funcUtils "github.com/jkandasa/autoeasy/pkg/utils/function"
	jenkinsProviderTY "github.com/jkandasa/autoeasy/plugin/provider/jenkins/types"
	"github.com/jkandasa/jenkinsctl/pkg/types/jenkins"
	"go.uber.org/zap"
)

// executes the build job
func (j *Jenkins) build(cfg *jenkinsProviderTY.ProviderConfig) (interface{}, error) {
	// get build data slice
	buildDataSlice, err := cfg.GetBuildData()
	if err != nil {
		return nil, err
	}

	responses := make([]interface{}, 0)

	for index := range buildDataSlice {
		buildData := buildDataSlice[index]
		response, err := j.buildSingle(&cfg.Config, &buildData)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}

	if len(buildDataSlice) == 1 {
		return responses[0], nil
	}
	return responses, nil

}

// executes the build
func (j *Jenkins) buildSingle(taskCfg *jenkinsProviderTY.TaskConfig, buildData *jenkinsProviderTY.BuildData) (interface{}, error) {
	retryCount := 1
	if taskCfg.RetryCount > 0 {
		retryCount = taskCfg.RetryCount
	}

	if buildData.Limit == 0 {
		buildData.Limit = 5
	}

	// retry loop
	for {
		retryCount--

		zap.L().Debug("invoking a job", zap.String("jobName", buildData.JobName))
		queueID, err := j.Client.Build(buildData.JobName, buildData.Parameters)
		if err != nil {
			zap.L().Error("error on invoking a job", zap.String("jobName", buildData.JobName), zap.Error(err))
			return nil, err
		}

		if !taskCfg.WaitForCompletion {
			return map[string]interface{}{"queueId": queueID}, nil
		}

		zap.L().Debug("invoked a job", zap.String("jobName", buildData.JobName), zap.Int64("queueId", queueID))

		buildNumber := int(0)
		isSuccess := false
		var response *jenkins.BuildResponse

		// verify the job completion status
		verifyFunc := func() (bool, error) {
			if buildNumber == 0 {
				buildResponse, err := j.Client.GetBuildByQueueID(buildData.JobName, queueID, buildData.Limit)
				if err != nil {
					zap.L().Error("error on getting build by queue id", zap.String("jobName", buildData.JobName), zap.Int64("queueId", queueID), zap.Error(err))
					return false, err
				}
				response = buildResponse
			} else {
				buildResponse, err := j.Client.GetBuild(buildData.JobName, buildNumber, false)
				if err != nil {
					zap.L().Error("error on getting build by build number", zap.String("jobName", buildData.JobName), zap.Int64("queueId", queueID), zap.Int("buildNumber", buildNumber), zap.Error(err))
					return false, err
				}
				response = buildResponse
			}

			if response == nil {
				return false, nil
			}

			if buildNumber == 0 {
				buildNumber = int(response.Number)
			}

			if response.IsRunning {
				return false, nil
			}

			if strings.EqualFold(response.Result, "SUCCESS") {
				isSuccess = true
			}

			return true, nil
		}

		// wait for completion of the job
		err = funcUtils.ExecuteWithTimeout(verifyFunc, taskCfg.Timeout, time.Second*10)
		if err != nil {
			return nil, err
		}

		if buildNumber != 0 {
			zap.L().Debug("job status", zap.String("jobName", buildData.JobName), zap.Int("buildNumber", buildNumber), zap.String("result", response.Result), zap.String("timeTaken", response.Duration.String()))
		}

		// return immediately, if the job completed successfully
		if isSuccess {
			return response, nil
		}

		if retryCount == 0 {
			return nil, errors.New("reached maximum retry count, no success job")
		}
	}
}
