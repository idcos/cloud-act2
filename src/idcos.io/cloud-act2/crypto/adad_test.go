//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package crypto

import (
	"encoding/json"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/generator"
	"testing"
	"time"
)

func Test_Chacha20(t *testing.T) {
	loadConfig()
	text := "12345678"
	ciphertext := GetClient().Encode(text)
	t.Log([]byte(ciphertext))
	decodeText, err := GetClient().Decode(ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if decodeText != text {
		t.Error("加密解密失败！", "plaintext", text, "ciphertext", ciphertext, "decodeText", decodeText)
	}
}

func loadConfig() {
	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")
	model.OpenConn(config.Conf)
}

func TestTaskDecode(t *testing.T) {
	loadConfig()

	optionMap := map[string]interface{}{
		"password": GetClient().Encode("123456"),
	}

	taskID := generator.GenUUID()

	jobTask := &model.JobTask{
		ID:        taskID,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Options:   dataexchange.ToJsonString(optionMap),
	}

	err := jobTask.Save()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		model.GetDb().Where("id = ?", taskID).Delete(jobTask)
	}()

	err = jobTask.GetByID(taskID)
	if err != nil {
		t.Fatal(err)
	}
	option := make(map[string]interface{})
	err = json.Unmarshal([]byte(jobTask.Options), &option)
	if err != nil {
		t.Fatal(err)
	}
	password := option["password"]
	pwd := password.(string)
	t.Log([]byte(pwd))
	plainText, err := GetClient().Decode(pwd)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(plainText)
}
