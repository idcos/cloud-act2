//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package saltclient

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/utils/httputil"
	"time"

	"github.com/astaxie/beego/httplib"
	"gopkg.in/yaml.v2"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/redis"
)

type (
	//Puller salt结果获取的定义
	Puller interface {
		GetJobResult(jid string, timeout time.Duration) (result []byte, err error)
	}
	//RedisPuller redis方式的获取者
	RedisPuller struct {
		client *SaltClient
	}
	//HttpPuller http方式的获取者
	HttpPuller struct {
		client *SaltClient
	}
)

//GetJobResult 获取job执行结果
func (p *RedisPuller) GetJobResult(jid string, timeout time.Duration) ([]byte, error) {
	logger := getLogger()

	key := "ret:" + jid
	logger.Debug("redis pull", "key", key)

	client := redis.GetRedisClient()

	intCmd := client.Exists(key)
	err := intCmd.Err()
	if err != nil {
		logger.Error("redis get exist key", "error", err)
		return nil, err
	}

	v := intCmd.Val()
	// not found
	if v == 0 {
		nanoSeconds := timeout.Nanoseconds()
		time.Sleep(time.Duration(nanoSeconds) * time.Nanosecond)
	}

	intCmd = client.HLen(key)
	length, err := intCmd.Result()
	if err != nil {
		logger.Error("get len result", "error", err, "key", key)
		return nil, err
	}
	if length == 0 {
		return nil, errors.New("get redis result timeout")
	}

	cmd := client.HGetAll(key)
	resultMap, err := cmd.Result()
	if err != nil {
		logger.Error("parser redis replay fail", "error", err)
		return nil, err
	}

	logger.Debug("result map", "result", fmt.Sprintf("%#v", resultMap))

	//组装数据
	translate := map[string]interface{}{}
	for minionID, value := range resultMap {
		mapValue := map[string]interface{}{}
		err = json.Unmarshal([]byte(value), &mapValue)
		if err != nil {
			logger.Error("get job result unmarshal", "error", err, "minionId", minionID)
		}
		translate[minionID] = mapValue["return"]
	}

	type returnListType []map[string]interface{}
	returnMap := map[string]returnListType{
		"return": {translate},
	}

	result, err := yaml.Marshal(returnMap)
	if err != nil {
		logger.Error("return map to result yaml fail", "resultMap", resultMap, "error", err)
		return nil, err
	}
	return result, nil
}

//GetJobResult 获取job执行结果
func (p *HttpPuller) GetJobResult(jid string, timeout time.Duration) (result []byte, err error) {
	logger := getLogger()
	client := p.client

	logger.Debug("start get job result", "jid", jid)

	req := httplib.Get(fmt.Sprintf("%s/jobs/%s", client.config.Server, jid))
	setHeaders(req, client.headers)
	// 要求salt-api使用x-yaml方式返回数据，此时x-yaml会格式化输出结果为二进制数据
	// 不能使用json输出，json存在编码，windows环境下的输出，在json编码输出时会异常
	// 会收到 Could not serialize the return data from Salt. 的错误信息
	req.Header(accept, applicationYAML)
	if client.config.sslSkipVerify {
		httputil.SkipSSL(req)
	}

	return req.Bytes()
}

//GetPuller 获取合适的获取者
func GetPuller(client *SaltClient) Puller {
	if config.Conf.DependRedis {
		return &RedisPuller{
			client: client,
		}
	}
	return &HttpPuller{
		client: client,
	}
}
