//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package heartbeat

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/log"
	"idcos.io/cloud-act2/server/common"
	"idcos.io/cloud-act2/service/salt"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/httputil"
	"idcos.io/cloud-act2/utils/iputil"
	"strings"
	"time"
)

func ReportDelay() {
	getLogger().Debug("will delay", "second", hclog.Fmt("%ds", config.Conf.Act2.DelayReport))
	time.Sleep(time.Duration(config.Conf.Act2.DelayReport) * time.Second)
}

func registerToMaster(heartbeat *SaltHeartbeat) error {
	logger := log.L().Named("heartbeat.proxy")

	logger.Debug("register", "info", dataexchange.ToJsonString(heartbeat))

	act2Server := config.Conf.Act2.ClusterServer
	registerURL := strings.TrimRight(act2Server, "/") + define.RegisterUri

	respData, err := httputil.HttpPost(registerURL, heartbeat)
	if err != nil {
		logger.Error("register to act2 server", "error", err)
		return err
	}

	logger.Debug("register act2 server ", "response", string(respData))

	jsonResult := common.JSONResult{}
	err = json.Unmarshal(respData, &jsonResult)
	if err != nil {
		logger.Error("register to master response not valid json", "error", err)
		return err
	}

	if jsonResult.Status == define.Error {
		return errors.New(jsonResult.Message)
	}
	return nil
}

//RegisterSaltInfo register salt info
func RegisterSaltInfo(all bool) error {
	logger := log.L().Named("heartbeat.proxy")

	ReportDelay()

	var heartbeat *SaltHeartbeat
	// 上报的时候，上报所有的数据
	if all {
		var err error
		heartbeat, err = GetSaltHeartbeat()
		if err != nil {
			return err
		}
	} else {
		masterInfo, err := salt.GetSaltMasterInfo(config.Conf.ChannelType)
		if err != nil {
			logger.Error(fmt.Sprintf("get %s master info", config.Conf.ChannelType), "error", err)
			return err
		}

		heartbeat = &SaltHeartbeat{
			MasterInfo: masterInfo,
		}
	}

	return registerToMaster(heartbeat)
}

func extractNetworkingIPs(networking map[string]interface{}) []string {
	var results []string

	interfaces := networking["interfaces"].(map[string]interface{})
	for _, value := range interfaces {
		ethValue := value.(map[string]interface{})
		ip := ethValue["ip"].(string)
		if !iputil.IsReversedIP(ip) {
			results = append(results, ip)
		}
	}
	return results
}

func extractProductID(uuid interface{}) string {
	// uuid的值是什么？
	return uuid.(string)
}

func queryPuppetAgentInfos() ([]*salt.MinionInfo, error) {
	logger := getLogger()

	dbServer := config.Conf.Puppet.PuppetDBServer
	dbServerURL := fmt.Sprintf("%s/pdb/query/v4/facts", strings.TrimRight(dbServer, "/"))

	respData, err := httputil.HttpPost(dbServerURL, nil)
	if err != nil {
		logger.Error("register to act2 server", "error", err)
		return nil, err
	}

	var facts []map[string]interface{}
	err = json.Unmarshal(respData, &facts)
	if err != nil {
		logger.Error("unmarshal facts", "error", err)
		return nil, err
	}

	factMap := map[string]*salt.MinionInfo{}
	for _, fact := range facts {
		certname := fact["certname"].(string)
		key := fact["name"].(string)
		value := fact["value"]

		minionInfo, ok := factMap[certname]
		if !ok {
			minionInfo = &salt.MinionInfo{}
		}

		switch key {
		case "networking":
			networking := value.(map[string]interface{})
			ips := extractNetworkingIPs(networking)
			minionInfo.IPs = ips
		case "uuid":
			productID := extractProductID(value)
			minionInfo.SerialNumber = productID
		}
		factMap[certname] = minionInfo
	}

	logger.Debug("fact map", "length", hclog.Fmt("%d", len(factMap)))

	var minionInfos []*salt.MinionInfo
	for _, minionInfo := range factMap {
		minionInfos = append(minionInfos, minionInfo)
	}

	return minionInfos, nil
}

func RegisterPuppetInfo() error {
	logger := getLogger()

	minionInfos, err := queryPuppetAgentInfos()
	if err != nil {
		logger.Error("puppet agent info", "error", err)
		return err
	}

	masterInfo, err := salt.GetSaltMasterInfo("puppet")
	if err != nil {
		logger.Error("puppet master info", "error", err)
		return err
	}

	heartbeat := &SaltHeartbeat{
		MasterInfo:  masterInfo,
		MinionInfos: minionInfos,
	}

	return registerToMaster(heartbeat)
}
