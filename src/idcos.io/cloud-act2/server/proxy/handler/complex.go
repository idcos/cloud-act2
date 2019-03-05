//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"encoding/json"
	"idcos.io/cloud-act2/server/common"
	"idcos.io/cloud-act2/service/complex/filemigrate"
	"io/ioutil"
	"net/http"
)

//GetRemoteFile 获取远程主机文件
func GetRemoteFile(w http.ResponseWriter, r *http.Request) {
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

	err = filemigrate.PullRemoteFile(info, w)
	if err != nil {
		common.HandleError(w, err)
		return
	}
}

//FileMigrateDownload 文件迁移的文件下载
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

	if info.MasterTransfer {
		err = filemigrate.GetFileByMaster(info, w)
	} else {
		err = filemigrate.PullRemoteFile(info, w)
	}

	if err != nil {
		common.HandleError(w, err)
		return
	}
}

func NotifyDownload(w http.ResponseWriter, r *http.Request) {
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

	result, timeout, err := filemigrate.PushFile(info)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if timeout {
		common.CommonHandleSuccess(w, "timeout")
	} else {
		common.CommonHandleSuccess(w, result)
	}
}
