//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/generator"
	"time"

	"github.com/jinzhu/gorm"
	"idcos.io/cloud-act2/define"

	"github.com/theplant/batchputs"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
)

/**
更新JobRecord
*/
func updateJobRecord(jobRecord *model.JobRecord, execStatus string, resultStatus string) error {
	logger := getLogger()

	jobRecord.ExecuteStatus = execStatus
	jobRecord.ResultStatus = resultStatus
	jobRecord.EndTime = time.Now()
	jsErr := jobRecord.Save()
	if jsErr != nil {
		logger.Error("fail to save JobRecord", "error", jsErr, "JobRecord", dataexchange.ToJsonString(jobRecord))
		return jsErr
	}
	return nil

}

func updateHostResult(hostResultInfos []model.HostResult) error {
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
			result.ExecuteStatus,
			result.ResultStatus,
			result.Stdout,
			result.Stderr,
			result.Message,
		})
	}

	columns := []string{"id", "host_id", "task_id", "host_ip", "start_time", "end_time", "execute_status", "result_status", "stdout", "stderr", "message"}

	if err := batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.HostResult{}.TableName(), "id", columns, hostResults); err != nil {
		logger.Error("fail update host result", "error", err)
		return err
	}
	return nil
}

func updateTaskProxy(taskProxies []model.JobTaskProxy) error {
	logger := getLogger()

	results := make([][]interface{}, len(taskProxies))
	for i, taskProxy := range taskProxies {
		results[i] = []interface{}{
			taskProxy.ID,
			taskProxy.TaskID,
			taskProxy.ProxyID,
			taskProxy.StartTime,
			taskProxy.EndTime,
			taskProxy.ExecuteStatus,
			taskProxy.ResultStatus,
		}
	}

	columns := []string{"id", "task_id", "proxy_id", "start_time", "end_time", "execute_status", "result_status"}

	if err := batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.JobTaskProxy{}.TableName(), "id", columns, results); err != nil {
		logger.Error("fail update job task proxy", "error", err)
		return err
	}
	return nil
}

func SaveTaskProxy(taskID string, proxyID string) error {
	taskProxy := &model.JobTaskProxy{
		ID:            generator.GenUUID(),
		TaskID:        taskID,
		ProxyID:       proxyID,
		StartTime:     time.Now(),
		EndTime:       time.Now(),
		ExecuteStatus: define.Doing,
	}
	return taskProxy.Save()
}

/**
save host doing to result
hostInfo列表
*/
func SaveHostResult(jobTask model.JobTask, hostInfos []common.ExecHost, execStatus string, proxyID string) error {
	logger := getLogger()

	var hostResults [][]interface{}
	for _, hostInfo := range hostInfos {
		hostResults = append(hostResults, []interface{}{
			generator.GenUUID(),
			jobTask.ID,
			hostInfo.HostID,
			proxyID,
			time.Now(),
			time.Now(),
			execStatus,
			"",
			hostInfo.HostIP,
			"",
			"",
			"",
		})
	}

	columns := []string{"id", "task_id", "host_id", "proxy_id", "start_time", "end_time", "execute_status", "result_status", "host_ip", "stdout", "stderr", "message"}

	if err := batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.HostResult{}.TableName(), "id", columns, hostResults); err != nil {
		logger.Error("fail save host result", "error", err)
		return err
	}
	return nil
}

func findTasksByRecordID(recordID string) (tasks []model.JobTask, err error) {
	logger := getLogger()

	if err = model.GetDb().Table("act2_job_task").Where("record_id = ?", recordID).Find(&tasks).Error; err != nil {
		logger.Error("fail to find job_task", "recordID", recordID, "error", err)
		return
	}

	return
}

func findUnDoneTasksByRecordID(recordID string) (tasks []model.JobTask, err error) {
	logger := getLogger()

	if err = model.GetDb().Table("act2_job_task").Where("record_id = ? and 'done' != execute_status", recordID).Find(&tasks).Error; err != nil {
		logger.Error("fail to find job_task", "recordID", recordID, "error", err)
		return
	}

	return
}

/**
根据jobRecordId查询hostResults列表
*/
func findHostResultsByTaskId(taskID string) ([]model.HostResult, error) {
	logger := getLogger()

	var hostResults []model.HostResult
	if err := model.GetDb().Table("act2_host_result").Where("task_id = ?", taskID).Find(&hostResults).Error; err != nil {
		logger.Error("fail to find host_result by job_record_id", "taskID", taskID, "error", err)
		return nil, err
	}
	return hostResults, nil
}

func FindHostResultsByRecordID(recordID string) (hostResults []model.HostResult, err error) {
	logger := getLogger()

	err = model.GetDb().Table("act2_job_record as record").
		Select("result.*").
		Joins("left join act2_job_task task on task.record_id  = record.id").
		Joins("left join act2_host_result result on result.task_id = task.id").
		Where("record_id = ?", recordID).Find(&hostResults).Error

	if err != nil {
		logger.Error("fail to find hostResults by job record", "recordID", recordID, "error", err)
		return
	}
	return
}

func FindHostResultsByHostIds(taskID string, hostIds []string) (hostResults []model.HostResult, err error) {
	err = model.GetDb().Where(" task_id = ? and host_id in (?)", taskID, hostIds).Find(&hostResults).Error
	return
}

func FindHostResultByRecordIDAndEntityID(recordID, entityID string) (hostResults model.HostResult, err error) {
	logger := getLogger()

	err = model.GetDb().Table("act2_job_record as record").
		Select("result.*").
		Joins("left join act2_job_task task on task.record_id  = record.id").
		Joins("left join act2_host_result result on result.task_id = task.id").
		Joins("left join act2_host host on result.host_id  = host.id").
		Where("record_id = ? and host.entity_id = ?", recordID, entityID).Find(&hostResults).Error

	if err != nil {
		logger.Error("fail to find hostResults by job record and entityId", "recordID", recordID, "entityID", entityID, "error", err)
		return
	}
	return
}

func findUnDoneHostResultsByRecordID(recordID string) (hostResults []model.HostResult, err error) {
	logger := getLogger()

	err = model.GetDb().Table("act2_job_record as record").
		Select("result.*").
		Joins("left join act2_job_task task on task.record_id  = record.id").
		Joins("left join act2_host_result result on result.task_id = task.id").
		Where("record_id = ? and 'done' != result.execute_status", recordID).Find(&hostResults).Error

	if err != nil {
		logger.Error("fail to find job_task", "recordID", recordID, "error", err)
		return
	}
	return
}

/** 根据主机uuid列表查询机器信息

SELECT ip.ip,host.id,host.system_id,idc.name,host.status  FROM act2_host_ip ip LEFT JOIN act2_host host ON host.id = ip.host_id INNER JOIN
  act2_idc idc ON idc.id = host.idc_id WHERE host.entity_id IN ('10.0.1.0');
*/
func findHostInfoByIDs(hostIds []string, provider string) (hostInfos []common.HostInfo, err error) {
	logger := getLogger()

	db := model.GetDb()

	err = db.Table("act2_host_ip as ip").
		Select("ip.ip as hostIp,host.id as hostId,host.entity_id as entityId, host.os_type as osType,host.minion_version as minionVersion,idc.name as idcName,idc.id as idcId ,host.status,proxy.type,proxy.server,proxy.options,proxy.id as proxyId").
		Joins("left join act2_host as host on  ip.host_id = host.id").
		Joins("inner join act2_idc as idc on  idc.id = host.idc_id").
		Joins("inner join act2_proxy as proxy on proxy.id = host.proxy_id").
		Where("host.entity_id in (?) and proxy.type = ?", hostIds, provider).
		Find(&hostInfos).Error

	if err != nil {
		logger.Error("find host info error; ", "error", err.Error())
		return
	}
	return
}

func FindHostInfoByEntityID(entityID string) (hostInfo common.HostInfo, err error) {
	hostInfo = common.HostInfo{}
	err = model.GetDb().Table("act2_host_ip as ip").
		Select("ip.ip as hostIp,host.id as hostId,host.entity_id as entityId, host.os_type as osType,host.minion_version as minionVersion,idc.name as idcName,idc.id as idcId ,host.status,proxy.type,proxy.server,proxy.options,proxy.id as proxyId").
		Joins("left join act2_host as host on  ip.host_id = host.id").
		Joins("inner join act2_idc as idc on  idc.id = host.idc_id").
		Joins("inner join act2_proxy as proxy on proxy.id = host.proxy_id").
		Where("host.entity_id = ?", entityID).
		First(&hostInfo).Error
	return
}

func findHostInfoByHostIDs(hostIDs []string) (hostInfoList []common.HostInfo, err error) {
	logger := getLogger()

	db := model.GetDb()
	hostInfoList = make([]common.HostInfo, 0, 10)
	err = db.Table("act2_host_ip as ip").
		Select("ip.ip as hostIp,host.id as hostId,host.entity_id as entityId, host.os_type as osType,idc.name as idcName,idc.id as idcId").
		Joins("left join act2_host as host on  ip.host_id = host.id").
		Joins("inner join act2_idc as idc on  idc.id = host.idc_id").
		Where("host.id in (?)", hostIDs).Find(&hostInfoList).Error

	if err != nil && !gorm.IsRecordNotFoundError(err) {
		logger.Error("find host info fail", "error", err)
		return
	}
	if gorm.IsRecordNotFoundError(err) {
		err = nil
	}
	return
}

func FindHostInfoByIDCName(idcName string) (hostInfos []common.HostInfo, err error) {
	logger := getLogger()

	db := model.GetDb()

	err = db.Table("act2_host_ip as ip").
		Select("ip.ip as hostIp,host.id as hostId,host.entity_id as entityId, host.os_type as osType,idc.name as idcName,idc.id as idcId ,host.status").
		Joins("left join act2_host as host on  ip.host_id = host.id").
		Joins("inner join act2_idc as idc on  idc.id = host.idc_id").
		Where("idc.name = ?", idcName).
		Order("ip.ip desc").
		Find(&hostInfos).Error

	if err != nil {
		logger.Error("findHostInfoByIdcName fail ", "error", err.Error())
		return
	}
	return
}
func FindAllIDCHostInfo() (hostInfos []common.HostInfo, err error) {
	logger := getLogger()

	db := model.GetDb()

	err = db.Table("act2_host_ip as ip").
		Select("ip.ip as hostIp,host.id as hostId,host.entity_id as entityId, host.os_type as osType,idc.name as idcName,idc.id as idcId ,host.status").
		Joins("left join act2_host as host on  ip.host_id = host.id").
		Joins("inner join act2_idc as idc on  idc.id = host.idc_id").
		Order("ip.ip desc").
		Find(&hostInfos).Error

	if err != nil {
		logger.Error("findHostInfoByIdcName fail ", "error", err.Error())
		return
	}
	return
}

func findHostInfoByIpAndIdc(ips []string, idc []string) (hostInfos []common.HostInfo, err error) {
	logger := getLogger()

	db := model.GetDb()

	err = db.Table("act2_host_ip as ip").
		Select("ip.ip as hostIp,host.id as hostId,host.entity_id as entityId, host.os_type as osType,idc.name as idcName,idc.id as idcId ,host.status").
		Joins("left join act2_host as host on  ip.host_id = host.id").
		Joins("inner join act2_idc as idc on  idc.id = host.idc_id").
		Where("ip.ip in (?)  and idc.name in (?)", ips, idc).
		Find(&hostInfos).Error

	if err != nil {
		logger.Error("findHostInfoByIpAndIdc fail ", "error", err.Error())
		return
	}
	return
}

//ProxyIDIdcID
type ProxyIDIdcID struct {
	ProxyID string `gorm:"proxy_id"`
	IdcID   string `gorm:"idc_id"`
}

func findProxyIDAndIdcIDByIdcNames(idcNames []string) (proxyIdcs []ProxyIDIdcID, err error) {
	proxyIdcs = make([]ProxyIDIdcID, 0, 10)

	err = model.GetDb().Table("act2_proxy").
		Select("max(id) as proxy_id,idc_id").
		Group("idc_id").
		Having("idc_id in (?)", model.GetDb().Table("act2_idc as idc").Select("idc.id").Where("name in (?)", idcNames).QueryExpr()).
		Find(&proxyIdcs).Error
	return
}

func FindProxyIDByIdcName(idcName string) (proxyID string, err error) {
	proxyIDs := make([]string, 0, 1)
	err = model.GetDb().Table("act2_proxy").
		Select("id").
		Where("idc_id in (?)", model.GetDb().Table("act2_idc as idc").Select("idc.id").Where("idc.name = ?", idcName).QueryExpr()).
		Pluck("id", &proxyIDs).Error
	if len(proxyIDs) > 0 {
		return proxyIDs[0], err
	}
	return "", err
}

func findHostInfoByIps(ips []string, typ string) (hostInfos []common.HostInfo, err error) {
	logger := getLogger()

	db := model.GetDb()

	err = db.Table("act2_host_ip as ip").
		Select("ip.ip as hostIp,host.id as hostId,host.entity_id as entityId, host.os_type as osType,host.minion_version as minionVersion,idc.name as idcName,idc.id as idcId ,host.status,host.proxy_id as proxyId").
		Joins("left join act2_host as host on  ip.host_id = host.id").
		Joins("inner join act2_idc as idc on  idc.id = host.idc_id").
		Where("ip.ip in (?) and proxy_id in (select id from act2_proxy where type=?)", ips, typ).
		Find(&hostInfos).Error

	if err != nil {
		logger.Error("find host info error; ", err.Error())
		return
	}
	return
}

func FindProxyByID(id string) (proxies model.Act2Proxy, err error) {
	logger := getLogger()

	err = model.GetDb().Table("act2_proxy as proxy").Where("id = ?", id).
		Find(&proxies).Error

	if err != nil {
		logger.Error("findProxyByProvider fail", "error", err.Error())
		return
	}

	return
}

func FindProxyByIdcName(idcNames []string) (proxies []common.ProxyInfo, err error) {
	logger := getLogger()

	err = model.GetDb().Table("act2_proxy as proxy").
		Select("proxy.id,proxy.server,proxy.type,idc.name as idcName,idc.id as idcId,proxy.status").
		Joins("left join act2_idc as idc on idc.id = proxy.idc_id").
		Where("idc.name in (?)", idcNames).
		Find(&proxies).Error

	if err != nil {
		logger.Error("findProxyByProviderAndIdcName fail", "error", err.Error())
		return
	}

	return
}

//SELECT record.id as jobRecordId, result.host_ip as hostIp,host.id as hostId,host.entity_id as entityId,result.result_status as status, idc.name as idcName,result.stdout,result.stderr,result.message FROM act2_job_record record
//LEFT JOIN act2_job_task task ON record.id = task.record_id
//LEFT JOIN act2_host_result result ON result.task_id = task.id
//INNER JOIN act2_host host ON result.host_id = host.id
//INNER JOIN act2_idc idc ON idc.id = host.idc_id
//WHERE record.id = '7a9a443a-c4a3-904a-765f-3b123a37642c';
func FindRecordResultsById(jobRecordId string) (results []model.RecordResult, err error) {

	err = model.GetDb().Table("act2_job_record as record").
		Select("record.id as recordId,record.start_time as startTime,record.end_time as endTime, result.host_ip as hostIp,host.id as hostId,host.entity_id as entityId, host.os_type as osType,result.result_status as status, idc.name as idcName,result.stdout,result.stderr,result.message").
		Joins("left join act2_job_task as task on record.id = task.record_id ").
		Joins("left join act2_host_result as result on task.id = result.task_id ").
		Joins("inner join act2_host as host on host.id = result.host_id").
		Joins("inner join act2_idc as idc on idc.id = host.idc_id").
		Where("record.id  =? ", jobRecordId).
		Find(&results).Error

	return
}

func FindRecordResultsCount() (count int64, err error) {
	err = model.GetDb().Table("act2_job_record").
		Select("id").
		Count(&count).Error

	return
}

func FindRecordResultsByPage(pageNo, pageSize int64) (results []model.RecordResult, err error) {
	err = model.GetDb().Raw(`SELECT record.id as recordId,record.start_time as startTime,record.end_time as endTime, result.host_ip as hostIp,host.id as hostId,host.entity_id as entityId,host.os_type as osType,result.result_status as status, idc.name as idcName,result.stdout,result.stderr,result.message
FROM act2_job_record as record
LEFT JOIN act2_job_task as task ON record.id = task.record_id
LEFT JOIN act2_host_result as result ON task.id= result.task_id
INNER JOIN act2_host as host ON host.id = result.host_id
INNER JOIN act2_idc as idc ON idc.id = host.idc_id
WHERE record.id IN (SELECT id from (SELECT id FROM act2_job_record ORDER BY start_time LIMIT ?,?) as temp)`, (pageNo-1)*pageSize, pageSize).Scan(&results).Error
	return
}
