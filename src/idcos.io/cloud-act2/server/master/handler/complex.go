//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"encoding/json"
	"errors"
	"idcos.io/cloud-act2/crypto"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/server/common"
	common3 "idcos.io/cloud-act2/service/complex/common"
	"idcos.io/cloud-act2/service/complex/filemigrate"
	"io/ioutil"
	"net/http"
)

func FileMigrateDownload(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	info := filemigrate.MigrateInfo{}
	err = json.Unmarshal(bytes, &info)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	err = filemigrate.IsFileTransfer(info, w)
	if err != nil {
		common.HandleError(w, err)
		return
	}
}

func FileMigrate(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	info := filemigrate.MasterMigrateInfo{}
	err = json.Unmarshal(bytes, &info)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// 需要对密码进行一些特殊的处理
	client := crypto.GetClient()
	if len(info.SourceHost.Password) > 0 {
		info.SourceHost.Password = client.Encode(info.SourceHost.Password)
	}

	if len(info.TargetHost.Password) > 0 {
		info.TargetHost.Password = client.Encode(info.TargetHost.Password)
	}

	err = fileMigrateCheckParam(&info)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	result, err := filemigrate.MasterFileMigrate(info)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, result.Content)
}

func fileMigrateCheckParam(info *filemigrate.MasterMigrateInfo) (err error) {
	err = checkComplexHost(&info.SourceHost)
	if err != nil {
		return err
	}

	err = checkComplexHost(&info.TargetHost)
	if err != nil {
		return err
	}

	if info.Timeout == 0 {
		info.Timeout = 300
	}

	if len(info.SourceFilePath) == 0 || len(info.TargetFilePath) == 0 {
		return errors.New("file path is required")
	}
	return nil
}

func checkComplexHost(hostInfo *common3.ComplexHost) (err error) {
	if hostInfo.Provider == define.MasterTypeSalt && len(hostInfo.EntityID) == 0 {
		return errors.New("entityId is required when provider is salt")
	}

	if hostInfo.Provider == define.MasterTypeSSH {
		if len(hostInfo.Username) == 0 ||
			len(hostInfo.Password) == 0 ||
			len(hostInfo.HostIP) == 0 ||
			len(hostInfo.IdcName) == 0 {
			return errors.New("username and password is required when provider is ssh")
		}
		if len(hostInfo.OsType) == 0 {
			hostInfo.OsType = define.Linux
		}
	}

	if len(hostInfo.Encoding) == 0 {
		hostInfo.Encoding = "utf-8"
	}
	return nil
}
