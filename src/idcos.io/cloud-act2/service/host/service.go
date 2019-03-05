//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package host

import (
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/generator"
	"strings"
	"time"

	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/webhook"

	"github.com/jinzhu/gorm"

	"github.com/pkg/errors"
	"github.com/theplant/batchputs"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
)

// 根据ip列表获取主机uuid列表信息
func FindHostsByIPs(param []byte) ([]model.Act2HostIP, error) {
	var ips []string
	json.Unmarshal(param, &ips)

	logger := getLogger()

	var hostIPs []model.Act2HostIP
	if err := model.GetDb().Where("ip in (?)", ips).Find(&hostIPs).Error; err != nil {
		logger.Error("fail to find hosts by IPs; ", "error", err)
		return nil, err
	}
	return hostIPs, nil
}

// 获取idc列表
func FindIdcs() (idcs []model.Act2Idc, err error) {
	err = model.GetDb().Find(&idcs).Error
	return
}

// 获取proxy列表
func FindAllProxy() (proxyInfos []common.ProxyInfo, err error) {
	err = model.GetDb().Table("act2_idc as idc").
		Select("proxy.id,proxy.server,proxy.status,proxy.type,idc.name as idcName,idc.id as idcId").
		Joins("inner join act2_proxy as proxy on proxy.idc_id = idc.id").
		Find(&proxyInfos).Error
	return
}

func StatHostProxyInfos() ([]common.HostProxyStatInfo, error) {
	var hostProxyStatInfos []common.HostProxyStatInfo

	err := model.GetDb().Table("act2_host as host").
		Select("count(id) as count, proxy_id as proxyId").
		Group("proxy_id").
		Find(&hostProxyStatInfos).Error
	return hostProxyStatInfos, err
}

// 统计host中idc的信息
func StatHostIdcs() ([]common.HostIdcStatInfo, error) {
	var hostStatInfos []common.HostIdcStatInfo
	err := model.GetDb().Table("act2_host as host").
		Select("count(idc_id) as count, idc_id as idcId").
		Group("idc_id").
		Find(&hostStatInfos).Error
	return hostStatInfos, err
}

// Register 注册接口，用于注册master和agent的信息
//TODO 需要将新上报的Proxy添加到proxyLock当中
func Register(reg common.RegParam) error {
	logger := getLogger()
	logger.Debug("act2master start process register ", "RegParam", dataexchange.ToJsonString(reg))

	var errMsg []string

	idc, proxy, err := ProcessIdcAndProxy(reg.Master)
	if err != nil {
		return err
	}

	var hostIPRows [][]interface{}
	var hostRows [][]interface{}

	hostMap, systemIds, ips := hostToMap(reg.Minion)

	hosts, err := model.FindHostsByEntityIds(systemIds)
	if err != nil {
		return err
	}

	hostIps, err := model.FindHostIPByIps(ips)
	if err != nil {
		return err
	}

	ac2HostMap := act2HostsToMap(hosts)

	for k, minion := range hostMap {
		value, ok := ac2HostMap[k]

		addTime := time.Now()

		hostID := generator.GenUUID()
		if ok {
			// 主机存在
			hostID = value.ID
			addTime = value.AddTime
		}

		if minion.Sn == "" {
			errMsg = append(errMsg, fmt.Sprintf("minion(%s)  sn is null", minion.IPs))
			continue
		}

		hostRows = append(hostRows, []interface{}{
			hostID,
			idc.ID,
			proxy.ID,
			minion.Sn,
			addTime,
			minion.OsType,
			minion.Status,
			minion.MinionVersion,
		})
		// concat two slice
		hostIPRows = append(hostIPRows, processHostIP(minion, hostID, hostIps)...)
	}

	logger.Debug("save host and proxy data", "idc", dataexchange.ToJsonString(idc), "hostIPRows", dataexchange.ToJsonString(hostIPRows))

	if err := saveData(idc, proxy, hostRows, hostIPRows); err != nil {
		logger.Error("fail save register info to db", "error", err)
		return err
	}

	if len(errMsg) > 0 {
		return errors.New(strings.Join(errMsg, "\n"))
	}

	return nil
}

func act2HostsToMap(hosts []model.Act2Host) (act2HostMap map[string]model.Act2Host) {
	act2HostMap = make(map[string]model.Act2Host)
	for _, act2Host := range hosts {
		act2HostMap[act2Host.EntityID] = act2Host
	}
	return
}

func hostToMap(minions []common.Minion) (hostMap map[string]common.Minion, sns []string, ips []string) {
	hostMap = make(map[string]common.Minion)
	for _, minion := range minions {
		hostMap[minion.Sn] = minion
		sns = append(sns, minion.Sn)
		ips = append(ips, minion.IPs...)
	}
	return
}

// 添加保存，事务处理，批量处理
func saveData(idc model.Act2Idc, act2Proxy model.Act2Proxy, hostRows [][]interface{}, hostIPRows [][]interface{}) error {
	logger := getLogger()
	db := model.GetDb()

	tx := db.Begin()

	if err := tx.Save(&idc).Error; err != nil {
		tx.Rollback()
		logger.Error("fail to save Act2Idc", "error", err)
		return err
	}

	if err := tx.Save(&act2Proxy).Error; err != nil {
		tx.Rollback()
		logger.Error("fail to save Act2Proxy", "error", err)
		return err
	}

	// host batch save ; tx is unused for batchput
	if err := batchputs.Put(db.DB(), define.MysqlDriver, model.Act2Host{}.TableName(), "id", model.GetAct2HostColumns(), hostRows); err != nil {
		tx.Rollback()
		logger.Error("fail to save Act2Host", "error", err)
		return err
	}

	// hostIP batch save; tx is unused for batchput
	if err := batchputs.Put(db.DB(), define.MysqlDriver, model.Act2HostIP{}.TableName(), "id", model.GetAct2HostIPColumns(), hostIPRows); err != nil {
		tx.Rollback()
		logger.Error("fail to save Act2HostIP", "error", err)
		return err
	}

	return tx.Commit().Error
}

// 处理hostIP
func processHostIP(minion common.Minion, hostId string, hostIPS []model.Act2HostIP) [][]interface{} {
	logger := getLogger()
	var hostIPRows [][]interface{}

	// exists host ip map ; k,v = HostIP,Act2HostIP
	hostIPMap := make(map[string]model.Act2HostIP)

	for _, hostIP := range hostIPS {
		hostIPMap[hostIP.IP] = hostIP
	}

	for _, IP := range minion.IPs {
		exitstIP := hostIPMap[IP]

		if exitstIP.ID == "" {
			exitstIP.ID = generator.GenUUID()

			hostIPRows = append(hostIPRows, []interface{}{
				generator.GenUUID(),
				hostId,
				IP,
				time.Now(),
			})
		} else {
			logger.Info("host ip already exists.", "HostIP", IP)
		}

	}
	return hostIPRows
}

// ProcessIdcAndProxy 处理Idc 和 proxy
func ProcessIdcAndProxy(master common.Master) (idc model.Act2Idc, proxy model.Act2Proxy, err error) {
	logger := getLogger()
	options, err := json.Marshal(master.Options)
	if err != nil {
		logger.Error("marshal options error," + err.Error())
		return
	}

	if master.Sn == "" {
		err = errors.New("heartbeat proxy sn is null")
		return
	}

	proxyID := master.Sn
	isNotFound := idc.FindIdcByName(master.Idc)

	if isNotFound {
		idc = model.Act2Idc{
			ID:   generator.GenUUID(),
			Name: master.Idc,
		}
	}
	idc.AddTime = time.Now()
	idcID := idc.ID
	isNotFound = proxy.FindByID(master.Sn)

	if !isNotFound {
		proxyID = proxy.ID
		if proxy.Status == define.Fail {
			//通知web hook proxy reconnect
			webhook.TriggerEvent(webhook.EventInfo{
				Event: define.WebHookEventProxyReconnect,
				Payload: common.ProxyPayload{
					ProxyID:     proxyID,
					ProxyServer: master.Server,
					ProxyType:   master.Type,
				},
			})
		}
	} else {
		//通知web hook new proxy
		webhook.TriggerEvent(webhook.EventInfo{
			Event: define.WebHookEventProxyNew,
			Payload: common.ProxyPayload{
				ProxyID:     proxyID,
				ProxyServer: master.Server,
				ProxyType:   master.Type,
			},
		})
	}

	proxy = model.Act2Proxy{
		ID:        proxyID,
		LastTime:  time.Now(),
		TwiceTime: common.GetLastTime(master),
		IdcID:     idcID,
		Server:    master.Server,
		Type:      master.Type,
		Status:    master.Status,
		Options:   string(options),
	}
	return
}

func getHostIPMap(hostIPs []model.Act2HostIP) map[string][]string {
	//根据ip分组
	hostIPMap := make(map[string][]string)

	for _, hostIP := range hostIPs {
		hostIDs, ok := hostIPMap[hostIP.IP]
		if ok {
			hostIDs = append(hostIDs, hostIP.HostID)
			hostIPMap[hostIP.IP] = hostIDs
		} else {
			hostIDs = make([]string, 0, 1)
			hostIDs = append(hostIDs, hostIP.HostID)
			hostIPMap[hostIP.IP] = hostIDs
		}
	}
	return hostIPMap
}

func checkDuplicateHost(hostIPs []model.Act2HostIP) (duplicateData map[string][]string) {
	logger := getLogger()

	//根据ip分组
	hostIPMap := getHostIPMap(hostIPs)

	hostIDs := make([]string, 0, 5)
	for _, mapHostIDs := range hostIPMap {
		if len(mapHostIDs) == 1 {
			continue
		}

		hostIDs = append(hostIDs, mapHostIDs...)
	}

	hosts, err := model.FindHostsByHostIDs(hostIDs)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		logger.Error("find host fail", "error", err)
		return
	}

	hostMap := make(map[string]model.Act2Host)
	for _, host := range hosts {
		hostMap[host.ID] = host
	}

	duplicateData = make(map[string][]string)
	for ip, mapHostIDs := range hostIPMap {
		if len(mapHostIDs) == 1 {
			continue
		}

		dupHosts := make([]string, 0, 5)
		for _, mapHostID := range mapHostIDs {
			dupHosts = append(dupHosts, hostMap[mapHostID].EntityID)
		}
		duplicateData[ip] = dupHosts
	}

	return
}

//ProxyChange 机器的proxy变更
func ProxyChange(proxy string, entityID string) (err error) {
	logger := getLogger()

	act2Proxy, err := model.FindProxyByID(proxy)
	if gorm.IsRecordNotFoundError(err) {
		return errors.New("proxy does not exist")
	}
	if err != nil {
		logger.Error("find act2 proxy fail", "proxyID", proxy)
		return err
	}

	logger.Info("proxy of host changed", "proxyID", proxy, "entityID", entityID)
	err = model.UpdateHostProxyByEntityID(entityID, proxy)
	if err != nil {
		return err
	}

	webhook.TriggerEvent(webhook.EventInfo{
		Event: define.WebHookEventHostChangeProxy,
		Payload: common.HostChangeProxyPayload{
			EntityID:    entityID,
			ProxyServer: act2Proxy.Server,
			ProxyType:   act2Proxy.Type,
		},
	})

	return nil
}

//FindAllHostInfo 获取所有的主机信息
func FindAllHostInfo(condition common.HostInfoCondition) (hostInfoList []model.HostInfo, err error) {
	logger := getLogger()

	hostInfoList, err = model.FindAllHostInfo(condition.EntityID, condition.IDC, condition.ProxyID, condition.IP)
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		logger.Error("query host info fail", "error", err)
		return nil, err
	}

	return hostInfoList, nil
}
