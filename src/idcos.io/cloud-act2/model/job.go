//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package model

import (
	"time"
)

/**
作业执行记录表

longtext声明：https://github.com/jinzhu/gorm/issues/510
*/
// JobRecord job record
type JobRecord struct {
	ID            string    `gorm:"column:id;primary_key" json:"id"`
	StartTime     time.Time `gorm:"column:start_time" json:"startTime"`
	EndTime       time.Time `gorm:"column:end_time" json:"endTime"`
	ExecuteStatus string    `gorm:"column:execute_status" json:"executeStatus"`
	ResultStatus  string    `gorm:"column:result_status" json:"resultStatus"`
	Callback      string    `gorm:"column:callback" json:"callback"`
	Hosts         string    `gorm:"column:hosts" sql:"type:text" json:"hosts"`
	Provider      string    `gorm:"column:provider" json:"provider"`
	Script        string    `gorm:"column:script" sql:"type:text" json:"script"`
	ScriptType    string    `gorm:"column:script_type" json:"scriptType"`
	Pattern       string    `gorm:"column:pattern" json:"pattern"`
	Timeout       int       `gorm:"column:timeout" json:"timeout"`
	Parameters    string    `gorm:"column:parameters" json:"parameters"`
	MasterID      string    `gorm:"column:master_id" json:"masterId"`
	User          string    `gorm:"column:user" json:"user"` // 外部调用的用户名
	ExecuteID     string    `gorm:"column:execute_id" json:"executeId"`
}

// JobTask job task
type JobTask struct {
	ID            string    `gorm:"column:id;primary_key"`
	RecordID      string    `gorm:"column:record_id"`
	StartTime     time.Time `gorm:"column:start_time"`
	EndTime       time.Time `gorm:"column:end_time"`
	ExecuteStatus string    `gorm:"column:execute_status"`
	ResultStatus  string    `gorm:"column:result_status"`
	Pattern       string    `gorm:"column:pattern"`
	Script        string    `gorm:"column:script" sql:"type:text"`
	Params        string    `gorm:"column:params"`
	Options       string    `gorm:"column:options"`
}

//JobTaskProxy proxy的执行记录表
type JobTaskProxy struct {
	ID            string    `gorm:"column:id;primary_key"`
	TaskID        string    `gorm:"column:task_id"`
	ProxyID       string    `gorm:"column:proxy_id"`
	StartTime     time.Time `gorm:"column:start_time"`
	EndTime       time.Time `gorm:"column:end_time"`
	ExecuteStatus string    `gorm:"column:execute_status"`
	ResultStatus  string    `gorm:"column:result_status"`
}

/**
  单个主机执行结果表
*/
type HostResult struct {
	ID            string    `gorm:"column:id;primary_key"`
	TaskID        string    `gorm:"column:task_id"`
	HostID        string    `gorm:"column:host_id"`
	ProxyID       string    `gorm:"column:proxy_id"`
	StartTime     time.Time `gorm:"column:start_time"`
	EndTime       time.Time `gorm:"column:end_time"`
	ExecuteStatus string    `gorm:"column:execute_status"`
	ResultStatus  string    `gorm:"column:result_status"`
	HostIP        string    `gorm:"column:host_ip"`
	Stdout        string    `gorm:"column:stdout"`
	Stderr        string    `gorm:"column:stderr"`
	Message       string    `gorm:"column:message"`
}

/**
TableName convert
*/
func (JobRecord) TableName() string {
	return "act2_job_record"
}

func (r *JobRecord) Create() {
	globalDb.Create(r)
}

func (r *JobRecord) DeleteByID() {
	globalDb.Where("id = ? ", r.ID).Delete(r)
}

/**
create or update
*/
func (r *JobRecord) Save() error {
	return globalDb.Save(r).Error
}

func (r *JobRecord) Update(args ...interface{}) {
	globalDb.Model(r).Update(args...)
}

func (r *JobRecord) GetByID(ID string) error {
	return globalDb.Where("id = ?", ID).Find(r).Error
}

func FindJobRecordByIDs(ids []string) (jobRecords []JobRecord, err error) {
	jobRecords = make([]JobRecord, 0, len(ids))

	length := len(ids)
	for length > 0 {
		//当in的数据超过6000时间会大幅度增加
		endIndex := 6000
		if length < 6000 {
			endIndex = length
		}

		tempJobRecords := make([]JobRecord, 0, 10)
		err = globalDb.Model(&JobRecord{}).Where("id in (?)", ids[:endIndex]).Find(&tempJobRecords).Error
		if err != nil {
			return jobRecords, err
		}

		jobRecords = append(jobRecords, tempJobRecords...)

		if endIndex == length {
			break
		}
		ids = ids[endIndex:]
		length = len(ids)
	}

	return
}

//FindAllMasterIDWithJobRecord 获取所有的masterID
func FindAllMasterIDWithJobRecord() (masterIDs []string, err error) {
	masterIDs = make([]string, 0, 10)
	err = globalDb.Model(&JobRecord{}).Select("master_id").Group("master_id").Pluck("master_id", &masterIDs).Error
	return
}

func (JobTask) TableName() string {
	return "act2_job_task"
}

func (r *JobTask) GetByID(ID string) error {
	return globalDb.Where("id = ?", ID).Find(r).Error
}

func (r *JobTask) Save() error {
	return globalDb.Save(r).Error
}

//FindAllProxyIDWithJobTask 查找所有的proxyId
func FindAllProxyIDWithJobTaskProxy() (proxyIDs []string, err error) {
	proxyIDs = make([]string, 0, 10)
	err = globalDb.Model(&JobTaskProxy{}).Select("proxy_id").Group("proxy_id").Pluck("proxy_id", &proxyIDs).Error
	return
}

//UpdateTaskStatus 更新task状态
func UpdateTaskStatusToDone(taskID, resultStatus string) (rowsAffected int64, err error) {
	db := GetDb().Model(&JobTask{}).Where("id = ? and execute_status = ?", taskID, "doing").
		Update(map[string]interface{}{"end_time": time.Now(), "execute_status": "done", "result_status": resultStatus})
	return db.RowsAffected, db.Error
}

//UpdateTaskProxyID 更新task的proxyID
func UpdateTaskProxyID(proxyID string, taskID string) (err error) {
	return globalDb.Model(&JobTask{}).Where("id = ?", taskID).Update("proxy_id", proxyID).Error
}

//FindHostResultUndoneByTaskID 根据taskID获取未完成的主机列表
func FindHostResultUndoneByTaskID(taskID string) (hostResults []HostResult, err error) {
	hostResults = make([]HostResult, 0, 10)
	err = globalDb.Model(&HostResult{}).Where("execute_status <> ? and task_id = ?", "done", taskID).Find(&hostResults).Error
	return
}

func (JobTaskProxy) TableName() string {
	return "act2_job_task_proxy"
}

func (r *JobTaskProxy) Save() error {
	return globalDb.Save(r).Error
}

//FindTaskProxyByTaskID 根据taskID查找所有的taskProxy
func FindTaskProxyByTaskID(taskID string) (taskProxies []JobTaskProxy, err error) {
	taskProxies = make([]JobTaskProxy, 0, 10)
	err = globalDb.Model(&JobTaskProxy{}).Where("task_id = ?", taskID).Find(&taskProxies).Error
	return
}

//FindTaskProxyByTaskIDAndProxyID 根据taskID和ProxyID来查询taskProxy
func FindTaskProxyByTaskIDAndProxyID(taskID, proxyID string) (taskProxy JobTaskProxy, err error) {
	taskProxy = JobTaskProxy{}
	err = globalDb.Model(&JobTaskProxy{}).Where("task_id = ? and proxy_id = ?", taskID, proxyID).Find(&taskProxy).Error
	return
}

func UpdateTaskProxyProxyID(taskID string, proxyID string) error {
	err := globalDb.Model(&JobTaskProxy{}).Where("task_id = ? ", taskID).Update("proxy_id", proxyID).Error
	return err
}

type TaskRecoveryInfo struct {
	ID        string    `gorm:"column:id"`
	StartTime time.Time `gorm:"column:start_time"`
	Timeout   int       `gorm:"column:timeout"`
}

//FindJobTaskByExecuteStatusWithRecent 根据执行状态获取近一天的task记录
func FindJobTaskByExecuteStatusWithLastDay(executeStatus string) (taskInfoList []TaskRecoveryInfo, err error) {
	taskInfoList = make([]TaskRecoveryInfo, 0, 100)

	err = GetDb().Table("act2_job_task as task").
		Select("task.id,task.start_time,record.timeout").
		Joins("left join act2_job_record record on task.record_id = record.id").
		Where("task.execute_status = ? and task.start_time > date_sub(curdate(),interval 1 day)", executeStatus).Find(&taskInfoList).Error
	return
}

func (HostResult) TableName() string {
	return "act2_host_result"
}

func (HostResult) GetHostResultsColumns() []string {
	return []string{"id", "start_time", "end_time", "execute_status", "result_status", "host_ip", "idc", "stdout", "stderr"}
}

func (r *HostResult) Create() {
	globalDb.Create(r)
}

func (r *HostResult) DeleteByID() {
	globalDb.Where("ID = ? ", r.ID).Delete(r)
}

/**
create or update
*/
func (r *HostResult) Save() {
	globalDb.Save(r)
}

func (r *HostResult) Update(args ...interface{}) {
	globalDb.Model(r).Update(args...)
}

func (r *HostResult) GetByID(ID string) {
	globalDb.Where("id = ?", ID).Find(r)
}

//FindHostResultByTaskIDAndProxyID 根据taskID查询主机结果
func FindHostResultByTaskID(taskID string) (hostResults []HostResult, err error) {
	hostResults = make([]HostResult, 0, 10)
	err = globalDb.Model(&HostResult{}).Where("task_id = ?", taskID).Find(&hostResults).Error
	return
}

//FindHostResultByTaskIDAndProxyID 根据taskID和proxyID查询主机结果
func FindHostResultByTaskIDAndProxyID(taskID string, proxyID string) (hostResults []HostResult, err error) {
	hostResults = make([]HostResult, 0, 10)
	err = globalDb.Model(&HostResult{}).Where("task_id = ? and proxy_id = ?", taskID, proxyID).Find(&hostResults).Error
	return
}

// JobRecord和主机执行结果及主机信息s
type RecordResult struct {
	JobRecordID string `gorm:"column:recordId" json:"recordId"`
	HostIP      string `gorm:"column:hostIp" json:"hostIp"`
	HostID      string `gorm:"column:hostId" json:"hostId"`
	EntityID    string `gorm:"column:entityId" json:"entityId"`
	OsType      string `gorm:"column:osType" json:"osType"`
	StartTime   string `gorm:"column:startTime" json:"startTime"`
	EndTime     string `gorm:"column:endTime" json:"endTime"`
	Status      string `gorm:"column:status" json:"status"`
	IdcName     string `gorm:"column:idcName" json:"idcName"`
	StdOut      string `gorm:"column:stdout" json:"stdout"`
	StdErr      string `gorm:"column:stderr" json:"stderr"`
	Message     string `gorm:"column:message" json:"message"`
}
