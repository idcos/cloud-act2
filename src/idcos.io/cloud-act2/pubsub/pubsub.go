//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package pubsub

import (
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/cachedata"
)

//PubSubClient 发布订阅
type PubSubClient interface {
	SendPublish(key string, message interface{}) error
	Subscribe(key string, byteChan chan []byte)
}

func GetPubSubClient() (PubSubClient, error) {
	if config.Conf.PubSub == define.Redis {
		return &cachedata.RedisCacheClient{}, nil
	} else {
		return nil, errors.New("unknown pub sub client:" + config.Conf.PubSub)
	}
}
