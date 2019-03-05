//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package schedule

import (
	"testing"
	"idcos.io/cloud-act2/model"
	"fmt"
	"idcos.io/cloud-act2/config"
)

func TestSchedule(t *testing.T) {

	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")

	err := model.OpenConn(config.Conf)
	if err != nil {
		fmt.Println(err.Error())
	}
	MasterSchedule()

	forever := make(chan bool)
	<-forever
}

func TestScanTimeoutJob(t *testing.T) {
	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")

	err := model.OpenConn(config.Conf)
	if err != nil {
		fmt.Println(err.Error())
	}

	ScanTimeoutJob()
}
