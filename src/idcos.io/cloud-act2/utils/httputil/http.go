//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package httputil

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"idcos.io/cloud-act2/build"
	"idcos.io/cloud-act2/utils/dataexchange"
	"strings"
	"time"

	"idcos.io/cloud-act2/define"

	"github.com/astaxie/beego/httplib"
	serverCommon "idcos.io/cloud-act2/server/common"
)

func IsHttpsUrl(url string) bool {
	return strings.HasPrefix(url, "https")
}

func JsonRequest(req *httplib.BeegoHTTPRequest) *httplib.BeegoHTTPRequest {
	req.Header("Accept", define.ApplicationJSON)
	req.Header("Content-Type", define.ApplicationJSON)
	return req
}

func SkipSSL(req *httplib.BeegoHTTPRequest) *httplib.BeegoHTTPRequest {
	if strings.HasPrefix(req.GetRequest().URL.String(), "https") {
		req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	return req
}

type RequestOption func(req *httplib.BeegoHTTPRequest) *httplib.BeegoHTTPRequest

func SetRequestOptions(req *httplib.BeegoHTTPRequest, options ...RequestOption) *httplib.BeegoHTTPRequest {
	for _, option := range options {
		req = option(req)
	}

	return req
}

func NewRequest(url string, method string, options ...RequestOption) *httplib.BeegoHTTPRequest {
	req := httplib.NewBeegoRequest(url, method)
	return SetRequestOptions(req, options...)
}

func NewDefaultPostRequest(url string) *httplib.BeegoHTTPRequest {
	return NewRequest(url, "POST", SkipSSL, JsonRequest)
}

func NewDefaultGetRequest(url string) *httplib.BeegoHTTPRequest {
	return NewRequest(url, "GET", SkipSSL, JsonRequest)
}

func NewDefaultPutRequest(url string) *httplib.BeegoHTTPRequest {
	return NewRequest(url, "PUT", SkipSSL, JsonRequest)
}

func NewDefaultDeleteRequest(url string) *httplib.BeegoHTTPRequest {
	return NewRequest(url, "DELETE", SkipSSL, JsonRequest)
}

func requestBody(req *httplib.BeegoHTTPRequest, body interface{}) (*httplib.BeegoHTTPRequest, error) {
	if body != nil {
		switch body.(type) {
		case string:
			req = req.Body(body)
		case []byte:
			req = req.Body(body)
		default:
			marshal, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			req = req.Body(marshal)
		}
	}
	return req, nil
}

func Request(req *httplib.BeegoHTTPRequest, body interface{}, options ...RequestOption) (serverCommon.JSONResult, error) {
	logger := getLogger()
	method := req.GetRequest().Method
	url := req.GetRequest().URL.String()

	var jsonResult serverCommon.JSONResult

	req, err := requestBody(req, body)
	if err != nil {
		logger.Error("marshal to json", "error", err)
		return jsonResult, err
	}
	req = SetRequestOptions(req, options...)

	resp, err := req.Bytes()
	if err != nil {
		logger.Error(fmt.Sprintf("fail to %s url", method), "url", url, "error", err)
		return jsonResult, err
	}

	err = json.Unmarshal(resp, &jsonResult)
	if err != nil {
		logger.Error(fmt.Sprintf("%s request response unmarshal", method), "url", url, "error", err)
		return jsonResult, err
	}

	logger.Trace(fmt.Sprintf("%s response", method), "url", url, "response param", dataexchange.ToJsonString(jsonResult))

	if define.Success != jsonResult.Status {
		logger.Debug(fmt.Sprintf("%s response status fail", method), "url", url, "error message", jsonResult.Message)
		return jsonResult, errors.New(jsonResult.Message)
	}

	return jsonResult, nil
}

/**
http post 请求
*/
func HttpPost(url string, bodyParam interface{}, options ...RequestOption) ([]byte, error) {
	logger := getLogger()

	req := NewDefaultPostRequest(url)
	req = SetRequestOptions(req, options...)
	req, err := requestBody(req, bodyParam)
	if err != nil {
		logger.Error("fail to marshal")
		return nil, err
	}

	resp, err := req.Bytes()
	logger.Debug("start to post ", "url", url, "resp", string(resp), "error", err)
	return resp, err
}

func HttpPut(url string, bodyParam interface{}, options ...RequestOption) ([]byte, error) {
	logger := getLogger()

	req := NewDefaultPutRequest(url)
	req = SetRequestOptions(req, options...)
	req, err := requestBody(req, bodyParam)
	if err != nil {
		logger.Error("fail to marshal")
		return nil, err
	}
	return req.Bytes()
}

func HttpGet(url string, options ...RequestOption) ([]byte, error) {
	req := NewDefaultGetRequest(url)
	req = SetRequestOptions(req, options...)
	return req.Bytes()
}

func HttpDelete(url string, bodyParam interface{}, options ...RequestOption) ([]byte, error) {
	logger := getLogger()

	req := NewDefaultDeleteRequest(url)
	req = SetRequestOptions(req, options...)
	req, err := requestBody(req, bodyParam)
	if err != nil {
		logger.Error("fail to marshal")
		return nil, err
	}

	logger.Debug("start to delete ", "url", url, "body", dataexchange.ToJsonString(bodyParam))
	return req.Bytes()
}

func InitHttpLibSetting(appName string, showDebug bool) {
	// 设置全局的httplib的请求参数信息
	userAgent := "idcos " + appName + " c:" + build.Commit + " v:" + build.GitBranch
	defaultSetting := httplib.BeegoHTTPSettings{
		UserAgent:        userAgent,
		ConnectTimeout:   60 * time.Second,
		ReadWriteTimeout: 60 * time.Second,
		Gzip:             true,
		DumpBody:         true,
		ShowDebug:        showDebug,
	}
	httplib.SetDefaultSetting(defaultSetting)

}
