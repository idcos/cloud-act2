//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
//package saltclient 操作saltapi的客户端
package saltclient

import (
	"fmt"
	"idcos.io/cloud-act2/utils/fileutil"
	"testing"
	"time"

	"idcos.io/cloud-act2/config"
)

func TestCreateFile(t *testing.T) {
	source, err := fileutil.SaveScriptToFileWithSaltSys(config.Conf.Salt.SYSPath, []byte("shell"), "shell")
	if err != nil {
		t.Error(err)
	}

	if len(source) == 0 {
		t.Error("source为空")
	}
}

func TestGetToken(t *testing.T) {
	c := Config{
		Server:        "http://localhost:8090",
		Username:      "saltapi",
		Password:      "******",
		sslSkipVerify: true,
	}
	token, _, err := getToken(c)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(token)
}

func TestAfter(*testing.T) {
	after := time.After(5 * time.Second)

	end := false
	for {
		select {
		case <-after:
			fmt.Println("超时")
			end = true
		default:
		}

		time.Sleep(3 * time.Second)
		fmt.Println("tick")
		if end {
			break
		}
	}
}

func TestFlushToken(t *testing.T) {
	client, err := NewSaltClient(Config{
		Server:   "http://192.168.1.17:8090",
		Username: "salt-api",
		Password: "******",
	})
	if err != nil {
		t.Fatal(err)
	}

	client.headers[authToken] = "1"
	body := MinionsPostBody{
		Fun:     "cmd.script",
		Tgt:     "salt-minion-01,salt-minion-02",
		TgtType: "list",
		Kwarg: map[string]interface{}{
			"source":  "salt://test.sh",
			"args":    "hello world",
			"runas":   "root",
			"timeout": 300,
		},
	}
	bytes, err := client.MinionExecute(&body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Print(string(bytes))
}
