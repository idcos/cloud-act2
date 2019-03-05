//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package salt

import (
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/utils/httputil"
	"regexp"
)

const (
	minionStartPattern = "^salt/minion/.+/start$"
)

//ProcessEventMsg 处理event消息
func ProcessEventMsg(saltEvent common.SaltEvent) (err error) {

	match := false
	if match, err = matchTag(minionStartPattern, saltEvent.Tag); err != nil {
		return
	}
	if match {
		err = processMinionStart(saltEvent.Data)
		if err != nil {
			return
		}
	}

	return
}

func processMinionStart(data map[string]interface{}) (err error) {
	logger := getLogger()

	id, ok := data["id"]
	if !ok {
		logger.Error("data not found id", "data", data)
		return errors.New("data not found id")
	}

	entityID, ok := id.(string)
	if !ok {
		logger.Error("id of data is not string type", "data", data)
		return errors.New("id of data is not string type")
	}

	body := common.HostProxyChangeInfo{
		ProxyID:  config.ComData.SN,
		EntityID: entityID,
	}
	url := config.Conf.Act2.ClusterServer + "/api/v1/host/proxy"
	bytes, err := httputil.HttpPut(url, body)
	if err != nil {
		logger.Error("http post host change fail", "responseBody", string(bytes))
		return errors.New("host post fail,error:" + err.Error())
	}

	logger.Info("http post host change success", "responseBody", string(bytes))
	return nil
}

func matchTag(pattern string, tag string) (match bool, err error) {
	logger := getLogger().With("pattern", pattern, "tag", tag)

	match, err = regexp.MatchString(pattern, tag)
	if err != nil {
		logger.Error("tag match error")
	}
	return
}
