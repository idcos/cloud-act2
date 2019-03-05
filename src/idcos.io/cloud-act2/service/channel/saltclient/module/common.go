//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import (
	"github.com/RussellLuo/timingwheel"
	"github.com/hashicorp/go-hclog"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/service/channel/common"
	serviceCommon "idcos.io/cloud-act2/service/common"
	wheel "idcos.io/cloud-act2/timingwheel"
	"idcos.io/cloud-act2/utils/stringutil"
)

func listenTimeout(timeout int, partitionResult *common.PartitionResult) (timer *timingwheel.Timer) {
	logger := getLogger()

	logger.Debug("start listen job timeout", "timeout", hclog.Fmt("%d s", timeout))
	return wheel.AddTask(timeout, func() {
		select {
		case <-partitionResult.CloseChan: //判断一下此任务是否已结束
		default:
			result := common.Result{
				Status: define.Timeout,
			}

			partitionResult.ResultChan <- result
			partitionResult.Close()
			//common.SendResultAndClose(result, returnContext, true)
			//returnContext.Close()
		}
	})
}

func extractExecHostsToHosts(execHosts []serviceCommon.ExecHost) []string {
	var hosts []string
	for _, execHost := range execHosts {
		hosts = append(hosts, execHost.EntityID)
	}
	return hosts
}

func extractEntityIDToHostIDMap(execHosts []serviceCommon.ExecHost) map[string]string {
	hostMinion := map[string]string{}
	for _, execHost := range execHosts {
		hostMinion[execHost.EntityID] = execHost.HostID
	}
	return hostMinion
}

func filterResultAndAdd(jobRecordID string, minionResults []common.MinionResult, timeout int) (filterResults []common.MinionResult) {
	//过滤已发送主机列表
	hosts := make([]string, 0, 1000)
	for _, minionResult := range minionResults {
		hosts = append(hosts, minionResult.HostID)
	}

	filterHosts := common.FilterHosts(jobRecordID, hosts)

	filterResults = make([]common.MinionResult, 0, 1000)
	for _, minionResult := range minionResults {
		if stringutil.StringScliceContains(filterHosts, minionResult.HostID) {
			filterResults = append(filterResults, minionResult)
		}
	}

	common.AddRecord(jobRecordID, filterHosts, timeout)
	return
}

func removeResultRecord(jobRecordID string) {
	common.RemoveRecord(jobRecordID)
}
