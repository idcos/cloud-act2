//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package report

import (
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/redis"
	"sync/atomic"
)

type (
	Recorder interface {
		AddSuccess() (err error)
		AddFail() (err error)
		AddTimeout() (err error)
		AddDoing() (err error)
		SubDoing() (err error)
		GetRecord(status string) (count int64, err error)
	}

	RamRecorder struct {
		doing   int64
		success int64
		fail    int64
		timeout int64
	}
)

var ramRecorder = &RamRecorder{}

func (recorder *RamRecorder) AddSuccess() (err error) {
	atomic.AddInt64(&recorder.success, 1)
	return
}

func (recorder *RamRecorder) AddFail() (err error) {
	atomic.AddInt64(&recorder.fail, 1)
	return
}

func (recorder *RamRecorder) AddTimeout() (err error) {
	atomic.AddInt64(&recorder.timeout, 1)
	return
}

func (recorder *RamRecorder) AddDoing() (err error) {
	atomic.AddInt64(&recorder.doing, 1)
	return
}

func (recorder *RamRecorder) SubDoing() (err error) {
	atomic.AddInt64(&recorder.doing, -1)
	return
}

func (recorder *RamRecorder) GetRecord(status string) (count int64, err error) {
	switch status {
	case define.Success:
		return recorder.success, nil
	case define.Fail:
		return recorder.fail, nil
	case define.Timeout:
		return recorder.timeout, nil
	case define.Doing:
		return recorder.doing, nil
	default:
		return 0, errors.New("unknown status")
	}
}

//GetRecorder 获取适合的记录者
func GetRecorder(prefix string) Recorder {
	if config.Conf.DependRedis {
		return &RedisRecorder{
			client: redis.GetRedisClient(),
			prefix: prefix,
		}
	}
	return ramRecorder
}
