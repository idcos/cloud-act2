//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package model

import (
	"testing"
	"idcos.io/cloud-act2/config"
	"fmt"
)

func TestFindByHostID(t *testing.T) {
	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")

	err := OpenConn(config.Conf)
	if err != nil {
		fmt.Println(err.Error())
	}

	var hostIPS []Act2HostIP

	FindByHostID("minion uuid2", &hostIPS)

	hostIPMap := make(map[string]Act2HostIP)

	for _, hostIP := range hostIPS {
		hostIPMap[hostIP.IP] = hostIP
	}

	hostIP := Act2HostIP{
		ID:     "11111",
		HostID: "minion uuid",
		IP:     "192.168.0.1",
	}

	_hostIP := hostIPMap[hostIP.IP]


	if _hostIP.ID == "" {
		_hostIP.ID = "12333"
	} else {
		fmt.Println("HostIP already exists.")
	}



}
