//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"idcos.io/cloud-act2/define"

	"idcos.io/cloud-act2/server/common"
	common2 "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/heartbeat"
	service "idcos.io/cloud-act2/service/host"
	"idcos.io/cloud-act2/service/job"
)

// 回调测试接口
func CallbackTest(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Debug("parse request", "error", err)
		common.HandleError(w, err)
		return
	}
	common.CommonHandleSuccess(w, string(bytes))
}

// 注册接口，用于master agent信息上报
func Register(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	reg := common2.RegParam{}
	err = json.Unmarshal(bytes, &reg)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	err = service.Register(reg)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, "")
}

// 获取idc列表
func FindIdcs(w http.ResponseWriter, _ *http.Request) {

	idcs, err := service.FindIdcs()
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, idcs)
}

// 获取idc列表
func FindAllProxy(w http.ResponseWriter, _ *http.Request) {

	proxyInfos, err := service.FindAllProxy()
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, proxyInfos)
}

// FindIdcProxy 获取idc列表
func FindIdcProxy(w http.ResponseWriter, r *http.Request) {
	idcName := strings.TrimSpace(r.URL.Query().Get("idc"))
	idcNames := []string{idcName}
	proxyInfos, err := job.FindProxyByIdcName(idcNames)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, proxyInfos)
}

// DelProxy 删除proxy
func DelProxy(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	proxyID := strings.TrimSpace(r.URL.Query().Get("id"))
	logger.Debug("delete proxy address", "proxyID", proxyID)

	proxy, err := job.FindProxyByID(proxyID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	proxy.Status = define.Deleted
	err = proxy.Save()
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, "")
}

// MasterHeart 心跳接口，用于master上报自己的状态
func MasterHeart(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	master := common2.Master{}
	err = json.Unmarshal(bytes, &master)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if regErr := heartbeat.MasterHeat(master); regErr != nil {
		common.HandleError(w, regErr)
		return
	}
	common.CommonHandleSuccess(w, "")
}

//UpdateHostProxy 主机的proxy变更
func UpdateHostProxy(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	form := common2.HostProxyChangeInfo{}
	err = json.Unmarshal(bytes, &form)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	proxyID := strings.TrimSpace(form.ProxyID)
	entityID := strings.TrimSpace(form.EntityID)

	if len(proxyID) == 0 {
		common.HandleError(w, errors.New("proxyId is required"))
		return
	}

	if len(entityID) == 0 {
		common.HandleError(w, errors.New("entityId is required"))
		return
	}

	err = service.ProxyChange(proxyID, entityID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, "change proxy success")
}

//FindAllHostInfo 获取所有的主机信息
func FindAllHostInfo(w http.ResponseWriter, r *http.Request) {
	entityID := r.URL.Query().Get("entityId")
	idc := r.URL.Query().Get("idc")
	ip := r.URL.Query().Get("ip")
	proxyID := r.URL.Query().Get("proxyId")
	condition := common2.HostInfoCondition{
		EntityID: entityID,
		IDC:      idc,
		IP:       ip,
		ProxyID:  proxyID,
	}

	hostInfoList, err := service.FindAllHostInfo(condition)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, hostInfoList)
}
