//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/service/common/report"
	"idcos.io/cloud-act2/utils/httputil"
	"idcos.io/cloud-act2/utils/promise"
	"time"

	slave "github.com/dgrr/GoSlaves"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/job"
)

var globalResultChan chan Result
var pool *slave.SlavePool

//Load 初始化
func Load() (err error) {
	//初始化线程池
	pool = &slave.SlavePool{
		Work: func(obj interface{}) {
			result := obj.(Result)
			processResult(result)
		},
	}
	pool.Open()

	globalResultChan = make(chan Result, 10)
	promise.NewGoPromise(func(chan struct{}) {
		listenResult()
	}, nil)
	return
}

func listenResult() {
	for {
		result := <-globalResultChan
		pool.Serve(result)
	}
}

//SendResult 发送结果
func SendResultAndClose(result Result, returnContext *ReturnContext, close bool) {
	returnContext.mutex.Lock()
	defer returnContext.mutex.Unlock()

	if !returnContext.CheckClose() {
		result.CallbackURL = returnContext.CallbackURL
		result.TaskID = returnContext.TaskID
		globalResultChan <- result
		if close {
			returnContext.Close()
			//记录报表
			reportStatistics(result.Status)
		}
	}
}

func reportStatistics(status string) {
	logger := getLogger()
	recorder := report.GetRecorder(config.ComData.SN)

	var err error
	switch status {
	case define.Success:
		err = recorder.AddSuccess()
	case define.Fail:
		err = recorder.AddFail()
	case define.Timeout:
		err = recorder.AddTimeout()
	default:
		logger.Error("report statistics fail,unknown job status", "status", status)
	}

	err = recorder.SubDoing()
	if err != nil {
		logger.Error("report statistics fail", "error", err)
	}
}

func processResult(result Result) {
	logger := getLogger()

	//删除实时输出的数据
	job.RemoveJobInfo(result.TaskID)

	callbackHostResults := make([]common.ProxyCallbackHostResult, 0, 1000)
	if len(result.MinionResults) > 0 {
		for _, minionResult := range result.MinionResults {
			callbackHostResult := common.ProxyCallbackHostResult{
				HostID:  minionResult.HostID,
				Status:  minionResult.Status,
				Stdout:  minionResult.Stdout,
				Stderr:  minionResult.Stderr,
				Message: minionResult.Message,
			}
			callbackHostResults = append(callbackHostResults, callbackHostResult)
		}
	}

	callbackResult := common.ProxyCallBackResult{
		Status:  result.Status,
		Message: result.Message,
		Content: common.ProxyResultContent{
			TaskID:      result.TaskID,
			ProxyID:     config.ComData.SN,
			HostResults: callbackHostResults,
		},
	}

	logger.Info("process result callback", "url", result.CallbackURL, "jobTaskID", result.TaskID)

	resultsJSONBytes, err := json.Marshal(callbackResult)
	if err != nil {
		logger.Debug("proxyCallBackResult to json string fail", "result", callbackResult, "error", err)
	}

	logger.Debug("callback act2-master", "callbackResult", string(resultsJSONBytes))
	resp, err := httputil.HttpPost(result.CallbackURL, resultsJSONBytes)
	if err != nil {
		// 添加重试机制，以便后续处理
		if result.RetryCount >= config.Conf.Act2.ResultCallRetry {
			logger.Error("callback url exceed max try count", "url", result.CallbackURL,
				"retry", fmt.Sprintf("%d", result.RetryCount), "taskId", result.TaskID)
			return
		} else {
			logger.Error("callback url fail", "url", result.CallbackURL, "error", err, "taskId", result.TaskID, "retry", result.RetryCount)
			promise.NewGoPromise(func(chan struct{}) {
				time.Sleep(time.Duration(2<<result.RetryCount) * time.Second)
				result.RetryCount += 1
				globalResultChan <- result
			}, nil)
		}

		return
	}
	logger.Info("callback result success", "taskId", result.TaskID)

	logger.Debug("callback result", "body", string(resp), "taskId", result.TaskID)
}
