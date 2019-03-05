//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"fmt"
	"idcos.io/cloud-act2/utils/dataexchange"
	"time"

	"github.com/pkg/errors"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/common/report"
)

type (
	Callback interface {
		process(callBackResult common.ProxyCallBackResult) error
	}
	TimeoutCallback struct{}
	FailCallback struct{}
	SuccessCallback struct{}
)

var taskIsDoneError = errors.New("update task fail, maybe the task is done")

func (callback *TimeoutCallback) process(callBackResult common.ProxyCallBackResult) (err error) {
	logger := getLogger()

	taskID := callBackResult.Content.TaskID

	jobTask, query, args, err := getTaskAndQueryWhere(callBackResult.Content.MasterSend, taskID, callBackResult.Content.ProxyID)
	if err != nil {
		return err
	}

	var hostResults []model.HostResult
	if err := model.GetDb().Where(query, args...).Find(&hostResults).Error; err != nil {
		logger.Error("fail to find host results", "taskID", taskID, "error", err)
		return err
	}
	logger.Debug("-------- find taskID match host results;", "taskID", taskID, "host results", dataexchange.ToJsonString(hostResults))

	err = batchUpdateHostResultStatus(hostResults, define.Done, define.Timeout, define.Timeout)
	if err != nil {
		logger.Error("batch update host result", "error", err)
		return
	}

	return isTaskDone(jobTask.ID, jobTask.RecordID, callBackResult.Content.MasterSend, callBackResult.Content.ProxyID)
}

func getTaskAndQueryWhere(masterSend bool, taskID, proxyID string) (jobTask model.JobTask, query string, args []interface{}, err error) {
	jobTask, err = findTaskAndCheck(taskID)
	if err != nil {
		return jobTask, "", nil, err
	}

	if masterSend {
		//更新task状态
		if err = changeStatus(taskID, define.Timeout); err != nil {
			return jobTask, "", nil, err
		}

		query = "task_id = ? and execute_status = ?"
		args = []interface{}{taskID, define.Doing}
	} else {
		query = "task_id = ? and execute_status = ? and proxy_id = ?"
		args = []interface{}{taskID, define.Doing, proxyID}
	}

	return jobTask, query, args, nil
}

func reportStatisticsWithProxy(proxyID string, status string) {
	logger := getLogger()

	key := getMasterProxyRedisPreKey(proxyID)
	recorder := report.GetRecorder(key)

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

	if err != nil {
		logger.Error("report statistics fail", "error", err)
	}
}

func getMasterProxyRedisPreKey(proxyID string) string {
	return "master|proxy|" + proxyID
}

func findTaskAndCheck(taskID string) (jobTask model.JobTask, err error) {
	logger := getLogger()

	jobTask = model.JobTask{}
	if err := jobTask.GetByID(taskID); err != nil {
		logger.Error("fail to find JobTask by id", "taskID", taskID, "error", err)
		return jobTask, err
	}

	if err := checkJobTask(jobTask); err != nil {
		logger.Error(err.Error(), "taskID", taskID)
		return jobTask, err
	}
	return jobTask, nil
}

func changeStatus(taskID, resultStatus string) (err error) {
	logger := getLogger()

	changeNum, err := model.UpdateTaskStatusToDone(taskID, resultStatus)
	if err != nil {
		logger.Error(err.Error(), "taskID", taskID)
		return err
	}
	if changeNum == 0 {
		logger.Info(taskIsDoneError.Error(), "taskID", taskID)
		return taskIsDoneError
	}

	return nil
}

func checkTaskDoneAndChangeTaskResultStatus(taskID string, hostResults []model.HostResult) (done bool, err error) {
	logger := getLogger()

	if checkDoneByHostResults(hostResults) {
		resultStatus := getResultStatusByHostResults(hostResults)
		rows, err := model.UpdateTaskStatusToDone(taskID, resultStatus)
		if err != nil {
			logger.Info("update job task fail", "error", err)
			return false, err
		}
		if rows == 0 {
			logger.Info("update job task fail,maybe this task already done")
		}

		return true, nil
	}

	return false, nil
}
func (callback *FailCallback) process(callBackResult common.ProxyCallBackResult) (err error) {
	logger := getLogger()
	taskID := callBackResult.Content.TaskID

	jobTask, query, args, err := getTaskAndQueryWhere(callBackResult.Content.MasterSend, taskID, callBackResult.Content.ProxyID)
	if err != nil {
		return err
	}

	message := callBackResult.Message
	logger.Debug(fmt.Sprintf("taskID %s association hosts exec fail message ：%s; start to update host result", taskID, callBackResult.Message))

	// 若返回状态为fail  返回体当中不存在host_id，只能根据job_record_id查询
	var hostResults []model.HostResult
	if err := model.GetDb().Where(query, args...).Find(&hostResults).Error; err != nil {
		logger.Error("fail to find host results", "taskID", taskID, "error", err)
		return err
	}
	logger.Debug("-------- find taskID match host results;", "taskID", taskID, "host results", dataexchange.ToJsonString(hostResults))

	batchUpdateHostResultStatus(hostResults, define.Done, define.Fail, message)

	return isTaskDone(taskID, jobTask.RecordID, callBackResult.Content.MasterSend, callBackResult.Content.ProxyID)
}

func (callback *SuccessCallback) process(callBackResult common.ProxyCallBackResult) (err error) {
	logger := getLogger()

	taskID := callBackResult.Content.TaskID
	jobTask, err := findTaskAndCheck(taskID)
	if err != nil {
		return err
	}

	// 若主机执行结果成功，则更新对应主机的执行结果记录
	var hostIDs []string
	var hostResultMap = make(map[string]common.ProxyCallbackHostResult)

	for _, hostResult := range callBackResult.Content.HostResults {
		hostIDs = append(hostIDs, hostResult.HostID)
		_, ok := hostResultMap[hostResult.HostID]
		if !ok {
			hostResultMap[hostResult.HostID] = hostResult
		}
	}

	var hostResults []*model.HostResult
	if err := model.GetDb().Where("host_id in (?) and task_id = ? ", hostIDs, taskID).Find(&hostResults).Error; err != nil {
		logger.Error("fail to find host results by IDs; ", "error", err)
		return err
	}

	for _, hostResult := range hostResults {
		result, ok := hostResultMap[hostResult.HostID]

		if ok {
			hostResult.Stdout = result.Stdout
			hostResult.Stderr = result.Stderr
			hostResult.ExecuteStatus = define.Done
			hostResult.ResultStatus = result.Status
			hostResult.EndTime = time.Now()
			hostResult.Message = result.Message
		}
	}

	logger.Debug("start to batchUpdateHostResults", "hostResults", dataexchange.ToJsonString(hostResults))
	if err = batchUpdateHostResults(hostResults); err != nil {
		return
	}

	logger.Debug("start to isUpdateJobRecord")
	return isTaskDone(taskID, jobTask.RecordID, false, callBackResult.Content.ProxyID)
}

func checkJobTask(jobTask model.JobTask) (err error) {
	if define.Done == jobTask.ExecuteStatus {
		return errors.New(fmt.Sprintf("jobTask has been done, %s", dataexchange.ToJsonString(jobTask)))
	}
	return
}

func isTaskDone(taskID string, recordID string, masterSend bool, proxyID string) (err error) {
	logger := getLogger()

	hostResults, err := model.FindHostResultByTaskID(taskID)
	if err != nil {
		logger.Error("query host result fail", "error", err)
		return
	}

	if masterSend {
		err = isAllProxyDone(taskID, hostResults)
	} else {
		err = isProxyDone(taskID, proxyID, filterHostResultsByProxyID(hostResults, proxyID))
	}
	if err != nil {
		return err
	}

	done, err := checkTaskDoneAndChangeTaskResultStatus(taskID, hostResults)
	if err != nil {
		return
	}

	if done {
		return isUpdateJobRecord(recordID)
	}
	return nil
}

func filterHostResultsByProxyID(hostResults []model.HostResult, proxyID string) (filterHostResults []model.HostResult) {
	filterHostResults = make([]model.HostResult, 0, len(hostResults))
	for _, hostResult := range hostResults {
		if hostResult.ProxyID == proxyID {
			filterHostResults = append(filterHostResults, hostResult)
		}
	}
	return
}

func isProxyDone(taskID, proxyID string, hostResults []model.HostResult) (err error) {
	logger := getLogger()

	taskProxy, err := model.FindTaskProxyByTaskIDAndProxyID(taskID, proxyID)
	if err != nil {
		logger.Error("query task proxy fail", "error", err)
		return
	}

	if checkProxyDoneAndChangeResultStatus(hostResults, &taskProxy) {

		//记录报表
		reportStatisticsWithProxy(taskProxy.ProxyID, taskProxy.ResultStatus)

		if err = taskProxy.Save(); err != nil {
			logger.Error("save task proxy fail", "taskProxyID", taskProxy.ID)
			return err
		}
	}
	return nil
}

func isAllProxyDone(taskID string, hostResults []model.HostResult) (err error) {
	logger := getLogger()

	taskProxies, err := model.FindTaskProxyByTaskID(taskID)
	if err != nil {
		logger.Error("query task proxy fail", "error", err)
		return
	}

	hostResultMap := hostResultGroupByProxyID(hostResults)
	updateProxies := make([]model.JobTaskProxy, 0, len(taskProxies))
	for _, taskProxy := range taskProxies {
		hostResultList, ok := hostResultMap[taskProxy.ProxyID]
		if !ok {
			logger.Error("abnormal data,not found hostResult", "proxyID", taskProxy.ProxyID)
			continue
		}

		if checkProxyDoneAndChangeResultStatus(hostResultList, &taskProxy) {
			updateProxies = append(updateProxies, taskProxy)

			//记录报表
			reportStatisticsWithProxy(taskProxy.ProxyID, taskProxy.ResultStatus)
		}
	}

	if len(updateProxies) > 0 {
		if err = updateTaskProxy(updateProxies); err != nil {
			return err
		}

	}
	return nil
}

func checkProxyDoneAndChangeResultStatus(hostResults []model.HostResult, taskProxy *model.JobTaskProxy) (done bool) {
	done = checkDoneByHostResults(hostResults)
	if done {
		resultStatus := getResultStatusByHostResults(hostResults)
		taskProxy.ExecuteStatus = define.Done
		taskProxy.ResultStatus = resultStatus
		taskProxy.EndTime = time.Now()
	}
	return
}

func hostResultGroupByProxyID(hostResults []model.HostResult) (groupMap map[string][]model.HostResult) {
	groupMap = make(map[string][]model.HostResult)
	for _, hostResult := range hostResults {
		hostResultList, ok := groupMap[hostResult.ProxyID]
		if !ok {
			hostResultList = make([]model.HostResult, 0, 10)
		}
		hostResultList = append(hostResultList, hostResult)
		groupMap[hostResult.ProxyID] = hostResultList
	}
	return
}

func checkDoneByHostResults(hostResults []model.HostResult) bool {
	for _, hostResult := range hostResults {
		if hostResult.ExecuteStatus != define.Done {
			return false
		}
	}
	return true
}

func getResultStatusByHostResults(hostResults []model.HostResult) string {
	logger := getLogger()

	successNum := 0
	failNum := 0
	timeoutNum := 0

	for _, hostResult := range hostResults {
		switch hostResult.ResultStatus {
		case define.Success:
			successNum++
		case define.Fail:
			failNum++
		case define.Timeout:
			timeoutNum++
		default:
			logger.Error("abnormal data,unknown resultStatus", "status", hostResult.ResultStatus)
		}
	}

	if timeoutNum == len(hostResults) {
		return define.Timeout
	}

	if successNum == len(hostResults) {
		return define.Success
	}
	return define.Fail
}
