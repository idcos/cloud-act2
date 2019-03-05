//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package schedule

import (
	"fmt"
	"idcos.io/cloud-act2/utils/promise"
	"sync"
	"time"

	"idcos.io/cloud-act2/define"

	"github.com/robfig/cron"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/heartbeat"
	"idcos.io/cloud-act2/service/job"
	"idcos.io/cloud-act2/webhook"
)

/**
Act2 master 轮询配置，监听job超时和salt master心跳
*/
func MasterSchedule() {
	logger := getLogger()

	c := cron.New()

	heartbeatTimeoutInterval := fmt.Sprintf("every %s", config.Conf.Heartbeat.TimeoutInterval)
	c.AddFunc(heartbeatTimeoutInterval, CronWrapper("masterHeartbeat", masterHeatMonitor))

	jobTimeoutInterval := fmt.Sprintf("every %s", config.Conf.JobExpire.TimeoutInterval)
	logger.Info("timeout interval", "cron", jobTimeoutInterval)
	c.AddFunc(jobTimeoutInterval, CronWrapper("scanTimeoutJob", ScanTimeoutJob))

	c.Start()
}

func CronWrapper(taskName string, handler func()) func() {
	logger := getLogger()
	return func() {
		startTime := time.Now()
		logger.Info(fmt.Sprintf("task %s start %v", taskName, startTime))

		handler()

		logger.Info(fmt.Sprintf("task %s run cron %v", taskName, time.Since(startTime)))
	}
}

func ScanTimeoutJob() {
	logger := getLogger()

	timeout := config.Conf.JobExpire.Expire
	records, err := job.FindTimeoutJobs(timeout)
	if err != nil {
		logger.Error("scan timeout job error", err)
	}

	if len(records) > 0 {
		err = job.ExpireJobs(records)
		if err != nil {
			logger.Error("update timeout job error", err)
		} else {
			logger.Info("%n records (%v) has been set timeout", len(records), records)
		}
	} else {
		logger.Info("no timeout job record")
	}

}

//ProxySchedule proxy schedule
func ProxySchedule() {
	logger := getLogger()

	c := cron.New()
	heartBeatCrontabExpress := fmt.Sprintf("@every %s", config.Conf.Heartbeat.RegisterInterval)
	logger.Info("heartbeat register interval", "cron", heartBeatCrontabExpress)

	c.AddFunc(heartBeatCrontabExpress, CronWrapper("heartbeat", heartBeat))

	reportCrontabExpress := fmt.Sprintf("@every %s", config.Conf.Heartbeat.ReportInterval)
	logger.Debug("report interval ", "cron", reportCrontabExpress)

	c.AddFunc(reportCrontabExpress, CronWrapper("report salt info", func() {
		heartbeat.RegisterSaltInfo(true)
	}))

	c.Start()
}

// master心跳监测程序
func masterHeatMonitor() {
	logger := getLogger()

	startTime := time.Now()

	logger.Info("start proxy heat monitor,", "startTime", startTime)

	var proxies []model.Act2Proxy
	db := model.GetDb()
	db.Find(&proxies)

	duration, _ := time.ParseDuration(config.Conf.Heartbeat.RegisterInterval)
	registerTimeout := int(duration.Minutes())

	var wait sync.WaitGroup

	for _, proxy := range proxies {
		if int(startTime.Sub(proxy.LastTime).Minutes())-registerTimeout > 1 {
			logger.Error("proxy register time more than register interval ")

			// 更新状态
			wait.Add(1)
			proxy.Status = define.Fail
			promise.NewGoPromise(func(chan struct{}) {
				defer wait.Done()
				proxy.Save()

				//触发web hook proxy lose
				webhook.TriggerEvent(webhook.EventInfo{
					Event: define.WebHookEventProxyLose,
					Payload: common.ProxyPayload{
						ProxyID:     proxy.ID,
						ProxyServer: proxy.Server,
						ProxyType:   proxy.Type,
					},
				})
			}, nil)
			break
		}
	}
	wait.Wait()
}

func heartBeat() {
	logger := getLogger()

	startTime := time.Now()
	channelType := config.Conf.ChannelType
	logger.Info(fmt.Sprintf("register %s info ", channelType), "start", startTime.String())

	if channelType == "salt" {
		err := heartbeat.RegisterSaltInfo(false)
		if err != nil {
			logger.Error("register to master with salt info", "error", err, "elapse", time.Since(startTime))
			return
		}
	} else if channelType == "puppet" {
		err := heartbeat.RegisterPuppetInfo()
		if err != nil {
			logger.Error("register puppet info", "error", err)
			return
		}
	} else {
		logger.Error("unknown channel type", "type", channelType)
	}

	logger.Info(fmt.Sprintf("register to master with %s info complete ", channelType), "elapse", time.Since(startTime))
}
