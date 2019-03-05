//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	"idcos.io/cloud-act2/build"
	"idcos.io/cloud-act2/config"
)

func publicAddresses() ([]net.IP, error) {
	var ret []net.IP

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return nil, err
		}

		ret = append(ret, ip)
	}

	return ret, nil
}

// Index index information
func Index(w http.ResponseWriter, _ *http.Request) {
	var host string
	var err error

	ips, err := publicAddresses()
	if err != nil {
		host, err = os.Hostname()
		if err != nil {
			host = config.Conf.IDC
		}
	} else {
		var ipStr []string
		for _, ip := range ips {
			ipStr = append(ipStr, ip.String())
		}
		host = strings.Join(ipStr, ",")
	}

	w.Write([]byte(fmt.Sprintf("hello cloud-act2 %s\n", host)))
}

// Status of server
func Status(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("pong"))
}

// Version version of cloud-act2
func Version(w http.ResponseWriter, _ *http.Request) {
	version := map[string]string{
		"apiVersion":   build.GitBranch,
		"buildTime":    build.Date,
		"gitCommit":    build.Commit,
		"architecture": runtime.GOARCH,
		"goVersion":    runtime.Version(),
		"osType":       runtime.GOOS,
	}
	versionData, _ := json.Marshal(version)
	w.Write(versionData)
}
