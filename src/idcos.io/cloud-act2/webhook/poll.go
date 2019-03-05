//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package webhook

import (
	"encoding/json"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/generator"
	"idcos.io/cloud-act2/utils/promise"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/astaxie/beego/httplib"
	slave "github.com/dgrr/GoSlaves"
	goRedis "github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/redis"
)

const (
	eventEmpty = "empty"
)

var pool *slave.SlavePool

//EventInfo 事件数据
type EventInfo struct {
	Event   string
	Payload interface{}
}

var eventMQ = make(chan EventInfo, 100)

//TriggerEvent 触发事件
func TriggerEvent(info EventInfo) {
	eventMQ <- info
}

type pullData struct {
	webHook model.Act2WebHook
	info    EventInfo
}

//Load 初始化
func Load() {
	//初始化线程池
	pool = &slave.SlavePool{
		Work: func(obj interface{}) {
			data := obj.(pullData)
			sendEventWithWebHook(data.webHook, data.info)
		},
	}
	pool.Open()

	promise.NewGoPromise(func(chan struct{}) {
		listenEvent()
	}, nil)
}

func listenEvent() {
	for {
		info := <-eventMQ
		processEventInfo(info)
	}
}

func processEventInfo(info EventInfo) {
	webHooks, err := getWebHooks(info.Event)
	if err != nil {
		return
	}

	sendEvent(webHooks, info)
}

func sendEvent(webHooks []model.Act2WebHook, info EventInfo) {
	for _, webHook := range webHooks {
		pool.Serve(pullData{
			webHook: webHook,
			info:    info,
		})
	}
}

func sendEventWithWebHook(webHook model.Act2WebHook, info EventInfo) (err error) {
	logger := getLogger()

	req := httplib.Post(webHook.Uri)

	err = setHeaders(req, webHook)
	if err != nil {
		return err
	}

	body, err := setBody(req, webHook, info)
	if err != nil {
		return err
	}

	resp, err := req.DoRequest()
	if err != nil {
		logger.Error("request web hook url fail", "webHook", dataexchange.ToJsonString(webHook), "error", err)
		return err
	}

	return saveWebHookRecord(req.GetRequest(), body, resp, webHook)
}

func saveWebHookRecord(req *http.Request, reqBody map[string]interface{}, resp *http.Response, webHook model.Act2WebHook) (err error) {
	logger := getLogger()

	//request
	reqHeaders := make(map[string]string, len(req.Header))
	for key, values := range req.Header {
		reqHeaders[key] = values[0]
	}

	reqMap := map[string]interface{}{
		"url":     req.URL.String(),
		"headers": reqHeaders,
		"body":    reqBody,
	}
	reqBytes, err := json.Marshal(reqMap)
	if err != nil {
		logger.Error("request map to json fail", "error", err)
		return err
	}

	//response
	respHeaders := make(map[string]string, len(resp.Header))
	for key, values := range resp.Header {
		respHeaders[key] = values[0]
	}

	var respBody interface{}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("response body read fail", "error", err)
	} else {
		err = json.Unmarshal(bytes, &respBody)
		if err != nil {
			respBody = string(bytes)
		}
	}
	respMap := map[string]interface{}{
		"headers": respHeaders,
		"body":    respBody,
	}
	respBytes, err := json.Marshal(respMap)
	if err != nil {
		logger.Error("response map to json fail", "error", err)
		return err
	}

	record := model.Act2WebHookRecord{
		ID:        generator.GenUUID(),
		WebHookID: webHook.ID,
		SendTime:  time.Now(),
		Event:     webHook.Event,
		Request:   string(reqBytes),
		Response:  string(respBytes),
	}

	err = record.Save()
	if err != nil {
		logger.Error("save web hook record fail")
		return err
	}

	return nil
}

func setHeaders(req *httplib.BeegoHTTPRequest, webHook model.Act2WebHook) (err error) {
	logger := getLogger()

	urlInfo, err := url.Parse(webHook.Uri)
	if err != nil {
		logger.Error("uri parse fail", "url", webHook.Uri, "error", err)
		return err
	}

	//系统相关headers
	req.Header("Content-Type", "application/json")
	req.Header(define.WebHookHeaderEvent, webHook.Event)
	req.Header(define.WebHookHeaderHost, urlInfo.Host)

	//用户定义headers
	if len(webHook.CustomHeaders) > 0 {
		headers := make(map[string]string, 10)
		err = json.Unmarshal([]byte(webHook.CustomHeaders), &headers)
		if err != nil {
			logger.Error("custom headers json to map fail", "data", webHook.CustomHeaders, "error", err)
			return err
		}

		for key, value := range headers {
			req.Header(key, value)
		}
	}

	if len(webHook.Token) > 0 {
		req.Header(define.WebHookHeaderToken, webHook.Token)
	}

	return nil
}

func setBody(req *httplib.BeegoHTTPRequest, webHook model.Act2WebHook, info EventInfo) (body map[string]interface{}, err error) {
	logger := getLogger()

	body = map[string]interface{}{
		"payload": info.Payload,
	}

	if len(webHook.CustomData) > 0 {
		var data interface{}
		err := json.Unmarshal([]byte(webHook.CustomData), &data)
		if err != nil {
			logger.Error("custom data json to interface fail", "data", webHook.CustomData, "error", err)
			return nil, err
		}

		body["customData"] = data
	}

	bytes, err := json.Marshal(body)
	if err != nil {
		logger.Error("body to json fail", "body", body, "error", err)
		return nil, err
	}
	req.Body(bytes)
	return body, nil
}

func getWebHooks(event string) (webHooks []model.Act2WebHook, err error) {
	logger := getLogger()

	client := redis.GetRedisClient()
	if client != nil {
		bytes, err := client.Get(define.RedisWebHookEventPre + event).Bytes()
		if err != nil && err != goRedis.Nil {
			logger.Error("get web hook list by redis fail", "event", event)
			return nil, err
		}

		webHooks = []model.Act2WebHook{}
		err = json.Unmarshal(bytes, &webHooks)
		if err != nil {
			logger.Error("redis event bytes to webHooks fail", "data", string(bytes), "error", err)
			return nil, err
		}

		return webHooks, nil

	}

	//如果redis中不存在，从数据库中读取,并将其放入redis
	webHooks, err = getWebHooksByDB(event)
	if err != nil {
		logger.Error("get web hook by db fail", "error", err)
		return nil, err
	}

	if client != nil {
		err = sendWebHookToRedis(event, webHooks)
		if err != nil {
			return nil, err
		}
	}

	return webHooks, nil

}

func getWebHooksByDB(event string) (webHooks []model.Act2WebHook, err error) {
	logger := getLogger()

	webHooks, err = model.FindWebHookInfoByEvent(event)
	if gorm.IsRecordNotFoundError(err) {
		return []model.Act2WebHook{}, nil
	}
	if err != nil {
		logger.Error("find web hook fail", "event", event, "error", err)
		return nil, err
	}

	return webHooks, nil
}

func sendWebHookToRedis(event string, webHooks []model.Act2WebHook) (err error) {
	logger := getLogger()

	bytes, err := json.Marshal(webHooks)
	if err != nil {
		logger.Error("web hook list to json fail", "error", err)
		return err
	}

	err = redis.GetRedisClient().Set(define.RedisWebHookEventPre+event, bytes, 0).Err()
	if err != nil {
		logger.Error("set redis fail", "error", err)
		return err
	}

	return nil
}

func sendWebHookEmptyToRedis(event string) (err error) {
	logger := getLogger()

	err = redis.GetRedisClient().Set(define.RedisWebHookEventPre+event, []byte(eventEmpty), 0).Err()
	if err != nil {
		logger.Error("set redis fail", "error", err)
		return err
	}

	return nil
}
