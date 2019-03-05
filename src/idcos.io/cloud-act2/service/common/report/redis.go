//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package report

import (
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/redis"
	"strconv"
)

type RedisRecorder struct {
	client *redis.Client
	prefix string
}

func (r *RedisRecorder) getUniqueKey(status string) string {
	return r.prefix + "|" + status
}

func (r *RedisRecorder) doIncr(key string) (err error) {
	cmd := r.client.Incr(key)
	return cmd.Err()
}

func (r *RedisRecorder) doDecr(key string) (err error) {
	cmd := r.client.Decr(key)
	return cmd.Err()
}

func (r *RedisRecorder) AddSuccess() (err error) {
	key := r.getUniqueKey(define.Success)
	return r.doIncr(key)
}

func (r *RedisRecorder) AddFail() (err error) {
	key := r.getUniqueKey(define.Fail)
	return r.doIncr(key)
}

func (r *RedisRecorder) AddTimeout() (err error) {
	key := r.getUniqueKey(define.Timeout)
	return r.doIncr(key)
}

func (r *RedisRecorder) AddDoing() (err error) {
	key := r.getUniqueKey(define.Doing)
	return r.doIncr(key)
}

func (r *RedisRecorder) SubDoing() (err error) {
	key := r.getUniqueKey(define.Doing)
	return r.doDecr(key)
}

func (r *RedisRecorder) GetRecord(status string) (int64, error) {
	logger := getLogger()

	key := r.getUniqueKey(status)
	cmd := r.client.Get(key)
	result, err := cmd.Result()
	if err != nil {
		logger.Error("get for redis fail", "error", err)
		return 0, err
	}
	count, err := strconv.ParseInt(result, 64, 10)
	if err != nil {
		logger.Error("convert redis num", "error", err)
	}
	return count, nil
}