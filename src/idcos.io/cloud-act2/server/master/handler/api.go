//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"fmt"
	"idcos.io/cloud-act2/utils/dataexchange"
	"io/ioutil"
	"net/http"
	"strconv"

	"idcos.io/cloud-act2/crypto"

	"idcos.io/cloud-act2/define"

	"idcos.io/cloud-act2/model"

	"encoding/json"

	"github.com/pkg/errors"
	"idcos.io/cloud-act2/server/common"
	serviceCommon "idcos.io/cloud-act2/service/common"
	service "idcos.io/cloud-act2/service/host"
	"idcos.io/cloud-act2/service/job"
	"idcos.io/cloud-act2/utils"
)

//AllIDCHosts 查询所有idc下所有主机
func AllIDCHosts(w http.ResponseWriter, _ *http.Request) {
	logger := getLogger()

	logger.Debug("start to find AllIDCHosts")

	hosts, err := job.FindAllIDCHostInfo()
	if err != nil {
		common.HandleError(w, err)
		return
	}
	common.CommonHandleSuccess(w, job.UniqueHostInfo(hosts))
}

//IDCHosts 查询idc下所有主机
func IDCHosts(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	idcName := r.URL.Query().Get("idc")

	logger.Debug("start to find proxy and hosts by idcName", "idc", idcName)

	hosts, err := job.FindHostInfoByIDCName(idcName)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	common.CommonHandleSuccess(w, job.UniqueHostInfo(hosts))
}

//HostResult proxy执行结果回调
func HostResult(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	bytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		common.HandleError(w, err)
		return
	}

	logger.Trace("host result", "body", string(bytes))

	var proxyCallBackResult serviceCommon.ProxyCallBackResult
	if err := json.Unmarshal(bytes, &proxyCallBackResult); err != nil {
		logger.Debug("fail to unmarshal value to proxyCallBackResult")
		common.HandleError(w, err)
		return
	}

	// 校验参数信息
	validErr := utils.Validate.Struct(proxyCallBackResult)
	if validErr != nil {
		common.HandleError(w, validErr)
		return
	}

	errChan := make(chan error, 1)
	proxyCallBackResult.ErrChan = errChan

	err = job.HostResultsCallback(proxyCallBackResult)

	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, "")
}

//HostListByIP 根据ip获取主机uuid接口
func HostListByIP(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	req, err := ioutil.ReadAll(r.Body)

	if err != nil {
		logger.Debug("parse request", "error", err)
	}

	hosts, err := service.FindHostsByIPs(req)

	if err != nil {
		common.HandleError(w, err)
		return
	}
	common.CommonHandleSuccess(w, hosts)
}

//FindJobRecordByID 根据JobRecordId获取JobRecord
func FindJobRecordByID(w http.ResponseWriter, r *http.Request) {
	jobRecordID := r.URL.Query().Get("id")

	if jobRecordID == "" {
		common.HandleError(w, errors.New("id can not be null"))
		return
	}

	var record model.JobRecord
	err := record.GetByID(jobRecordID)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	common.CommonHandleSuccess(w, record)
}

//FindRecordResultByID 根据id查找任务执行结果
func FindRecordResultByID(w http.ResponseWriter, r *http.Request) {
	jobRecordID := r.URL.Query().Get("id")

	if jobRecordID == "" {
		common.HandleError(w, errors.New("id can not be null"))
		return
	}

	results, err := job.FindRecordResultsById(jobRecordID)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	common.CommonHandleSuccess(w, results)
}

//FindJobRecordByPage jobRecord的分页查询
func FindJobRecordByPage(w http.ResponseWriter, r *http.Request) {
	pageNo, err := strconv.Atoi(r.URL.Query().Get("pageNo"))
	if err != nil {
		common.HandleError(w, errors.New("pageNo is not int"))
		return
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		common.HandleError(w, errors.New("pageSize is not int"))
		return
	}

	if pageNo == 0 {
		pageNo = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	result, err := job.FindJobRecordPage(int64(pageNo), int64(pageSize))
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, result)
}

//FindRecordResultsByID 根据ip获取主机uuid接口
func FindRecordResultsByID(w http.ResponseWriter, r *http.Request) {

	logger := getLogger()
	jobRecordID := r.URL.Query().Get("jobRecordId")

	if jobRecordID == "" {
		common.HandleError(w, errors.New("jobRecordId can not be null"))
		return
	}

	hostResults, err := job.FindRecordResultsById(jobRecordID)

	if err != nil {
		common.HandleError(w, err)
		return
	}

	logger.Debug("FindRecordResultsById success ", "results", dataexchange.ToJsonString(hostResults))
	common.CommonHandleSuccess(w, hostResults)
}

//UpdateEntityIDByHostID 根据hostId更新entityId的接口
func UpdateEntityIDByHostID(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	var hostInfo struct {
		HostID   string `json:"hostId"`
		EntityID string `json:"entityId"`
		OsType   string `json:"osType"`
		ProxyID  string `json:"proxyId"`
	}
	err := common.ReadJSONRequest(r, &hostInfo)
	if err != nil {
		logger.Error("read json request ", "error", err)
		common.HandleError(w, err)
		return
	}

	hostID := hostInfo.HostID
	entityID := hostInfo.EntityID

	if hostID == "" {
		common.HandleError(w, errors.New("hostId can not be null"))
		return
	}

	if entityID == "" {
		common.HandleError(w, errors.New("entityId can not be null"))
		return
	}
	logger.Debug("updateEntityIdByHostId param", "entityId", entityID, "hostId", hostID)

	var host model.Act2Host
	notFound, err := host.FindByHostID(hostID)
	if notFound {
		common.HandleError(w, errors.New("record not found"))
		return
	}

	if err != nil {
		common.HandleError(w, err)
		return
	}

	originEntityID := host.EntityID

	if originEntityID != entityID {
		logger.Debug(fmt.Sprintf("start to change origin entityId %s to %s, and os type to %s", originEntityID, entityID, hostInfo.OsType))
		host.EntityID = entityID
		host.OsType = hostInfo.OsType
		host.ProxyID = hostInfo.ProxyID
		err = host.Save()
		if err != nil {
			common.HandleError(w, err)
			return
		}
	}
	common.CommonHandleSuccess(w, "")
}

func checkExecParam(execParam serviceCommon.ExecParam) (err error) {
	if execParam.Pattern == define.FileModule {
		return checkFileExecParam(execParam)
	} else if execParam.Pattern == define.ScriptModule {
		return checkScriptParam(execParam)
	}
	return
}

func checkScriptParam(execParam serviceCommon.ExecParam) error {
	logger := getLogger()
	_, ok := execParam.Params["args"].(string)
	if !ok {
		logger.Error("script not have args param")
		return errors.New("script not have args param")
	}
	return nil
}

func checkFileExecParam(execParam serviceCommon.ExecParam) (err error) {
	logger := getLogger()

	_, ok := execParam.Params["target"].(string)
	if !ok {
		logger.Error("params not exists target args")
		return errors.New("params should have target args")
	}
	_, ok = execParam.Params["fileName"].(string)
	if !ok {
		logger.Error("params not exists fileName args")
		return errors.New("params should have fileName args")
	}
	return
}

//ExecByHostIDs 直接通过主机uuid执行接口
func ExecByHostIDs(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	bytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		common.HandleError(w, err)
		return
	}
	var param serviceCommon.ConfJobIDExecParam
	if err := json.Unmarshal(bytes, &param); err != nil {
		logger.Debug("fail to unmarshal value to param", "error", err)
		common.HandleError(w, err)
		return
	}

	// 校验参数信息
	valiErr := utils.Validate.Struct(param)
	if valiErr != nil {
		common.HandleError(w, valiErr)
		return
	}
	valiErr = checkExecParam(param.ExecParam)
	if valiErr != nil {
		common.HandleError(w, valiErr)
		return
	}

	// 对密码进行混淆处理
	if len(param.ExecParam.Password) > 0 {
		client := crypto.GetClient()
		param.ExecParam.Password = client.Encode(param.ExecParam.Password)
	}

	var user string
	value := r.Context().Value("user")
	if value != nil {
		user = value.(string)
	}

	jobRecordID, err := job.ProcessAndExecByID(user, param)
	if err != nil {
		logger.Debug("fail exec job by ids", "error", err)
		common.HandleError(w, err)
		return
	}
	common.CommonHandleSuccess(w, jobRecordID)
}

//ExecByHostIPs 通过ip执行接口
func ExecByHostIPs(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	bytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		common.HandleError(w, err)
		return
	}

	var param serviceCommon.ConfJobIPExecParam
	if err := json.Unmarshal(bytes, &param); err != nil {
		logger.Debug("fail to unmarshal value to param", "error", err)
		common.HandleError(w, err)
		return
	}

	// 校验参数信息
	valiErr := utils.Validate.Struct(param)
	if valiErr != nil {
		common.HandleError(w, valiErr)
		return
	}
	valiErr = checkExecParam(param.ExecParam)
	if valiErr != nil {
		common.HandleError(w, valiErr)
		return
	}

	// 对密码进行混淆处理
	if len(param.ExecParam.Password) > 0 {
		client := crypto.GetClient()
		param.ExecParam.Password = client.Encode(param.ExecParam.Password)
	}

	var user string
	value := r.Context().Value("user")
	if value != nil {
		user = value.(string)
	}

	jobRecordID, err := job.ProcessAndExecByIP(user, param)
	if err != nil {
		logger.Debug("fail exec job by ids", "error", err)
		common.HandleError(w, err)
		return
	}
	common.CommonHandleSuccess(w, jobRecordID)
}

//SystemHeartbeat 系统心跳
func SystemHeartbeat(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	idcName := r.FormValue("idc")
	entityID := r.FormValue("entityId")

	err := job.PullHosts(idcName, entityID)
	if err != nil {
		logger.Debug("fail to pull proxy and minions", "error", err)
		common.HandleError(w, err)
		return
	}
	common.HandleSuccess(w, nil)
}

//Stat 任务统计报表
func JobStat(w http.ResponseWriter, r *http.Request) {
	result, err := job.GetMasterStat()
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, result)
}
