//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package heartbeat

import (
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"sync"
	"testing"
)

func TestMasterHeat(t *testing.T) {

	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")

	model.OpenConn(config.Conf)
	str := `{"sn":"344323214143","server":"https://10.0.0.24","options":{"username":"saltapi","password":"******"},"status":"running","idc":"湖南机房","type":"salt"}`

	regParam := common.Master{}

	json.Unmarshal([]byte(str), &regParam)

	//MasterHeat(regParam)

	var wait sync.WaitGroup
	for i := 0; i < 1000; i++ {
		fmt.Println(i)
		wait.Add(1)

		go func() {
			MasterHeat(regParam)

			defer wait.Done()
		}()
	}
	wait.Wait()

}
