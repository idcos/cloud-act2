//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"encoding/json"
	"idcos.io/cloud-act2/config"

	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/redis"
)

//GetPublishID 获取发布消息的id
func GetPublishID(taskID string, entityID string) string {
	return taskID + "|" + entityID
}

//RealTimeToRedis 将实时输出发送到redis
func RealTimeToRedis(realTimeForm common.RealTimeForm) (err error) {
	if !config.Conf.DependRedis {
		return
	}

	logger := getLogger()

	jobInfo, ok := GetInfoByJid(realTimeForm.Jid)
	if !ok {
		logger.Trace("jid not found", "jid", realTimeForm.Jid)
		return
	}

	realTimeRedisForm := common.RealTimeRedisForm{
		Now:    realTimeForm.Now,
		Stdout: realTimeForm.Stdout,
		Stderr: realTimeForm.Stderr,
	}

	bytes, err := json.Marshal(realTimeRedisForm)
	if err != nil {
		logger.Error("real time data to json fail", "data", realTimeRedisForm, "error", err)
		return
	}

	logger.Trace("send redis publish", "tankID", jobInfo.TaskID, "entityID", realTimeForm.EntityID, "body", string(bytes))

	client := redis.GetRedisClient()
	cmd := client.Publish(GetPublishID(jobInfo.TaskID, realTimeForm.EntityID), string(bytes))
	err = cmd.Err()
	if err != nil {
		logger.Error("send data to redis fail", "error", err)
		return
	}
	return
}
