//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"idcos.io/cloud-act2/utils/arrays"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/generator"

	"idcos.io/cloud-act2/utils/debug"
	"idcos.io/cloud-act2/utils/promise"

	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
)

func getHostInfoAndIdc(ips []string, idcs []string, param common.ConfJobIPExecParam) ([]common.HostInfo, []string, error) {
	var hostInfos []common.HostInfo
	var err error

	logger := getLogger()

	switch param.Provider {
	case define.MasterTypeSalt:
		hostInfos, err = saltProvider(ips)
		if err != nil {
			return nil, nil, err
		}

		//salt下的idc需要从host列表里面filter
		idcs = extractHostIdcs(hostInfos)
		idcs = arrays.RemoveDuplicateKey(idcs)
	case define.MasterTypeSSH:
		hostInfos, err = sshProvider(ips, idcs, param)

		if err != nil {
			return nil, nil, err
		}
	case define.MasterTypePuppet:
		hostInfos, err = puppetProvider(ips)
		if err != nil {
			return nil, nil, err
		}

		//salt下的idc需要从host列表里面filter
		idcs = extractHostIdcs(hostInfos)
		idcs = arrays.RemoveDuplicateKey(idcs)
	default:
		logger.Warn(fmt.Sprintf("not support provider [%s]", param.Provider))
		err = errors.New(fmt.Sprintf("not support provider [%s]", param.Provider))
		return nil, nil, err
	}

	return hostInfos, idcs, err
}

func addIdcWhenHostIdcEmpty(execHosts []common.ExecHost) ([]common.ExecHost, error) {
	idcName := ""
	var err error
	var execHost common.ExecHost
	var newExecHosts []common.ExecHost
	for i, _ := range execHosts {
		if len(execHosts[i].IdcName) == 0 {
			execHost, idcName, err = setExecHostIDC(execHosts[i], idcName)
			if err != nil {
				return nil, err
			}
			newExecHosts = append(newExecHosts, execHost)
		} else {
			newExecHosts = append(newExecHosts, execHosts[i])
		}
	}
	return newExecHosts, nil
}

func setExecHostIDC(execHost common.ExecHost, idcName string) (common.ExecHost, string, error) {
	logger := getLogger()

	host := execHost
	if len(idcName) == 0 {
		idc, err := model.FindFirstIDC()
		if err == gorm.ErrRecordNotFound {
			logger.Error("not found default idc")
			return host, "", errors.New("not found default idc")
		}
		if err != nil {
			logger.Error("find act2 idc fail")
			return host, "", err
		}
		host.IdcName = idc.Name

		// 当idc为空的时候，proxy id也可能是空的
		proxies, err := model.FindProxiesByIdcID(idc.ID)
		if err != nil {
			logger.Error("find proxy", "error", err)
			return host, "", err
		}

		if len(proxies) == 0 {
			logger.Error("idc have no proxy")
			return host, "", errors.New("no proxy exist")
		}

		proxy := proxies[0]
		host.ProxyID = proxy.ID

		return host, idc.Name, nil
	}

	host.IdcName = idcName
	return host, idcName, nil
}

// findNotExistIPHosts
func findNotExistIPHosts(hosts []common.ExecHost) ([]common.ExecHost, error) {
	var ips []string
	for _, host := range hosts {
		ips = append(ips, host.HostIP)
	}
	act2HostIPs, err := model.FindHostIPByIps(ips)
	if err != nil {
		return nil, err
	}

	var newIPs []string
	for _, act2HostIP := range act2HostIPs {
		newIPs = append(newIPs, act2HostIP.IP)
	}

	var emptyIPHosts []common.ExecHost
	for _, host := range hosts {
		if !arrays.Contain(host.HostIP, newIPs) {
			emptyIPHosts = append(emptyIPHosts, host)
		}
	}

	return emptyIPHosts, nil
}

type ProxyIdcInfo struct {
	IdcID   string
	IdcName string
	ProxyID string
}

func getProxyIdcInfo() (*ProxyIdcInfo, error) {
	idc, err := model.FindFirstIDC()
	if err != nil {
		return nil, err
	}

	proxies, err := model.FindProxiesByIdcID(idc.ID)
	if err != nil {
		return nil, err
	}

	if len(proxies) == 0 {
		return nil, err
	}

	proxyIdc := ProxyIdcInfo{
		IdcID:   idc.ID,
		IdcName: idc.Name,
		ProxyID: proxies[0].ID,
	}

	return &proxyIdc, nil
}

func saveNotExistIPHosts(hosts []common.ExecHost) error {
	proxyIdcInfo, err := getProxyIdcInfo()
	if err != nil {
		return err
	}


	var hostIPRows [][]interface{}
	var hostRows [][]interface{}
	for _, host := range hosts {
		hostId := generator.GenUUID()
		hostRows = append(hostRows, []interface{}{
			hostId,
			proxyIdcInfo.IdcID,
			"",                  // proxy id
			generator.GenUUID(), // entity id
			time.Now(),
			host.OsType, // os_type
			define.Running,
			"",
		})

		hostIPRows = append(hostIPRows, []interface{}{
			generator.GenUUID(),
			hostId,
			host.HostIP,
			time.Now(),
		})

	}

	err = saveData(hostRows, hostIPRows)

	if err != nil {
		return err
	}

	return nil
}

func ProcessAndExecByIP(user string, param common.ConfJobIPExecParam) (jobRecordId string, err error) {
	logger := getLogger()

	// 如果ip中的idc为空，则使用所有对应的ip，都是用默认的idc
	hosts, err := findNotExistIPHosts(param.ExecHosts)
	if err != nil {
		logger.Error("find not exists ip hosts", "error", err)
		return "", err
	}

	err = saveNotExistIPHosts(hosts)
	if err != nil {
		logger.Error("save not exists ip hosts", "error", err)
		return "", err
	}

	//设置默认的idc
	param.ExecHosts, err = addIdcWhenHostIdcEmpty(param.ExecHosts)
	if err != nil {
		return "", err
	}

	logger.Debug("exec hosts info", "host", dataexchange.ToJsonString(param.ExecHosts))

	var idcs []string
	var ips []string
	for _, execHost := range param.ExecHosts {

		if execHost.HostIP != "" {
			ips = append(ips, execHost.HostIP)
		}

		if execHost.IdcName != "" {
			idcs = append(idcs, execHost.IdcName)
		}
	}

	// 移除重复的idcs
	idcs = arrays.RemoveDuplicateKey(idcs)

	hostInfos, idcs, err := getHostInfoAndIdc(ips, idcs, param)
	if err != nil {
		return "", err
	}

	// save JobRecord
	var jobRecord model.JobRecord
	jobRecord.ID = generator.GenUUID()
	jobRecord.User = user

	logger.Debug("start to save jobRecord", "hostInfos", dataexchange.ToJsonString(hostInfos))

	if err = saveJobRecord(param.Provider, hostInfos, param.ExecParam, &jobRecord, param.Callback, param.ExecuteID, define.Doing); err != nil {
		return
	}

	promise.NewGoPromise(func(chan struct{}) {
		ipExecAsync(param, jobRecord, idcs, hostInfos)
	}, nil)

	return jobRecord.ID, nil
}

func saltProvider(ips []string) (hostInfos []common.HostInfo, err error) {
	logger := getLogger()

	hostInfos, err = findHostInfoByIps(ips, "salt")
	if err != nil {
		logger.Debug("salt provider find host info error", "error", err.Error())
		return
	}
	if len(hostInfos) == 0 {
		err = errors.New(fmt.Sprintf("salt provider find hostInfo list is empty , ips-->%s", ips))
		return
	}
	return
}

func puppetProvider(ips []string) (hostInfos []common.HostInfo, err error) {
	logger := getLogger()

	hostInfos, err = findHostInfoByIps(ips, "puppet")
	if err != nil {
		logger.Debug("salt provider find host info error", "error", err.Error())
		return
	}
	if len(hostInfos) == 0 {
		err = errors.New(fmt.Sprintf("puppet provider find hostInfo list is empty , ips-->%s", ips))
		return
	}
	return
}

func sshProvider(ips []string, idcs []string, param common.ConfJobIPExecParam) ([]common.HostInfo, error) {
	logger := getLogger()

	if err := checkSshParam(param); err != nil {
		logger.Error("check ssh param", "error", err)
		return nil, err
	}

	// 直接搜索idcs对应的proxy，然后下发给对应的proxy去执行

	hostInfos, err := findHostInfoByIpAndIdc(ips, idcs)
	if err != nil {
		logger.Error("ssh provider find host info error", "error", err.Error())
		return nil, err
	}

	logger.Debug("host info", "host info", dataexchange.ToJsonString(hostInfos))

	idcProxyMap, err := getIdcProxyMap(idcs)
	if err != nil {
		logger.Error("not find idc proxy", "error", err)
		return nil, err
	}

	var newHostInfos []common.HostInfo
	for i := 0; i < len(hostInfos); i++ {
		hostInfo := hostInfos[i]
		proxyID, ok := idcProxyMap[hostInfo.IdcID]
		if !ok {
			logger.Error("not found proxy by idc", "idc", hostInfo.IdcID)
			return nil, errors.New("not found proxy by idc")
		}
		hostInfo.ProxyID = proxyID

		newHostInfos = append(newHostInfos, hostInfo)
	}

	logger.Debug("change host proxy", "host info", dataexchange.ToJsonString(newHostInfos))

	// 传入ip，同时传入了idc，但是没有查到结果，此种情况属于要执行ssh的场景
	//if len(newHostInfos) != len(param.ExecHosts) {
	//	// 若查出来的数量不匹配，需要做新增处理
	//	newHostInfos, err = processSsh(param.ExecHosts, newHostInfos)
	//	if err != nil {
	//		logger.Error("process ssh ", "error", err)
	//		return nil, err
	//	}
	//}
	return newHostInfos, nil
}

func getIdcProxyMap(idcNames []string) (idcProxyMap map[string]string, err error) {
	logger := getLogger()

	proxyIDIdcIDs, err := findProxyIDAndIdcIDByIdcNames(idcNames)
	if err != nil {
		logger.Error("query proxy idc fail", "error", err)
		return
	}

	idcProxyMap = make(map[string]string)
	for _, proxyIDIdcID := range proxyIDIdcIDs {
		idcProxyMap[proxyIDIdcID.IdcID] = proxyIDIdcID.ProxyID
	}

	return
}

func ipExecAsync(param common.ConfJobIPExecParam, jobRecord model.JobRecord, idcs []string, hostInfos []common.HostInfo) {
	defer debug.Recover()

	logger := getLogger()

	logger.Debug("start to method ipExecAsync")
	options := map[string]interface{}{
		"password":   param.ExecParam.Password,
		"username":   param.ExecParam.RunAs,
		"scriptType": param.ExecParam.ScriptType,
	}

	if define.FileModule == param.ExecParam.Pattern && define.UrlType == param.ExecParam.ScriptType {
		var urls []string
		if err := json.Unmarshal([]byte(param.ExecParam.Script), &urls); err != nil {
			logger.Error("script unmarshal json error", "error", err.Error())
			return
		}

		for _, url := range urls {
			task, taskSaveErr := saveTask(url, param.ExecParam.Params, jobRecord.ID, param.ExecParam.Pattern, options)
			if taskSaveErr != nil {
				logger.Error("fail to save JobTask", "error", taskSaveErr.Error())
				return
			}
			ipExecPattIdc(idcs, hostInfos, task, param, jobRecord.Timeout)
		}
	} else {
		task, taskSaveErr := saveTask(param.ExecParam.Script, param.ExecParam.Params, jobRecord.ID, param.ExecParam.Pattern, options)
		if taskSaveErr != nil {
			logger.Error("fail to save JobTask", "error", taskSaveErr.Error())
			return
		}

		ipExecPattIdc(idcs, hostInfos, task, param, jobRecord.Timeout)
	}
	return
}

func ipExecPattIdc(idcs []string, hostInfos []common.HostInfo, jobTask model.JobTask, param common.ConfJobIPExecParam, timeout int) {
	logger := getLogger()

	for _, idc := range idcs {
		idcHosts := filterHostsByIdc(idc, hostInfos)

		if len(idcHosts) == 0 {
			logger.Warn("idc has empty hosts, continue it ", "idc", idc)
			continue
		}

		// ip to map
		ipMap := make(map[string]common.ExecHost)

		for _, execHost := range param.ExecHosts {
			ipMap[execHost.HostIP] = execHost
		}

		// 处理JobExecParam
		var execHosts []common.ExecHost

		for _, hostInfo := range idcHosts {
			execHost, ok := ipMap[hostInfo.HostIP]
			if !ok {
				continue
			}

			encoding := execHost.Encoding
			if len(encoding) == 0 && param.Provider == define.MasterTypeSalt {
				encoding = getEncodingByVersionAndOsType(hostInfo.MinionVersion, hostInfo.OsType)
			}

			execHosts = append(execHosts, common.ExecHost{
				HostID:   hostInfo.HostID,
				HostIP:   hostInfo.HostIP,
				EntityID: hostInfo.EntityID,
				IdcName:  hostInfo.IdcName,
				HostPort: execHost.HostPort,
				OsType:   execHost.OsType,
				Encoding: encoding,
				ProxyID:  hostInfo.ProxyID,
			})
		}

		jobExecParam := common.ProxyJobExecParam{
			Provider:  param.Provider,
			Callback:  getCallbackUrl(),
			ExecHosts: execHosts,
			TaskID:    jobTask.ID,
			ExecParam: param.ExecParam,
		}

		promise.NewGoPromise(func(chan struct{}) {
			JobExec(jobExecParam, idc, jobTask, timeout)
		}, nil)
	}
}

/**
处理ssh情况，若库当中不存在，刚进行新增处理
*/
func processSsh(allExecHosts []common.ExecHost, existHosts []common.HostInfo) (hostInfos []common.HostInfo, err error) {

	logger := getLogger()

	var hostMap = make(map[string]common.HostInfo)
	for _, hostInfo := range existHosts {
		hostMap[hostInfo.HostIP] = hostInfo
	}

	// find all idc
	idcMap, err := model.GetAllIdcNameMap()
	if err != nil {
		return
	}

	var hostIPRows [][]interface{}
	var hostRows [][]interface{}

	for _, host := range allExecHosts {

		ip, ok := hostMap[host.HostIP]
		if ok {
			logger.Warn("database have ip record", "ip", ip)
			continue
		}

		idc, ok := idcMap[host.IdcName]

		// idc不存在，则直接异常
		if !ok {
			err = errors.New(fmt.Sprintf("idc [%s] do not exist, please confirm", host.IdcName))
			return
		}

		hostId := generator.GenUUID()

		hostRows = append(hostRows, []interface{}{
			hostId,
			idc.ID,
			"",                  // proxy id
			generator.GenUUID(), // entity id
			time.Now(),
			host.OsType, // os_type
			define.Running,
			"",
		})

		hostIPRows = append(hostIPRows, []interface{}{
			generator.GenUUID(),
			hostId,
			host.HostIP,
			time.Now(),
		})

		hostInfos = append(hostInfos, common.HostInfo{
			HostID:   hostId,
			HostIP:   host.HostIP,
			IdcName:  host.IdcName,
			IdcID:    idc.ID,
			EntityID: "",
		})
	}

	err = saveData(hostRows, hostIPRows)

	if err != nil {
		return
	}

	hostInfos = append(hostInfos, existHosts...)

	return
}
