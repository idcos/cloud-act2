//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/timingwheel"
	"idcos.io/cloud-act2/utils/dataexchange"
	"time"
)

//timeout: 单位：s
func FindTimeoutJobs(timeout int) ([]*model.JobRecord, error) {
	logger := getLogger()

	var records []*model.JobRecord

	db := model.GetDb()
	// time.Now().Add(time.Duration(timeout) * time.Second)
	db = db.Where("execute_status = ? and (start_time + interval ? second) < now()", define.Doing, timeout).Find(&records)
	if db.Error != nil {
		logger.Error(fmt.Sprintf("find timeout job record error %s", db.Error))
		return nil, errors.Wrap(db.Error, "find timeout job record error")
	}

	return records, nil
}

func ExpireJobs(records []*model.JobRecord) error {
	logger := getLogger()

	var ids []string
	for _, record := range records {
		ids = append(ids, record.ID)
	}

	db := model.GetDb()

	db = db.Model(&model.JobRecord{}).Where("id in (?)", ids).Update(map[string]interface{}{
		"execute_status": "done",
		"result_status":  "timeout",
	})
	if db.Error != nil {
		logger.Error(fmt.Sprintf("expire job timeout error %s", db.Error))
		return db.Error
	}

	return nil
}

func ListenTaskTimeout(timeout int, taskID string) {
	logger := getLogger()

	logger.Info("start listen job record timeout", "taskID", taskID)

	timingwheel.AddTask(timeout, func() {
		sendTimeoutResult(taskID)
	})
}

//RestartRecovery 超时控制的重启恢复
func RestartRecovery() (err error) {
	logger := getLogger()

	logger.Info("start recovery timeout")

	//获取所有执行中但未结束的任务
	taskInfoList, err := model.FindJobTaskByExecuteStatusWithLastDay(define.Doing)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}

	logger.Trace("Need to resume the task", "num", len(taskInfoList), "tasks", dataexchange.ToJsonString(taskInfoList))
	for _, taskInfo := range taskInfoList {
		timeoutDate := taskInfo.StartTime.Add(time.Duration(taskInfo.Timeout) * time.Second)
		processRecoveryTask(timeoutDate, taskInfo.ID)
	}
	return
}

func processRecoveryTask(timeoutDate time.Time, taskID string) {
	timeout := timeoutDate.Sub(time.Now()).Seconds()
	if timeout <= 1 {
		sendTimeoutResult(taskID)
		return
	}
	ListenTaskTimeout(int(timeout), taskID)
}

func sendTimeoutResult(taskID string) {
	logger := getLogger()

	callbackResult := common.ProxyCallBackResult{
		Status:  define.Timeout,
		Message: "act2-master to act2-proxy timeout",
		Content: common.ProxyResultContent{
			TaskID:     taskID,
			MasterSend: true,
		},
	}
	err := HostResultsCallback(callbackResult)
	if err != nil {
		logger.Error("process timeout host results fail", "taskID", taskID, "error", err)
	}
}
