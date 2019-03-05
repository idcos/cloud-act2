//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package heartbeat

import (
	"fmt"

	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/host"
	"idcos.io/cloud-act2/service/salt"
)

type SaltHeartbeat struct {
	MasterInfo  *salt.MasterInfo   `json:"master"`
	MinionInfos []*salt.MinionInfo `json:"minions"`
}

func GetSaltHeartbeat() (*SaltHeartbeat, error) {
	logger := getLogger()
	masterInfo, err := salt.GetSaltMasterInfo("salt")
	if err != nil {
		logger.Error(fmt.Sprintf("get salt master info error %v", err))
		return nil, err
	}

	// 需要告警，但不能不提交上报
	minions, err := salt.GetMinionInfo()
	if err != nil {
		logger.Warn("get salt minion info failed", "error", err)
		minions = []*salt.MinionInfo{}
	}

	return &SaltHeartbeat{
		MasterInfo:  masterInfo,
		MinionInfos: minions,
	}, nil
}

// master 心跳接口，用于提供master每几分钟就上报一次
func MasterHeat(master common.Master) error {
	idc, proxy, err := host.ProcessIdcAndProxy(master)
	if err != nil {
		return err
	}
	return saveData(idc, proxy)
}

func saveData(idc model.Act2Idc, proxy model.Act2Proxy) error {
	logger := getLogger()
	db := model.GetDb()

	tx := db.Begin()

	if err := tx.Save(&idc).Error; err != nil {
		tx.Rollback()
		logger.Error("fail to save Act2Proxy", "error", err)
		return err
	}
	if err := tx.Save(&proxy).Error; err != nil {
		tx.Rollback()
		logger.Error("fail to save Act2Proxy", "error", err)
		return err
	}

	return tx.Commit().Error
}
