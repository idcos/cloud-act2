//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
//Package execute 执行相关
package execute

import (
	"idcos.io/cloud-act2/define"
	"strings"

	"idcos.io/cloud-act2/config"
	channelCommon "idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/common/service"
	"idcos.io/cloud-act2/service/common"
)

//Job 执行脚本
func Job(jobExecParam common.ProxyJobExecParam) (err error) {
	logger := getLogger()

	executeParam := channelCommon.ExecScriptParam{
		Pattern:        jobExecParam.ExecParam.Pattern,
		Script:         jobExecParam.ExecParam.Script,
		ScriptType:     jobExecParam.ExecParam.ScriptType,
		Params:         jobExecParam.ExecParam.Params,
		RunAs:          strings.TrimSpace(jobExecParam.ExecParam.RunAs),
		Password:       jobExecParam.ExecParam.Password,
		Timeout:        jobExecParam.ExecParam.Timeout,
		Env:            jobExecParam.ExecParam.Env,
		ExtendData:     jobExecParam.ExecParam.ExtendData,
		RealtimeOutput: jobExecParam.ExecParam.RealTimeOutput,
	}

	var callback string
	if jobExecParam.Callback != "" {
		callback = jobExecParam.Callback
	} else {
		callback = strings.TrimRight(config.Conf.Act2.ClusterServer, "/") + define.CallbackUri
	}

	returnContext := &channelCommon.ReturnContext{
		TaskID:      jobExecParam.TaskID,
		CallbackURL: callback,
		CloseChan:   make(chan int),
	}

	logger.Debug("start execute", "JobTaskID", jobExecParam.TaskID)
	service.Execute(jobExecParam.Provider, jobExecParam.TaskID, jobExecParam.ExecHosts, executeParam, returnContext)

	// logger.Debug("execute done", "JobTaskID", jobExecParam.TaskID)
	return
}
