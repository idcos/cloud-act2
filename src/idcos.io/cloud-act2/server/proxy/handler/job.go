//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/utils/generator"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"idcos.io/cloud-act2/crypto"

	"github.com/pkg/errors"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"

	"strings"

	"idcos.io/cloud-act2/server/common"
	serviceCommon "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/execute"
	"idcos.io/cloud-act2/service/job"
	"idcos.io/cloud-act2/utils"
)

var (
	// 同步情况下的任务处理方式
	syncTasks     = make(map[string]chan serviceCommon.ProxyCallBackResult)
	syncTaskMutex sync.Mutex
)

func isValidProxyParam(proxyParam serviceCommon.ProxyJobExecParam) bool {
	if proxyParam.ExecParam.Pattern == define.ScriptModule {
		if args, ok := proxyParam.ExecParam.Params["args"]; ok {
			if _, ok := args.(string); !ok {
				return false
			}
		}
	} else if proxyParam.ExecParam.Pattern == define.FileModule {
		if target, ok := proxyParam.ExecParam.Params["target"]; ok {
			if _, ok := target.(string); !ok {
				return false
			}
		} else {
			return false
		}

		if fileName, ok := proxyParam.ExecParam.Params["fileName"]; ok {
			if _, ok := fileName.(string); !ok {
				return false
			}
		} else {
			return false
		}

	}

	return true
}

//Execute execute
func Execute(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	proxyExecParam := serviceCommon.ProxyJobExecParam{}
	err = json.Unmarshal(bytes, &proxyExecParam)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if len(proxyExecParam.ExecParam.Password) > 0 {
		client := crypto.GetClient()
		proxyExecParam.ExecParam.Password = client.Encode(proxyExecParam.ExecParam.Password)
	}

	logger.Debug("api /execute", "data", fmt.Sprintf("%#v", proxyExecParam))

	// 参数矫正，即填充默认值
	var hosts []serviceCommon.ExecHost
	for _, host := range proxyExecParam.ExecHosts {
		if strings.TrimSpace(host.Encoding) == "" {
			host.Encoding = "UTF-8"
		}
		hosts = append(hosts, host)
	}
	proxyExecParam.ExecHosts = hosts

	err = utils.Validate.Struct(&proxyExecParam)
	if err != nil {
		logger.Error("validate proxy exec param", "error", err)
		common.HandleError(w, err)
		return
	}

	if !isValidProxyParam(proxyExecParam) {
		common.HandleError(w, errors.New("invalid proxy param given"))
		return
	}

	err = execute.Job(proxyExecParam)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	common.HandleSuccess(w, nil)
}

func SyncCallback(w http.ResponseWriter, r *http.Request) {
	proxyResult := serviceCommon.ProxyCallBackResult{}
	err := common.ReadJSONRequest(r, &proxyResult)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	syncTaskMutex.Lock()
	proxyChan, ok := syncTasks[proxyResult.Content.TaskID]
	syncTaskMutex.Unlock()
	if ok {
		proxyChan <- proxyResult
	}
}

//SyncExecute sync execute
func SyncExecute(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	syncJobExecParam := serviceCommon.SyncJobExecParam{}
	err = json.Unmarshal(bytes, &syncJobExecParam)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if len(syncJobExecParam.ExecParam.Password) > 0 {
		client := crypto.GetClient()
		syncJobExecParam.ExecParam.Password = client.Encode(syncJobExecParam.ExecParam.Password)
	}

	logger.Debug("api /sync/execute", "data", fmt.Sprintf("%#v", syncJobExecParam))

	err = utils.Validate.Struct(&syncJobExecParam)
	if err != nil {
		logger.Error("validate proxy exec param", "error", err)
		common.HandleError(w, err)
		return
	}

	var callbcakURL string
	port := strings.TrimLeft(config.Conf.Port, ":")
	if port == "" {
		callbcakURL = "http://127.0.0.1:5555/api/v1/job/sync/callback"
	} else {
		callbcakURL = fmt.Sprintf("http://127.0.0.1:%s/api/v1/job/sync/callback", port)
	}

	uuid := generator.GenUUID()

	proxyExecParam := serviceCommon.ProxyJobExecParam{
		ExecHosts: syncJobExecParam.ExecHosts,
		ExecParam: syncJobExecParam.ExecParam,
		Provider:  syncJobExecParam.Provider,
		Callback:  callbcakURL,
		TaskID:    uuid,
	}

	// 参数矫正，即填充默认值
	var hosts []serviceCommon.ExecHost
	for _, host := range proxyExecParam.ExecHosts {
		if strings.TrimSpace(host.Encoding) == "" {
			host.Encoding = "UTF-8"
		}
		hosts = append(hosts, host)
	}
	proxyExecParam.ExecHosts = hosts

	proxyChan := make(chan serviceCommon.ProxyCallBackResult)
	syncTaskMutex.Lock()
	syncTasks[uuid] = proxyChan
	syncTaskMutex.Unlock()

	err = execute.Job(proxyExecParam)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// 等待执行结果
	// timeout := syncJobExecParam.ExecParam.Timeout + 2

	common.HandleJSONResponse(w, getSyncExecuteResult(proxyChan, syncJobExecParam, uuid))
	//common.handleJSONResponse(w, result)
}

func getSyncExecuteResult(proxyChan chan serviceCommon.ProxyCallBackResult, syncJobExecParam serviceCommon.SyncJobExecParam, uuid string) []byte {
	const timeFormat = "2006-01-02 15:04:05"
	startTime := time.Now().Format(timeFormat)

	var result serviceCommon.ProxyCallBackResult
	select {
	case result = <-proxyChan:
	case <-time.After(time.Duration(syncJobExecParam.ExecParam.Timeout) * time.Second):
		result = serviceCommon.ProxyCallBackResult{
			Status:  define.Fail,
			Message: "timeout",
			Content: serviceCommon.ProxyResultContent{
				TaskID: uuid,
			},
		}
	}
	result.Content.StartTime = startTime
	result.Content.EndTime = time.Now().Format(timeFormat)

	v, _ := json.Marshal(result)
	return v
}

//RealTime 接收实时输出结果
func RealTime(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if config.Conf.CacheType != define.Redis {
		common.HandleError(w, errors.New("not support when no redis"))
		return
	}

	logger.Trace("get real time result", "body", string(bytes))

	realTimeForm := serviceCommon.RealTimeForm{}
	err = json.Unmarshal(bytes, &realTimeForm)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	err = job.RealTimeToRedis(realTimeForm)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, "process success")
}

type statResult struct {
	Doing   int64 `json:"doing"`
	Success int64 `json:"success"`
	Fail    int64 `json:"fail"`
	Timeout int64 `json:"timeout"`
}

//Stat
func Stat(w http.ResponseWriter, r *http.Request) {
	doing, success, fail, timeout, err := job.ProxyStat()
	if err != nil {
		common.HandleError(w, err)
		return
	}

	result := statResult{
		Doing:   doing,
		Success: success,
		Fail:    fail,
		Timeout: timeout,
	}

	common.CommonHandleSuccess(w, map[string]interface{}{"job": result})
}
