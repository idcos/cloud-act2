//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package cachedata

import (
	"errors"
	redis2 "github.com/go-redis/redis"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/redis"
	"strconv"
	"time"
)

type CacheClient interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, expiration time.Duration) error
	Del(key string) error
	Exists(key string) (bool, error)
	HLen(key string) (int64, error)
	HGetAll(key string) (map[string]string, error)
}

type (
	RedisCacheClient struct{}
)

func GetCacheClient() (CacheClient, error) {
	if config.Conf.CacheType == define.Redis {
		return &RedisCacheClient{}, nil
	} else {
		return nil, errors.New("unknown cache type:" + config.Conf.CacheType)
	}
}

func (client *RedisCacheClient) SendPublish(key string, message interface{}) error {
	return redis.GetRedisClient().Publish(key, message).Err()
}

func (client *RedisCacheClient) Get(key string) (interface{}, error) {
	return redis.GetRedisClient().Get(key).Bytes()
}

func (client *RedisCacheClient) Set(key string, value interface{}, expiration time.Duration) error {
	return redis.GetRedisClient().Set(key, value, expiration).Err()
}

func (client *RedisCacheClient) Del(key string) error {
	return redis.GetRedisClient().Del(key).Err()
}

func (client *RedisCacheClient) Subscribe(key string, byteChan chan []byte) {
	logger := getLogger()

	pubsub := redis.GetRedisClient().Subscribe(key)
	go func(msgChan <-chan *redis2.Message, byteChan chan []byte) {
		defer pubsub.Close()
		for {
			msg, err := pubsub.Receive()
			if err != nil {
				logger.Error("receive message", "error", err)
				continue
			}

			switch v := msg.(type) {
			case redis2.Message:
				logger.Debug("SUBSCRIBE redis message success", "channel", v.Channel, "message", string(v.Payload))
				byteChan <- []byte(v.Payload)
			case redis2.Subscription:
				logger.Debug("SUBSCRIBE redis Subscription success", "channel", v.Channel, "count", v.Count)
				byteChan <- []byte(strconv.Itoa(v.Count))
			case error:
				logger.Error("SUBSCRIBE redis message error")
			}
		}
	}(pubsub.Channel(), byteChan)
}

func (client *RedisCacheClient) Exists(key string) (bool, error) {
	intCmd := redis.GetRedisClient().Exists(key)
	if intCmd.Err() != nil {
		return false, intCmd.Err()
	}

	return intCmd.Val() == 1, nil
}

func (client *RedisCacheClient) HLen(key string) (int64, error) {
	intCmd := redis.GetRedisClient().HLen(key)
	return intCmd.Result()
}

func (client *RedisCacheClient) HGetAll(key string) (map[string]string, error) {
	return redis.GetRedisClient().HGetAll(key).Result()
}
