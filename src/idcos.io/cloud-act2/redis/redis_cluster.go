//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package redis

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/config"
)

func getClusterSlots(masters []config.RedisServer, slavers []config.RedisServer) ([]redis.ClusterSlot, error) {
	logger := getLogger()

	if len(slavers) != len(masters) {
		logger.Error("redis slaver count not equal master count")
		return nil, errors.New("configuration redis master-slaver error")
	}

	const totalSlots = 16383

	var slots []redis.ClusterSlot
	length := len(masters)
	start := int(0)
	step := totalSlots / length
	if (step * length) < totalSlots {
		step = step + 1
	}

	for index, master := range masters {
		salver := slavers[index]
		nodes := []redis.ClusterNode{
			{Addr: fmt.Sprintf("%s:%v", master.Server, master.Port)},
			{Addr: fmt.Sprintf("%s:%v", salver.Server, salver.Port)},
		}

		logger.Debug("redis cluster", "nodes", fmt.Sprintf("%#v", nodes))

		end := start + step
		if end >= totalSlots {
			end = totalSlots
		}

		slot := redis.ClusterSlot{
			Start: start,
			End:   end,
			Nodes: nodes,
		}

		logger.Debug("redis cluster", "slot", fmt.Sprintf("%#v", slot))

		slots = append(slots, slot)
	}

	return slots, nil
}

// Client
type Client struct {
	redis.Cmdable
	cluster bool
}

// NewClient create new redis client
func NewClient(config config.RedisConfig) (*Client, error) {
	logger := getLogger()

	logger.Info("will start redis with", "cluster", config.Cluster)

	if config.Cluster {
		client := redis.NewClusterClient(&redis.ClusterOptions{
			ClusterSlots: func() ([]redis.ClusterSlot, error) {
				return getClusterSlots(config.Addr.Masters, config.Addr.Slavers)
			},
			RouteRandomly: true,
			Password:      config.Password,
			IdleTimeout:   time.Duration(config.IdleTimeout),
		})
		client.Ping()

		// ReloadState reloads cluster state. It calls ClusterSlots func
		// to get cluster slots information.
		err := client.ReloadState()
		if err != nil {
			return nil, err
		}

		return &Client{Cmdable: client, cluster: true}, nil
	} else {
		if len(config.Addr.Masters) <= 0 {
			return nil, errors.New("not config redis address")
		}
		master := config.Addr.Masters[0]
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%v:%v", master.Server, master.Port),
			Password: config.Password,
			DB:       0,
		})

		return &Client{Cmdable: client, cluster: false}, nil
	}
}

func (c *Client) Subscribe(channel ...string) *redis.PubSub {
	if c.cluster {
		clusterClient := c.Cmdable.(*redis.ClusterClient)
		return clusterClient.Subscribe(channel...)
	} else {
		client := c.Cmdable.(*redis.Client)
		return client.Subscribe(channel...)
	}
}

type Close func()

var (
	redisClient *Client
	once        sync.Once
)

func InitRedisClient(redis config.RedisConfig) error {
	client, err := NewClient(redis)
	if err != nil {
		return err
	}
	redisClient = client
	return nil
}

func GetRedisClient() *Client {
	return redisClient
}
