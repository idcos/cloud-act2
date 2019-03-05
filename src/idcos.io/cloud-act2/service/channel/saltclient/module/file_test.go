//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import (
	"fmt"
	"testing"

	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/log"
	"idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/saltclient"
	serviceCommon "idcos.io/cloud-act2/service/common"
)

func TestFileDownload(t *testing.T) {
	err := config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")
	if err != nil {
		t.Errorf("file distribution %s", err)
	}

	if config.Conf.IsMaster() {
		log.InitLogger("master")
	} else {
		log.InitLogger("proxy")
	}

	cfg := saltclient.Config{
		Server:   "https://10.0.0.11:8001",
		Username: "salt-api",
		Password: "******",
		Debug:    true,
	}
	saltClient, err := saltclient.NewSaltClient(cfg)
	if err != nil {
		t.Error(err)
		return
	}

	safeChan := common.PartitionResult{}

	go func() {
		fileModule := NewFileModule(saltClient)
		execHosts := []serviceCommon.ExecHost{
			serviceCommon.ExecHost{
				EntityID: "F30823F4-4A35-4CFF-82F9-183E97D73921",
			},
		}
		r, err := fileModule.Execute(execHosts, common.ExecScriptParam{
			Pattern:    "file",
			Script:     "[\"http://10.0.2.2/nfs/node_modules.tar.gz\"]",
			ScriptType: "url",
			RunAs:      "root",
			Password:   "******",
			Params: map[string]interface{}{
				"target": "/tmp/node_modules.tar.gz",
			},
		}, &safeChan)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("jid = %s", r)
	}()

	r := <-safeChan.ResultChan
	fmt.Println(r)

}
