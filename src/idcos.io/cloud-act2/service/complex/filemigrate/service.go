//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package filemigrate

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/server/common"
	common3 "idcos.io/cloud-act2/service/complex/common"
	"idcos.io/cloud-act2/service/job"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/httputil"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	apiDownload = "/api/v1/complex/file/migrate/download"
	apiPull     = "/api/v1/complex/file/migrate/pull"
	apiNotify   = "/api/v1/complex/file/migrate/notify"
)

func PullRemoteFile(info MigrateInfo, w http.ResponseWriter) (error) {
	logger := getLogger()

	file, err := GetRemoteStream(info)
	if err != nil {
		logger.Error("get remote stream fail", "info", dataexchange.ToJsonString(info), "error", err)
		return err
	}

	_, err = io.Copy(w, file)
	if err != nil {
		logger.Error("io copy fail", "error", err)
		return err
	}

	return nil
}

func GetFileByMaster(info MigrateInfo, w io.Writer) (err error) {
	logger := getLogger()

	url := config.Conf.Act2.ClusterServer + apiDownload

	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	resp, err := httplib.Get(url).Body(bytes).DoRequest()
	if err != nil {
		logger.Error("request get fail", "url", url, "error", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return getMessageByResponse(resp, "master download file fail")
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

func getMessageByResponse(resp *http.Response, errorPrefex string) error {
	logger := getLogger()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(errorPrefex + " and read response body fail")
	}

	jsonResult := common.JSONResult{}
	err = json.Unmarshal(bytes, &jsonResult)
	if err != nil {
		return fmt.Errorf(errorPrefex+" and response body to jsonResult fail,body:%s", string(bytes))
	}

	logger.Error(errorPrefex, "message", jsonResult.Message)
	return fmt.Errorf(errorPrefex+",message:%s", jsonResult.Message)
}

func IsFileTransfer(info MigrateInfo, w http.ResponseWriter) (err error) {
	logger := getLogger()

	proxy := &model.Act2Proxy{}
	sourceProxyID := info.SourceHost.ProxyID
	notFound := proxy.FindByID(sourceProxyID)
	if notFound {
		logger.Error("not found source proxy", "proxyId", sourceProxyID)
		return errors.New("not found source proxy:" + sourceProxyID)
	}

	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}

	url := proxy.Server + apiPull
	resp, err := httplib.Get(url).Body(bytes).DoRequest()
	if err != nil {
		logger.Error("request get fail", "url", url, "error", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return getMessageByResponse(resp, "proxy pull file fail")
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

func MasterFileMigrate(info MasterMigrateInfo) (result common.JSONResult, err error) {
	err = completeComplexHost(&info.SourceHost)
	if err != nil {
		return common.JSONResult{}, err
	}
	err = completeComplexHost(&info.TargetHost)
	if err != nil {
		return common.JSONResult{}, err
	}

	migrateInfo := MigrateInfo{
		SourceHost:     info.SourceHost,
		TargetHost:     info.TargetHost,
		SourceFilePath: info.SourceFilePath,
		TargetFilePath: info.TargetFilePath,
		Timeout:        info.Timeout,
		MasterTransfer: getMasterTransfer(info.SourceHost, info.TargetHost),
	}

	proxy := &model.Act2Proxy{}
	notFound := proxy.FindByID(migrateInfo.TargetHost.ProxyID)
	if notFound {
		return common.JSONResult{}, fmt.Errorf("not found target host proxy,proxyID:%s", migrateInfo.TargetHost.ProxyID)
	}

	url := proxy.Server + apiNotify
	bytes, err := httputil.HttpPost(url, migrateInfo)
	if err != nil {
		return common.JSONResult{}, err
	}

	result = common.JSONResult{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return common.JSONResult{}, err
	}

	return result, nil
}

func getMasterTransfer(sourceHost common3.ComplexHost, targetHost common3.ComplexHost) (transfer bool) {
	if sourceHost.Provider == define.MasterTypeSalt && targetHost.Provider == define.MasterTypeSalt {
		if targetHost.ProxyID == sourceHost.ProxyID {
			return false
		} else {
			return true
		}
	}

	if sourceHost.IdcName == targetHost.IdcName {
		return false
	} else {
		return true
	}
}

func completeComplexHost(complexHost *common3.ComplexHost) (err error) {
	if complexHost.Provider != define.MasterTypeSSH {
		hostInfo, err := job.FindHostInfoByEntityID(complexHost.EntityID)
		if err != nil {
			return err
		}

		complexHost.ProxyID = hostInfo.ProxyID
		complexHost.IdcName = hostInfo.IdcName
	} else {
		proxyID, err := job.FindProxyIDByIdcName(complexHost.IdcName)
		if err != nil {
			return err
		}

		complexHost.ProxyID = proxyID
	}

	return nil
}
