//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package ssh

import (
	"fmt"
	"idcos.io/cloud-act2/crypto"
	"testing"

	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/log"
	"idcos.io/cloud-act2/service/channel/common"
	common2 "idcos.io/cloud-act2/service/common"
	"os"
)

func TestClient(t *testing.T) {

	loadConfig()
	sshClient := NewSSHClient()

	execHost := common2.ExecHost{
		HostIP:   "192.168.1.217",
		HostPort: 22,
		IdcName:  "杭州",
		OsType:   "linux",
	}

	password := crypto.GetClient().Encode("******")

	e, r, err := sshClient.Execute(execHost, common.ExecScriptParam{
		RunAs:    "root",
		Password: password,
		Script:   "df -h",
		Pattern:  "script",
		Encoding: "utf-8",
	})

	if err != nil {
		t.Errorf("error %s", err)
	}

	fmt.Println(r)
	fmt.Println(e)

}

func TestFileDistribution(t *testing.T) {
	loadConfig()

	if config.Conf.IsMaster() {
		log.InitLogger("master")
	} else {
		log.InitLogger("proxy")
	}
	execHost := common2.ExecHost{
		HostIP:   "192.168.1.17",
		HostPort: 22,
		IdcName:  "杭州",
		OsType:   "linux",
	}

	password := crypto.GetClient().Encode("******")

	sshClient := NewSSHClient()
	e, r, err := sshClient.Execute(execHost, common.ExecScriptParam{
		Pattern:    "file",
		Script:     "[\"http://10.0.2.2/upload/test-srcipts.tar.gz\"]",
		ScriptType: "url",
		RunAs:      "root",
		Password:   password,
		Params: map[string]interface{}{
			"target":   "/tmp/",
			"fileName": "test.tar.gz",
		},
	})
	if err != nil {
		t.Errorf("error %s", err)
	}

	fmt.Println(r)
	fmt.Println(e)

}

func loadConfig() {
	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2-proxy.yaml")
	log.InitLogger("proxy")
}

func TestSSHScript(t *testing.T) {
	loadConfig()

	if config.Conf.IsMaster() {
		log.InitLogger("master")
	} else {
		log.InitLogger("proxy")
	}
	execHost := common2.ExecHost{
		HostIP:   "192.168.1.17",
		HostPort: 22,
		IdcName:  "杭州",
		OsType:   "linux",
	}

	password := crypto.GetClient().Encode("******")

	sshClient := NewSSHClient()
	e, r, err := sshClient.Execute(execHost, common.ExecScriptParam{
		Pattern:    "script",
		Script:     "echo \"$@\"",
		ScriptType: "bash",
		RunAs:      "root",
		Password:   password,
		Params: map[string]interface{}{
			"args": "hello 中文",
		},
		Encoding: "utf8",
	})
	if err != nil {
		t.Errorf("error %s", err)
	}

	fmt.Printf("stdout=%s", r.String())
	fmt.Printf("stderr=%s", e.String())
}

func TestWindowsSSHScript(t *testing.T) {
	loadConfig()

	if config.Conf.IsMaster() {
		log.InitLogger("master")
	} else {
		log.InitLogger("proxy")
	}
	execHost := common2.ExecHost{
		HostIP:   "10.0.10.110",
		EntityID: "xxxxxxxxxxxxxxx",
		HostPort: 22,
		IdcName:  "杭州",
		OsType:   "windows",
	}

	password := os.Getenv("PASSWORD")
	password = crypto.GetClient().Encode(password)

	sshClient := NewSSHClient()
	stdout, stderr, err := sshClient.Execute(execHost, common.ExecScriptParam{
		Pattern:    "script",
		Script:     "@echo hello\r\n@echo 中文测试\r\n@ipconfig\r\n",
		ScriptType: "bat",
		RunAs:      "Administrator",
		Password:   password,
		Params: map[string]interface{}{
			"args": "hello",
		},
		Encoding: "gb18030",
	})
	if err != nil {
		t.Errorf("error %s", err)
	}

	t.Logf("stdout=%s", stdout.String())
	t.Logf("stderr=%s", stderr.String())
}
