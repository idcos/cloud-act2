//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"encoding/json"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/redis"
	"idcos.io/cloud-act2/service/common"
	"testing"
)

func TestProcessAndExecByIP(t *testing.T) {
	loadConfig()

	str := `{"callback":"","execHosts":[{"encoding":"utf-8","entityId":"","hostId":"","hostIp":"11.0.0.95","hostPort":22,"idcName":"","osType":"linux"}],"execParam":{"env":null,"extendData":"","params":{"args":"zlong"},"password":"","pattern":"script","runas":"root","script":"while [[ $# -gt 0 ]]; do\nkey=${1%=*}\nvalue=${1#*=}\ncase $key in\n    --name)\n        echo \"name=$value\"\n        ;;\n    *)\n        echo \"${1##*-}\"\nesac\nshift\ndone\n","scriptType":"bash","timeout":300},"executeId":"","provider":"salt"}`

	var param common.ConfJobIPExecParam

	json.Unmarshal([]byte(str), &param)

	ProcessAndExecByIP("test", param)
}

func loadConfig() {
	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")
	model.OpenConn(config.Conf)
}

func TestGetMasterStat(t *testing.T) {
	loadConfig()

	err := redis.InitRedisClient(config.Conf.Redis)
	if err != nil {
		t.Fatal(err)
	}

	result, err := GetMasterStat()
	if err != nil {
		panic(err)
	}

	t.Log(result)
}
