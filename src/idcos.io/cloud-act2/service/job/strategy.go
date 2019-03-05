//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"idcos.io/cloud-act2/crypto"
	"idcos.io/cloud-act2/service/common/passwordmask"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/httputil"

	"idcos.io/cloud-act2/utils/debug"
	"idcos.io/cloud-act2/webhook"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/host"
)

// JobExec 作业执行， 机房信息，主机列表信息，有多少个server，执行参数信息
func JobExec(param common.ProxyJobExecParam, idc string, jobTask model.JobTask, timeout int) {
	defer debug.Recover()

	logger := getLogger()

	//triggerEvent
	triggerJobRunEvent(param, jobTask.RecordID)

	//group hosts by proxyID
	groupMap := hostToProxyGroup(param.ExecHosts)

	//proxy map
	proxyMap, err := getProxyMap()
	if err != nil {
		return
	}

	for proxyID, hosts := range groupMap {
		proxy, ok := proxyMap[proxyID]
		if !ok {
			logger.Error("not found proxy", "proxyID", proxyID, "hosts", dataexchange.ToJsonString(hosts))
			continue
		}
		param.ExecHosts = hosts
		if err := commonExec(param, idc, jobTask, timeout, &proxy); err != nil {
			logger.Error("common exec", "error", err, "idc", idc)
		}
	}
}

func triggerJobRunEvent(param common.ProxyJobExecParam, recordID string) {
	param.ExecParam.Password = passwordmask.GetPasswordMask(param.ExecParam.Password)
	//triggerEvent
	webhook.TriggerEvent(webhook.EventInfo{
		Event: define.WebHookEventJobRun,
		Payload: common.JobRunPayload{
			Info:      common.ParamToPayload(param),
			StartTime: time.Now(),
			RecordID:  recordID,
		},
	})
}

func getProxyMap() (proxyMap map[string]common.ProxyInfo, err error) {
	logger := getLogger()

	proxyInfos, err := host.FindAllProxy()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("not found proxy")
		}
		logger.Error("query proxy info fail", "error", err)
		return
	}

	proxyMap = make(map[string]common.ProxyInfo)
	for _, proxyInfo := range proxyInfos {
		proxyMap[proxyInfo.ID] = proxyInfo
	}
	return
}

func hostToProxyGroup(execHosts []common.ExecHost) map[string][]common.ExecHost {
	groupMap := make(map[string][]common.ExecHost)
	for _, execHost := range execHosts {
		hostList, ok := groupMap[execHost.ProxyID]
		if !ok {
			hostList = make([]common.ExecHost, 0, 10)
		}
		hostList = append(hostList, execHost)
		groupMap[execHost.ProxyID] = hostList
	}
	return groupMap
}

func getOsType(execHosts []common.ExecHost) string {
	// 屏蔽 windows下的runas、实时输出等信息
	osType := define.Win
	for _, execHost := range execHosts {
		if execHost.OsType != define.Win {
			osType = define.Linux
			break
		}
	}
	return osType
}

func triggerRemoteTask(param common.ProxyJobExecParam, idc string, jobTask model.JobTask, proxy *common.ProxyInfo) error {
	logger := getLogger()

	if len(param.ExecParam.Password) > 0 {
		client := crypto.GetClient()
		password, err := client.Decode(param.ExecParam.Password)
		if err != nil {
			logger.Error("decode password error", "err", err)
			return err
		}
		param.ExecParam.Password = password
	}

	logger.Info("job exec", "idc", idc, "proxy", proxy.ID)

	url := proxy.Server + define.JobExecUri
	logger.Debug("start post to proxy", "server", url, "param", dataexchange.ToJsonString(param))
	_, proxyErr := httputil.HttpPost(url, param)

	if proxyErr != nil {
		logger.Error("proxy run job ", "error", proxyErr)
		err := model.UpdateTaskProxyProxyID(jobTask.ID, "")
		if err != nil {
			logger.Warn("update task proxy", "error", err, "taskID", jobTask.ID)
		}

		// 更新结果
		callbackResult := common.ProxyCallBackResult{
			Status:  define.Fail,
			Message: fmt.Sprintf("proxy of idc %s not work ", idc),
			Content: common.ProxyResultContent{
				TaskID:     jobTask.ID,
				MasterSend: true,
			},
		}
		err = HostResultsCallback(callbackResult)
		if err != nil {
			logger.Error("common exec host result callback", "error", err)
		}

		return err
	} else {
		err := model.UpdateTaskProxyProxyID(jobTask.ID, proxy.ID)
		if err != nil {
			logger.Warn("update task proxy", "error", err, "taskID", jobTask.ID)
		}
		logger.Info("start proxy execute success")
	}
	return nil
}

func commonExec(param common.ProxyJobExecParam, idc string, jobTask model.JobTask, timeout int, proxy *common.ProxyInfo) error {
	logger := getLogger()

	osType := getOsType(param.ExecHosts)
	if osType == define.Win {
		param.ExecParam.RealTimeOutput = false
		if param.Provider == define.MasterTypeSalt {
			param.ExecParam.RunAs = ""
			param.ExecParam.Password = ""
		}
	}

	// save proxy
	logger.Debug("start to save jobTaskProxy")
	if err := SaveTaskProxy(jobTask.ID, proxy.ID); err != nil {
		logger.Error("save taskProxy error", "error", err.Error())
		return err
	}

	if err := triggerRemoteTask(param, idc, jobTask, proxy); err != nil {
		logger.Error("trigger remote task fail", "error", err)
		return err
	}

	// save hostResults
	logger.Debug("start to save hostResult")
	if err := SaveHostResult(jobTask, param.ExecHosts, define.Doing, proxy.ID); err != nil {
		logger.Error("save hostResult error", "error", err.Error())
		return err
	}

	// 获取hostid列表
	var hostIds []string
	for _, execHost := range param.ExecHosts {
		hostIds = append(hostIds, execHost.HostID)
	}
	hostResults, err := FindHostResultsByHostIds(jobTask.ID, hostIds)
	if err != nil {
		return err
	}

	// save to db
	updateErr := batchUpdateHostResultStatus(hostResults, define.Doing, "", "")
	if updateErr != nil {
		return updateErr
	}

	//监听超时
	ListenTaskTimeout(timeout, jobTask.ID)
	return nil
}

var proxyLock = ProxyLock{
	IdcProxies: make(map[string][]*ProxyStore),
	Lock:       &sync.Mutex{},
}

type ProxyStore struct {
	Index    int32
	IdcID    string
	IdcName  string
	ProxyID  string
	Provider string
	Proxy    common.ProxyInfo
}

type ProxyLock struct {
	Lock       *sync.Mutex
	IdcProxies map[string][]*ProxyStore
}

func (pl *ProxyLock) RemoveProxy(info *common.ProxyInfo) {
	if info == nil {
		return
	}

	proxyStores := pl.IdcProxies[info.IdcName+info.Type]

	pl.Lock.Lock()

	delIndex := 0
	for i := len(proxyStores); i >= 0; i-- {
		if info.ID != proxyStores[i].ProxyID {
			continue
		}
		delIndex = i
		break
	}

	splitProxies := SplitSlice(proxyStores, delIndex)

	Rebuild(splitProxies)

	zero := int32(0)
	atomic.LoadInt32(&zero)

	pl.IdcProxies[info.IdcName+info.Type] = splitProxies

	pl.Lock.Unlock()
}

func Rebuild(old []*ProxyStore) {
	for i, proxy := range old {
		proxy.Index = int32(i)
	}
	return
}

func (pl *ProxyLock) AddProxy(info *common.ProxyInfo) {

	if info == nil {
		return
	}

	proxyStores := pl.GetProxy(info.IdcName, info.Type)

	proxyStores = append(proxyStores, &ProxyStore{
		Index:    int32(len(proxyStores)),
		IdcID:    info.IdcID,
		IdcName:  info.IdcName,
		Provider: info.Type,
		Proxy:    *info,
		ProxyID:  info.ID,
	})

	zero := int32(0)
	atomic.LoadInt32(&zero)

	pl.Lock.Lock()
	pl.IdcProxies[info.IdcName+info.Type] = proxyStores
	pl.Lock.Unlock()

}

func (pl *ProxyLock) GetProxy(idc, provider string) []*ProxyStore {
	if len(pl.IdcProxies) == 0 {
		err := pl.mapperProxy()
		getLogger().Debug("all proxy value is", "proxies", dataexchange.ToJsonString(pl.IdcProxies), "error", err)
	}

	proxies, _ := pl.IdcProxies[idc+provider]

	return proxies
}

func (pl *ProxyLock) mapperProxy() error {
	proxies, err := host.FindAllProxy()
	if err != nil {
		return err
	}
	idcProvideMap := make(map[string][]*ProxyStore)
	for _, proxyInfo := range proxies {
		existProxies, _ := idcProvideMap[proxyInfo.IdcName+proxyInfo.Type]
		existProxies = append(existProxies, &ProxyStore{
			Index:    int32(len(existProxies)),
			IdcID:    proxyInfo.IdcID,
			IdcName:  proxyInfo.IdcName,
			Provider: proxyInfo.Type,
			Proxy:    proxyInfo,
			ProxyID:  proxyInfo.ID,
		})
		idcProvideMap[proxyInfo.IdcName+proxyInfo.Type] = existProxies
		// 兼容ssh方式执行，等于puppet、salt等都可以处理ssh
		idcProvideMap[proxyInfo.IdcName+"ssh"] = existProxies
	}
	pl.Lock.Lock()
	pl.IdcProxies = idcProvideMap
	pl.Lock.Unlock()
	return nil
}

func SplitSlice(slices []*ProxyStore, delIndex int) []*ProxyStore {
	m := slices[0:delIndex]
	c := slices[delIndex+1:]
	return append(m, c...)
}
