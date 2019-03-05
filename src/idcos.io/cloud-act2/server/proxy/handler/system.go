//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"idcos.io/cloud-act2/server/common"
	"idcos.io/cloud-act2/service/heartbeat"
	"idcos.io/cloud-act2/service/proxy"
	"idcos.io/cloud-act2/utils/promise"
	"net/http"
)

func Heartbeat(w http.ResponseWriter, _ *http.Request) {
	logger := getLogger()
	promise.NewGoPromise(func(chan struct{}) {
		err := heartbeat.RegisterSaltInfo(true)
		if err != nil {
			logger.Error("register master ", "error", err)
			return
		}
	}, nil)

	common.HandleSuccess(w, nil)
}

func HostPing(w http.ResponseWriter, r *http.Request) {
	var host struct {
		Host string `json:"host"`
	}
	err := common.ReadJSONRequest(r, &host)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	ping := proxy.NewCmdPing(host.Host)
	result, err := ping.Ping()
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, result)
}
