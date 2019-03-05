//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/host"
	"idcos.io/cloud-act2/utils/httputil"
	"idcos.io/cloud-act2/utils/promise"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
)

func FindIdcProxiesByName() ([]common.ProxyInfo, error) {
	infos, err := host.StatHostIdcs()
	if err != nil {
		return nil, err
	}

	proxyInfos, err := host.FindAllProxy()
	if err != nil {
		return nil, err
	}

	count := infos[0].Count
	idcID := infos[0].IdcID

	for _, info := range infos[1:] {
		if info.Count < count {
			count = info.Count
			idcID = info.IdcID
		}
	}

	for _, proxy := range proxyInfos {
		if proxy.IdcID == idcID {
			proxyInfos = append(proxyInfos, proxy)
		}
	}
	return proxyInfos, nil
}

// PingService
type PingService interface {
	Ping() (bool, error)
}

type CmdPing struct {
	host string
}

func NewCmdPing(host string) *CmdPing {
	return &CmdPing{
		host: host,
	}
}

func (p *CmdPing) Ping() (bool, error) {
	cmd := exec.Command("ping", p.host, "-c", "2", "-i", "0.1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}

	if err := cmd.Start(); err != nil {
		return false, err
	}

	v, _ := ioutil.ReadAll(stdout)
	resp := string(v)
	return strings.Contains(resp, "bytes from"), nil
}

type Act2Server struct {
	server string
}

func (srv *Act2Server) ConnectedTest(minionServer string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/host/ping", srv.server)
	bytes, err := httputil.HttpPost(url, map[string]interface{}{
		"host": minionServer,
	})
	if err != nil {
		return false, err
	}

	var r struct {
		Content bool   `json:"content"`
		Message string `json:"message"`
		Status  string `json:"status"`
	}

	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return false, err
	}

	if r.Status == "false" {
		return false, errors.New(r.Message)
	} else {
		return r.Content, nil
	}
}

func SearchProxyConnectedServers(proxyInfos []common.ProxyInfo, minionServer string) []common.ProxyInfo {
	logger := getLogger()

	wg := sync.WaitGroup{}
	wg.Add(len(proxyInfos))

	var connectedProxyInfos []common.ProxyInfo
	mutex := sync.Mutex{}

	for _, proxyInfo := range proxyInfos {
		promise.NewGoPromise(func(chan struct{}) {
			defer wg.Done()
			act2Server := Act2Server{
				server: proxyInfo.Server,
			}
			b, err := act2Server.ConnectedTest(minionServer)
			if err != nil {
				logger.Error("connect test error", "error", err)
				return
			}

			if b {
				mutex.Lock()
				connectedProxyInfos = append(connectedProxyInfos, proxyInfo)
				mutex.Unlock()
			}
		}, nil)
	}
	wg.Wait()

	return connectedProxyInfos

}

// SearchAllConnectedServers 搜索所有可以连接的服务
func SearchAllConnectedServers(minionServer string) ([]common.ProxyInfo, error) {
	proxyInfos, err := host.FindAllProxy()
	if err != nil {
		return nil, err
	}

	connectedServers := SearchProxyConnectedServers(proxyInfos, minionServer)
	return connectedServers, nil
}

func findMinimumStatInfos(statInfos []common.HostIdcStatInfo) common.HostIdcStatInfo {
	minimumStatInfo := statInfos[0]
	for _, statInfo := range statInfos[1:] {
		if statInfo.Count < minimumStatInfo.Count {
			minimumStatInfo = statInfo
		}
	}
	return minimumStatInfo
}

// FindMinimumConnectedIdcProxyServers 查找最少连接数的idc proxy服务器
func FindMinimumConnectedIdcProxyServers(proxyInfos []common.ProxyInfo) ([]string, error) {
	logger := getLogger()
	statInfos, err := host.StatHostIdcs()
	if err != nil {
		logger.Error("stat host proxy info", "error", err)
		return nil, err
	}

	minimumStatInfo := findMinimumStatInfos(statInfos)

	var servers []string
	for _, proxyInfo := range proxyInfos {
		if proxyInfo.IdcID == minimumStatInfo.IdcID {
			servers = append(servers, proxyInfo.Server)
		}
	}

	return servers, nil
}
