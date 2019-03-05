//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"idcos.io/cloud-act2/server/common"
	"idcos.io/cloud-act2/service/proxy"
	"idcos.io/cloud-act2/utils/arrays"
	"net/http"
	"net/url"
	"strings"
)

// ProxyAllocHandler proxy alloc handler
func ProxyAllocHandler(w http.ResponseWriter, r *http.Request) {
	var minionServer struct {
		Minion string `json:"host"`
	}
	err := common.ReadJSONRequest(r, &minionServer)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	proxyInfos, err := proxy.SearchAllConnectedServers(minionServer.Minion)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	servers, err := proxy.FindMinimumConnectedIdcProxyServers(proxyInfos)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	servers = arrays.RemoveDuplicateKey(servers)

	var ips []string
	for _, server := range servers {
		parse, err := url.Parse(server)
		if err != nil {
			continue
		}

		host := parse.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}

		ips = append(ips, host)
	}

	common.CommonHandleSuccess(w, ips)
	return
}
