//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package webhook

import (
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/redis"
	"testing"
	"time"
)

func loadConfig() {
	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")
	model.OpenConn(config.Conf)

	err := redis.InitRedisClient(config.Conf.Redis)
	if err != nil {
		panic(err)
	}
}

func TestAddWebHook(t *testing.T) {
	loadConfig()

	AddWebHook(common.WebHookAddForm{
		Event:define.WebHookEventJobRun,
		Uri:"http://localhost:6868/api/v1/web/hook/test",
		Token:"1111test",
		CustomHeaders: map[string]string{
			"test-key":"test-value",
		},
		CustomData:map[string]interface{}{
			"time":time.Now(),
			"text":"test",
			"number":123,
		},
	})
}

func TestUpdateWebHook(t *testing.T) {
	loadConfig()

	UpdateWebHook(common.WebHookUpdateForm{
		ID:"e616b30e-bc06-8908-7ad9-0c1168b43b91",
		Event:define.WebHookEventJobRun,
		Uri:"http://localhost:6868/api/v1/web/hook/test",
		CustomData:map[string]interface{}{
			"time":time.Now(),
			"text":"test",
			"number":123,
		},
	})
}

func TestDeleteByID(t *testing.T) {
	loadConfig()

	DeleteByID("e616b30e-bc06-8908-7ad9-0c1168b43b91")
}