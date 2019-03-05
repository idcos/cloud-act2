//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"errors"
	"fmt"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/generator"
	"idcos.io/cloud-act2/utils/httputil"
	"idcos.io/cloud-act2/utils/stringutil"
	"time"

	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/utils/promise"
	"idcos.io/cloud-act2/webhook"

	"strings"

	"idcos.io/cloud-act2/service/common/report"

	"github.com/jinzhu/gorm"
	"github.com/theplant/batchputs"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
)

/**
根据主机ip列表执行作业
*/

// 将hostInfos进行unique，防止出现一个entityId对应多个ip的情况
func UniqueHostInfo(hostInfos []common.HostInfo) (uniqueHostInfo []common.HostInfo) {
	hostInfoMap := make(map[string]common.HostInfo)
	for _, hostInfo := range hostInfos {
		existhost, ok := hostInfoMap[hostInfo.EntityID]

		if ok {
			hostInfo.HostIP = fmt.Sprintf("%s,%s", hostInfo.HostIP, existhost.HostIP)
		}

		hostInfoMap[hostInfo.EntityID] = hostInfo
	}
	for _, v := range hostInfoMap {
		uniqueHostInfo = append(uniqueHostInfo, v)
	}

	return uniqueHostInfo
}

func saveTask(script string, params map[string]interface{}, recordID string, pattern string, option map[string]interface{}) (task model.JobTask, err error) {
	task = model.JobTask{
		ID:            generator.GenUUID(),
		StartTime:     time.Now(),
		EndTime:       time.Now(),
		ExecuteStatus: define.Doing,
		ResultStatus:  "",
		RecordID:      recordID,
		Pattern:       pattern,
		Params:        dataexchange.ToJsonString(params),
		Script:        script,
		Options:       dataexchange.ToJsonString(option),
	}

	if err = task.Save(); err != nil {
		return
	}

	return
}

const (
	matchTypeEntity = "entity"
	matchTypeIp     = "ip"
)

//ProxyStat proxy的任务记录
func ProxyStat() (doing, success, fail, timeout int64, err error) {
	doing, err = getCountByStatus(define.Doing)
	if err != nil {
		return
	}

	success, err = getCountByStatus(define.Success)
	if err != nil {
		return
	}

	fail, err = getCountByStatus(define.Fail)
	if err != nil {
		return
	}

	timeout, err = getCountByStatus(define.Timeout)
	return
}

func getCountByStatus(status string) (num int64, err error) {
	return report.GetRecorder(config.ComData.SN).GetRecord(status)
}

/**
主机回调结果
*/
func HostResultsCallback(callBackResult common.ProxyCallBackResult) error {

	logger := getLogger()

	logger.Debug("act2 get proxy callback results", "param", dataexchange.ToJsonString(callBackResult))

	// publish record done to redis
	// TODO: should send message to chan, then chan do add done to redis or memory
	//pubsub功能不应该影响主业务
	//client, err := pubsub.GetPubSubClient()
	//if err != nil {
	//	logger.Error("get cache client fail", "error", err)
	//} else {
	//	err = client.SendPublish(define.RecordDone, true)
	//	if err != nil {
	//		logger.Warn("record host done", "error", err)
	//	}
	//}

	var callback Callback
	switch callBackResult.Status {
	case define.Timeout:
		callback = &TimeoutCallback{}
	case define.Fail:
		callback = &FailCallback{}
	case define.Success:
		callback = &SuccessCallback{}
	}

	if err := callback.process(callBackResult); err != nil {
		return err
	}
	return nil
}

func extractHostIdcs(hostInfos []common.HostInfo) []string {
	var idcs []string
	for _, hostInfo := range hostInfos {
		idcs = append(idcs, hostInfo.IdcName)
	}
	return idcs
}

func filterHostsByIdc(idc string, hostInfos []common.HostInfo) (idcHosts []common.HostInfo) {
	for _, hostInfo := range hostInfos {
		if idc != hostInfo.IdcName {
			continue
		}
		idcHosts = append(idcHosts, hostInfo)
	}
	return
}

func checkSshParam(param common.ConfJobIPExecParam) error {
	// check username password osType
	var errMsg []string
	if param.ExecParam.RunAs == "" {
		errMsg = append(errMsg, "username can not be null")
	}

	if param.ExecParam.Password == "" {
		errMsg = append(errMsg, "password can not be null")
	}

	for _, exec := range param.ExecHosts {
		if exec.OsType == "" && !stringutil.StringScliceContains(errMsg, "osType can not be null") {
			errMsg = append(errMsg, "osType can not be null")
		}
		if exec.HostIP == "" && !stringutil.StringScliceContains(errMsg, "ip can not be null") {
			errMsg = append(errMsg, "ip can not be null")
		}
	}

	if len(errMsg) > 0 {
		return errors.New(strings.Join(errMsg, ";"))
	}
	return nil
}

func saveData(hostRows [][]interface{}, hostIPRows [][]interface{}) error {
	logger := getLogger()
	db := model.GetDb()

	// host batch save ; tx is unused for batchput
	if err := batchputs.Put(db.DB(), define.MysqlDriver, model.Act2Host{}.TableName(), "id", model.GetAct2HostColumns(), hostRows); err != nil {
		logger.Error("fail to save Act2Host", "error", err)
		return err
	}

	// hostIP batch save; tx is unused for batchput
	if err := batchputs.Put(db.DB(), define.MysqlDriver, model.Act2HostIP{}.TableName(), "id", model.GetAct2HostIPColumns(), hostIPRows); err != nil {
		logger.Error("fail to save Act2HostIP", "error", err)
		return err
	}
	return nil

}

/**
主机信息
*/
func PullHosts(idc string, entityId string) (err error) {

	idcMap, err := model.GetAllIdcNameMap()
	if err != nil {
		return
	}
	var proxes []model.Act2Proxy
	if entityId != "" {
		var proxy model.Act2Proxy
		err = model.GetDb().Table("act2_proxy").Where("id = ?", entityId).Find(&proxy).Error

		if err != nil {
			return
		}
		proxes = append(proxes, proxy)
	}

	if idc != "" {
		idcInfo, ok := idcMap[idc]
		if !ok {
			err = errors.New("fail to find idc " + idc)
			return
		}

		err = model.GetDb().Table("act2_proxy").Where("idc_id = ?", idcInfo.ID).Find(&proxes).Error

		if err != nil {
			return
		}
	} else {
		err = model.GetDb().Table("act2_proxy").Find(&proxes).Error

		if err != nil {
			return
		}
	}

	for _, proxy := range proxes {
		if define.Running != proxy.Status {
			continue
		}
		server := proxy.Server
		server = server + define.SystemHeartbeatUri

		promise.NewGoPromise(func(chan struct{}) {
			httputil.HttpPut(server, nil)
		}, nil)
	}

	return nil
}

func getCallbackUrl() string {
	return strings.TrimRight(config.Conf.Act2.ClusterServer, "/") + define.CallbackUri
}

//FindJobRecordPage 分页查询作业执行记录
func FindJobRecordPage(pageNo, pageSize int64) (pagination common.Pagination, err error) {

	//count

	count, err := FindRecordResultsCount()

	if err != nil {
		return
	}

	if count == 0 {
		pagination = common.Pagination{
			PageNo:     pageNo,
			PageSize:   pageSize,
			PageCount:  0,
			TotalCount: 0,
		}
		return
	}

	results, err := FindRecordResultsByPage(pageNo, pageSize)
	if err != nil {
		return
	}

	pagination = common.Pagination{
		PageNo:     pageNo,
		PageSize:   pageSize,
		PageCount:  common.GetPageCount(pageSize, count),
		TotalCount: count,
		List:       results,
	}

	return
}

/**
作业执行记录
*/
func saveJobRecord(provider string, hostInfos []common.HostInfo, param common.ExecParam, jobRecord *model.JobRecord,
	callback string, executeID string, executeStatus string) error {

	logger := getLogger()

	var hostIds []string
	for _, hostInfo := range hostInfos {
		hostIds = append(hostIds, hostInfo.HostID)
	}

	jobRecord.StartTime = time.Now()
	jobRecord.EndTime = time.Now()
	jobRecord.ExecuteStatus = executeStatus
	jobRecord.ResultStatus = ""
	jobRecord.Callback = callback
	jobRecord.Provider = provider
	jobRecord.ScriptType = param.ScriptType
	jobRecord.Script = param.Script
	jobRecord.Pattern = param.Pattern
	jobRecord.Timeout = param.Timeout
	jobRecord.Hosts = dataexchange.ToJsonString(hostIds)
	jobRecord.Parameters = dataexchange.ToJsonString(param)
	jobRecord.MasterID = config.ComData.SN
	jobRecord.ExecuteID = executeID

	err := jobRecord.Save()
	if err != nil {
		logger.Error("fail save job record, "+err.Error(), "JobRecord", dataexchange.ToJsonString(jobRecord))
		return errors.New(fmt.Sprintf("fail save jobrecord ,error %s, JobRecord %s", err.Error(), dataexchange.ToJsonString(jobRecord)))
	}

	return nil
}

func batchUpdateHostResultStatus(hostResultInfos []model.HostResult, execStatus string, resultStatus string, message string) error {
	logger := getLogger()

	var hostResults [][]interface{}
	for _, result := range hostResultInfos {
		hostResults = append(hostResults, []interface{}{
			result.ID,
			result.HostID,
			result.TaskID,
			result.HostIP,
			result.StartTime,
			time.Now(),
			execStatus,
			resultStatus,
			result.Stdout,
			result.Stderr,
			message,
		})
	}

	columns := []string{"id", "host_id", "task_id", "host_ip", "start_time", "end_time", "execute_status", "result_status", "stdout", "stderr", "message"}

	if err := batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.HostResult{}.TableName(), "id", columns, hostResults); err != nil {
		logger.Error("fail update host result", "error", err.Error())
		return err
	}

	return nil
}

func batchUpdateTaskStatus(originTasks []model.JobTask, execStatus string, resultStatus string) error {
	logger := getLogger()

	var jobTasks [][]interface{}
	for _, task := range originTasks {
		jobTasks = append(jobTasks, []interface{}{
			task.ID,
			task.RecordID,
			task.StartTime,
			time.Now(),
			execStatus,
			resultStatus,
		})
	}

	columns := []string{"id", "record_id", "start_time", "end_time", "execute_status", "result_status"}

	if err := batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.JobTask{}.TableName(), "id", columns, jobTasks); err != nil {
		logger.Error("fail update job task", "error", err.Error())
		return err
	}

	return nil
}

func batchUpdateHostResults(hostResults []*model.HostResult) error {
	logger := getLogger()

	var hostInfo []model.HostResult

	for _, hostResult := range hostResults {
		hostInfo = append(hostInfo, model.HostResult{
			ID:            hostResult.ID,
			TaskID:        hostResult.TaskID,
			HostID:        hostResult.HostID,
			StartTime:     hostResult.StartTime,
			EndTime:       hostResult.EndTime,
			ExecuteStatus: hostResult.ExecuteStatus,
			ResultStatus:  hostResult.ResultStatus,
			HostIP:        hostResult.HostIP,
			Stdout:        hostResult.Stdout,
			Stderr:        hostResult.Stderr,
			Message:       hostResult.Message,
		})
	}

	if err := updateHostResult(hostInfo); err != nil {
		logger.Error("fail to update HostResults", "error", err, "hostResults", dataexchange.ToJsonString(hostResults))
		return err
	}
	return nil
}

/*
	校验jobRecordId下面所有的主机结果是否都执行完毕
*/
func isAllDone(tasks []model.JobTask) (string, error) {

	resultStatusMap := make(map[string]common.NullStruct)

	isAllDone := true
	for _, task := range tasks {

		if define.Done != task.ExecuteStatus {
			isAllDone = false
			break
		}

		_, ok := resultStatusMap[task.ResultStatus]
		if !ok {
			resultStatusMap[task.ResultStatus] = common.NullStruct{}
		}
	}

	if !isAllDone {
		return "", nil
	}
	return checkResultStatus(resultStatusMap), nil
}

/**
是否更新JobRecord

校验JobRecordId下面所有的主机是否执行结束，若结束，则对JobRecord进行更新，同时添加回调第三方
*/
func isUpdateJobRecord(recordID string) error {
	logger := getLogger()

	var record model.JobRecord
	if err := record.GetByID(recordID); err != nil {
		logger.Error("fail to find JobRecord by id", "recordId", recordID, "error", err)
		return err
	}

	// 获取所有的task列表
	tasks, err := findTasksByRecordID(record.ID)
	if err != nil {
		logger.Error("find host results by jobRecordId fail", "jobRecordId", record.ID, "error", err)
		return err
	}
	// 校验是否更新JobRecord列表
	resultStatus, checkErr := isAllDone(tasks)

	if checkErr != nil {
		logger.Error("check host results all done fail", "error", checkErr)
		return checkErr
	}

	// 若为空，则说明不需要更新
	if resultStatus == "" {
		return nil
	}

	updateErr := updateJobRecord(&record, define.Done, resultStatus)
	if updateErr != nil {
		logger.Error("update job record fail", "error", updateErr)
		return updateErr
	}

	//触发web hook
	triggerJobDoneEvent(recordID, resultStatus)

	//记录报表
	reportStatistics(record.MasterID, resultStatus)

	return jobRecordCallback(record)

}

func triggerJobDoneEvent(recordID string, status string) {
	logger := getLogger()

	webhook.TriggerEvent(webhook.EventInfo{
		Event: define.WebHookEventJobDone,
		Payload: common.JobDonePayload{
			RecordID: recordID,
			Status:   status,
		},
	})

	eventInfo := webhook.EventInfo{
		Payload: common.JobDoneStatusPayload{recordID},
	}
	switch status {
	case define.Success:
		eventInfo.Event = define.WebHookEventJobSuccess
	case define.Fail:
		eventInfo.Event = define.WebHookEventJobFail
	case define.Timeout:
		eventInfo.Event = define.WebHookEventJobTimeout
	default:
		logger.Error("unknown job record result status")
	}

	webhook.TriggerEvent(eventInfo)
}

func getMasterPreKey(masterID string) string {
	return "master|" + masterID + "|job|"
}

func reportStatistics(masterID string, status string) {
	logger := getLogger()

	key := getMasterPreKey(masterID)
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
		return
	}

	if err != nil {
		logger.Error("report statistics fail", "error", err)
	}
}

/**
处理结果状态
结果状态分为三种情况：timeout、fail、succeed

若resultStatusMap的大小不为1，说明是三种状态混合，那肯定是fail
若resultStatusMap的大小为1，说明是三种状态的某一种，直接将key返回即可
*/
func checkResultStatus(resultStatusMap map[string]common.NullStruct) string {
	size := len(resultStatusMap)

	if size != 1 {
		return define.Fail
	}

	for k := range resultStatusMap {
		return k
	}
	return define.Fail
}

/**
jobRecord 执行完成之后回调
*/
func jobRecordCallback(record model.JobRecord) error {
	logger := getLogger()

	// 回调to third system； 若callback为空，则不回调
	if record.Callback == "" {
		logger.Info("callback is null, do not callback")
		return nil
	}
	logger.Debug(fmt.Sprintf("start to callback to third system, callbackurl: %s", record.Callback))

	results, err := FindHostResultsByRecordID(record.ID)
	if err != nil {
		return err
	}

	_, err = httputil.HttpPost(record.Callback, buildJobRecordCallbackBody(record, results))
	if err != nil {
		return err
	}
	return nil
}

/**
组装回调数据
*/
func buildJobRecordCallbackBody(record model.JobRecord, results []model.HostResult) common.JobCallbackParam {
	logger := getLogger()

	var hostResultCallbacks []common.HostResultCallback

	var hostIds []string
	for _, result := range results {
		hostIds = append(hostIds, result.HostID)
	}

	var act2Hosts []model.Act2Host
	model.GetDb().Where("id in (?)", hostIds).Find(&act2Hosts)

	hostMap := hostsToMap(act2Hosts)
	idcMap, _ := model.GetAllIdcMap()

	for _, result := range results {
		host, ok := hostMap[result.HostID]
		if !ok {
			logger.Error("fail to find host", "hostId", result.HostID)
			continue
		}
		idc, _ := idcMap[host.IdcID]

		back := common.HostResultCallback{
			EntityID: host.EntityID,
			HostIP:   result.HostIP,
			IdcName:  idc.Name,
			Message:  result.Message,
			Stderr:   result.Stderr,
			Stdout:   result.Stdout,
			Status:   result.ResultStatus,
		}

		hostResultCallbacks = append(hostResultCallbacks, back)
	}

	jobCallbackParam := common.JobCallbackParam{
		JobRecordID:   record.ID,
		ExecuteID:     record.ExecuteID,
		ResultStatus:  record.ResultStatus,
		ExecuteStatus: record.ExecuteStatus,
		HostResults:   hostResultCallbacks,
	}

	return jobCallbackParam
}

func hostsToMap(act2Hosts []model.Act2Host) (hostMap map[string]model.Act2Host) {
	hostMap = make(map[string]model.Act2Host)
	for _, host := range act2Hosts {
		hostMap[host.ID] = host
	}
	return
}

//MasterStat master状态数据
type MasterStat struct {
	EntityID string     `json:"entityId"`
	Done     ResultStat `json:"done"`
}

type MasterProxyStat struct {
	ID   string     `json:"id"`
	IDC  string     `json:"idc"`
	Done ResultStat `json:"done"`
}

//ResultStat 结果状态数据
type ResultStat struct {
	Success int64 `json:"success"`
	Fail    int64 `json:"fail"`
	Timeout int64 `json:"timeout"`
}

//MasterStat master作业记录
func GetMasterStat() (result map[string]interface{}, err error) {
	result = make(map[string]interface{})

	//master
	masterStats, err := getMasterStats()
	if err != nil {
		return nil, err
	}
	result["master"] = masterStats

	//proxy
	proxyStats, err := getMasterProxyStats()
	if err != nil {
		return nil, err
	}
	result["proxy"] = proxyStats
	return result, nil
}

func getMasterStats() (masterStats []MasterStat, err error) {
	logger := getLogger()

	masterIDs, err := model.FindAllMasterIDWithJobRecord()
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		logger.Error("find all master id by job record fail", "error", err)
		return nil, err
	}
	if gorm.IsRecordNotFoundError(err) {
		err = nil
	}

	masterStats = make([]MasterStat, 0, 10)
	for _, masterID := range masterIDs {
		if len(masterID) == 0 {
			continue
		}

		stat, err := getMasterStatByMasterID(masterID)
		if err != nil {
			logger.Error("get master stat fail", "masterID", masterID, "error", err)
			return nil, err
		}
		masterStats = append(masterStats, stat)
	}
	return masterStats, nil
}

func buildProxyStats(proxyIDCMap map[string]model.Act2Idc, proxyIDs []string) ([]MasterProxyStat, error) {
	proxyStats := make([]MasterProxyStat, 0, 10)
	for _, proxyID := range proxyIDs {
		if len(proxyID) == 0 {
			continue
		}

		idcName := ""
		idc, ok := proxyIDCMap[proxyID]
		if ok {
			idcName = idc.Name

		} else {
			idcName = "unknown idc"
		}

		stat, err := getMasterProxyStatByProxyID(proxyID, idcName)
		if err != nil {
			return nil, err
		}

		proxyStats = append(proxyStats, stat)
	}
	return proxyStats, nil
}

func getMasterProxyStats() (proxyStats []MasterProxyStat, err error) {
	logger := getLogger()

	proxyIDs, err := model.FindAllProxyIDWithJobTaskProxy()
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		logger.Error("find all proxy id by job task fail", "error", err)
		return nil, err
	}
	if gorm.IsRecordNotFoundError(err) {
		err = nil
	}

	proxyIDCMap, err := model.GetProxyIDIDCMapByProxyIDs(proxyIDs)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		logger.Error("find proxy and idc fail", "error", err)
		return nil, err
	}
	if gorm.IsRecordNotFoundError(err) {
		err = nil
	}

	return buildProxyStats(proxyIDCMap, proxyIDs)
}

func getMasterProxyStatByProxyID(proxyID, idcName string) (stat MasterProxyStat, err error) {
	preKey := getMasterProxyRedisPreKey(proxyID)

	success, err := getMasterCountByStatus(preKey, define.Success)
	if err != nil {
		return
	}

	fail, err := getMasterCountByStatus(preKey, define.Fail)
	if err != nil {
		return
	}

	timeout, err := getMasterCountByStatus(preKey, define.Timeout)
	if err != nil {
		return
	}

	stat = MasterProxyStat{
		ID:  proxyID,
		IDC: idcName,
		Done: ResultStat{
			Success: success,
			Fail:    fail,
			Timeout: timeout,
		},
	}

	return stat, nil
}

func getMasterStatByMasterID(masterID string) (stat MasterStat, err error) {

	preKey := getMasterPreKey(masterID)

	success, err := getMasterCountByStatus(preKey, define.Success)
	if err != nil {
		return
	}

	fail, err := getMasterCountByStatus(preKey, define.Fail)
	if err != nil {
		return
	}

	timeout, err := getMasterCountByStatus(preKey, define.Timeout)
	if err != nil {
		return
	}

	stat = MasterStat{
		EntityID: masterID,
		Done: ResultStat{
			Success: success,
			Fail:    fail,
			Timeout: timeout,
		},
	}

	return
}

func getMasterCountByStatus(preKey string, status string) (count int64, err error) {
	//logger := getLogger()
	recorder := report.GetRecorder(preKey)
	return recorder.GetRecord(status)
}

func getEncodingByVersionAndOsType(version string, osType string) string {
	if strings.HasPrefix(version, "2017.7") && osType == define.Win {
		return "gb18030"
	}
	return ""
}
