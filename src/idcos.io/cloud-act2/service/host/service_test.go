//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package host

import (
	"fmt"
	"testing"

	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"encoding/json"
)

func TestRegister(t *testing.T) {

	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")

	err := model.OpenConn(config.Conf)
	if err != nil {
		fmt.Println(err.Error())
	}

	regParam := common.RegParam{}
	bytestr := []byte(`{"master":{"idc":"hangzhou","options":{"password":"******","username":"saltapi"},"server":"http://192.168.1.20:5555","sn":"344323214143","status":"running","type":"salt"},"minions":[{"ips":["10.0.0.0","10.0.0.1"],"sn":"6C12A913-756C-4A6B-B149-35E6351BA938","status":"running"},{"ips":["10.0.1.3","10.0.1.1"],"sn":"","status":"running"}]}`)

	json.Unmarshal(bytestr, regParam)

	err = Register(regParam)
	if err != nil {
		fmt.Println(err.Error())
	}
}
