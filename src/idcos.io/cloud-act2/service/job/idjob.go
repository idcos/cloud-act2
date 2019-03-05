//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"idcos.io/cloud-act2/utils/arrays"
	"idcos.io/cloud-act2/utils/generator"

	"idcos.io/cloud-act2/utils/promise"

	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
)

/**
根据主机id列表执行作业
*/
func ProcessAndExecByID(user string, param common.ConfJobIDExecParam) (jobRecordId string, err error) {
	logger := getLogger()
	hostInfos, err := findHostInfoByIDs(param.EntityIDs, param.Provider)

	if err != nil {
		logger.Error("find host info error; ", "error", err.Error(), "provider", param.Provider, "entityIDs", fmt.Sprintf("%q", param.EntityIDs))
		return
	}

	if len(hostInfos) == 0 {
		logger.Info("host info list is empty", "ids", fmt.Sprintf("%q", param.EntityIDs), "provider", param.Provider)
		err = errors.New(fmt.Sprintf("Act2Master not find host by entityIds %q", param.EntityIDs))
		return
	}

	hostInfos = UniqueHostInfo(hostInfos)

	// TODO: 如果数据量不一致，应该做一些进一步的处理
	if len(hostInfos) != len(param.EntityIDs) {
		logger.Warn("not found all host info")
	}

	var jobRecord model.JobRecord
	jobRecord.ID = generator.GenUUID()
	jobRecord.User = user

	if err = saveJobRecord(param.Provider, hostInfos, param.ExecParam, &jobRecord, param.Callback, param.ExecuteID, define.Doing); err != nil {
		return
	}

	promise.NewGoPromise(func(chan struct{}) {
		IDExecAsync(param, jobRecord, hostInfos)
	}, nil)

	return jobRecord.ID, nil
}

func IDExecAsync(param common.ConfJobIDExecParam, jobRecord model.JobRecord, hostInfos []common.HostInfo) (err error) {
	logger := getLogger()

	idcHostMap, idcs := hostInfoMapper(hostInfos)
	if err != nil {
		return
	}

	options := map[string]interface{}{
		"password":   param.ExecParam.Password,
		"username":   param.ExecParam.RunAs,
		"scriptType": param.ExecParam.ScriptType,
	}

	if define.FileModule == param.ExecParam.Pattern && define.UrlType == param.ExecParam.ScriptType {
		var urls []string
		if err = json.Unmarshal([]byte(param.ExecParam.Script), &urls); err != nil {
			return
		}

		for _, url := range urls {
			task, taskSaveErr := saveTask(url, param.ExecParam.Params, jobRecord.ID, param.ExecParam.Pattern, options)
			if taskSaveErr != nil {
				logger.Error("fail to save JobTask", "error", taskSaveErr.Error())
				return taskSaveErr
			}

			IDExecPattIdc(idcs, idcHostMap, param, hostInfos, task, jobRecord.Timeout)
		}
	} else {
		task, taskSaveErr := saveTask(param.ExecParam.Script, param.ExecParam.Params, jobRecord.ID, param.ExecParam.Pattern, options)
		if taskSaveErr != nil {
			logger.Error("fail to save JobTask", "error", taskSaveErr.Error())
			return taskSaveErr
		}

		IDExecPattIdc(idcs, idcHostMap, param, hostInfos, task, jobRecord.Timeout)
	}
	return
}

func IDExecPattIdc(idcs []string, idcHostMap map[string][]common.HostInfo, param common.ConfJobIDExecParam, hostInfos []common.HostInfo, jobTask model.JobTask, timeout int) (err error) {
	logger := getLogger()

	for _, idc := range idcs {
		idcHosts := idcHostMap[idc]

		if len(idcHosts) == 0 {
			logger.Warn("idc has empty hosts, continue it ", "idc", idc)
			continue
		}

		// 处理JobExecParam
		var execHosts []common.ExecHost

		for _, hostInfo := range idcHosts {
			execHosts = append(execHosts, common.ExecHost{
				HostID:   hostInfo.HostID,
				HostIP:   hostInfo.HostIP,
				EntityID: hostInfo.EntityID,
				IdcName:  hostInfo.IdcName,
				ProxyID:  hostInfo.ProxyID,
				Encoding: getEncodingByVersionAndOsType(hostInfo.MinionVersion, hostInfo.OsType),
			})
		}

		jobExecParam := common.ProxyJobExecParam{
			Provider:  param.Provider,
			Callback:  getCallbackUrl(),
			ExecHosts: execHosts,
			TaskID:    jobTask.ID,
			ExecParam: param.ExecParam,
		}

		promise.NewGoPromise(func(chan struct{}) {
			JobExec(jobExecParam, idc, jobTask, timeout)
		}, nil)

	}
	return
}

/**
list to server map
*/
func hostInfoMapper(hostInfos []common.HostInfo) (idcHostMap map[string][]common.HostInfo, idcs []string) {

	idcHostMap = make(map[string][]common.HostInfo)
	for _, hostInfo := range hostInfos {
		idcHosts, _ := idcHostMap[hostInfo.IdcName]
		idcHosts = append(idcHosts, common.HostInfo{
			HostID:        hostInfo.HostID,
			HostIP:        hostInfo.HostIP,
			EntityID:      hostInfo.EntityID,
			IdcName:       hostInfo.IdcName,
			IdcID:         hostInfo.IdcID,
			ProxyID:       hostInfo.ProxyID,
			MinionVersion: hostInfo.MinionVersion,
		})

		idcHostMap[hostInfo.IdcName] = idcHosts

		if !arrays.Contain(hostInfo.IdcName, idcs) {
			idcs = append(idcs, hostInfo.IdcName)
		}
	}
	return
}
