//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
//Package saltclient salt基于saltapi的客户端
package saltclient

import (
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/utils/httputil"
	"strings"
	"time"

	"compress/gzip"
	"io/ioutil"

	"github.com/astaxie/beego/httplib"
	"github.com/pkg/errors"
)

const (
	https              = "https"
	responseErrorStart = "<!"
	response401        = "<title>401 Unauthorized</title>"
	accept             = "Accept"
	applicationJSON    = "application/json"
	applicationYAML    = "application/x-yaml"
	textEventStream    = "text/event-stream"
	authToken          = "X-Auth-Token"
	contentType        = "Content-type"
	cacheControl       = "Cache-Control"
	noCache            = "no-cache"
	connection         = "Connection"
	keepAlive          = "keep-alive"
)

const (
	connectTimeout   = 3 * time.Second
	readWriteTimeout = 300 * time.Second
)

//SaltClient salt的操作客户端
type SaltClient struct {
	config       Config
	headers      map[string]string
	tokenExpired time.Time
}

//NewSaltClient 创建SaltClient
func NewSaltClient(config Config) (client *SaltClient, err error) {
	logger := getLogger()

	//判断是否为https
	config.Server = strings.Trim(config.Server, " ")
	if strings.Index(config.Server, https) == 0 {
		config.sslSkipVerify = true
	}

	token, expired, err := getToken(config)
	if err != nil {
		logger.Error("fail to get token", "error", err)
		return nil, err
	}
	headers := make(map[string]string, 5)
	headers[accept] = applicationJSON
	headers[contentType] = applicationJSON
	headers[authToken] = token
	logger.Debug("get token success", "token", token, "expired", expired)
	return &SaltClient{config, headers, expired}, nil
}

func (client *SaltClient) CheckTokenExpired() (expiredState int) {
	duration := client.tokenExpired.Sub(time.Now())
	if duration.Hours() > 1 {
		expiredState = define.TokenNotExpire
	} else if duration.Hours() < 1 && duration.Minutes() > 1 {
		expiredState = define.TokenWillExpire
	} else {
		expiredState = define.TokenExpired
	}
	return
}

func (client *SaltClient) FlushToken() {
	logger := getLogger()

	token, expired, err := getToken(client.config)
	if err != nil {
		logger.Error(`获取token异常`, "error", err)
		return
	}

	logger.Info("flush token success", "expired", expired)
	client.headers[authToken] = token
	client.tokenExpired = expired
}

func (client *SaltClient) MinionExecute(body *MinionsPostBody) ([]byte, error) {
	return client.minionExecuteRetry(body, 0)
}

func (client *SaltClient) minionExecuteRetry(body *MinionsPostBody, retries int) ([]byte, error) {
	logger := getLogger()

	req := httplib.Post(fmt.Sprintf("%s/minions", client.config.Server)).SetTimeout(connectTimeout, readWriteTimeout)
	req, err := req.JSONBody(body)
	if err != nil {
		logger.Error("exec post body to json fail", "error", err)
		return nil, err
	}
	setHeaders(req, client.headers)
	if client.config.sslSkipVerify {
		httputil.SkipSSL(req)
	}

	byts, err := req.Bytes()
	if err != nil {
		logger.Error("request /minions to exec fail", "error", err)
		return nil, err
	}

	//token有小几率在失效时间之前提前失效,最多重试3次
	if strings.Index(string(byts), responseErrorStart) == 0 && retries < 3 {
		getLogger().Error("request salt api fail", "responseBody", string(byts))
		if client.check401AndFlushToken(byts) {
			setHeaders(req, client.headers)
			retries++
			byts, err = client.minionExecuteRetry(body, retries)
		}
	}

	return byts, nil
}

func (client *SaltClient) check401AndFlushToken(bytes []byte) (ok bool) {
	result := string(bytes)

	if strings.Index(result, response401) >= 0 {
		client.FlushToken()
		ok = true
		return
	}

	ok = false
	return
}

func (client *SaltClient) GetJobResult(jid string, timeout time.Duration) (result []byte, err error) {
	return GetPuller(client).GetJobResult(jid, timeout)
}

//GetEventReq 获取request
func (client *SaltClient) GetEventReq() (req *httplib.BeegoHTTPRequest) {
	req = httplib.Get(fmt.Sprintf("%s/events", client.config.Server))
	setHeaders(req, client.headers)
	req.Header(cacheControl, noCache).
		Header(accept, textEventStream).
		Header(connection, keepAlive)

	req.SetTimeout(time.Duration(3*time.Second), time.Duration(24*365*time.Hour))

	if client.config.sslSkipVerify {
		httputil.SkipSSL(req)
	}

	return
}

func (client *SaltClient) SetHeaders(request *httplib.BeegoHTTPRequest, headers map[string]string) {
	setHeaders(request, headers)
}

func setHeaders(request *httplib.BeegoHTTPRequest, headers map[string]string) {
	for k, v := range headers {
		request.Header(k, v)
	}
}

func getToken(config Config) (token string, expired time.Time, err error) {
	logger := getLogger()

	req := httplib.Post(fmt.Sprintf("%s/login", config.Server)).
		Header(accept, applicationJSON).
		Header(contentType, applicationJSON).
		Param("username", config.Username).
		Param("password", config.Password).
		Param("eauth", "pam")
	if config.sslSkipVerify {
		httputil.SkipSSL(req)
	}

	resp, err := req.DoRequest()
	if err != nil {
		logger.Error("post to get token fail", "error", err)
		return
	}
	defer resp.Body.Close()

	var bytes []byte
	if resp.Header.Get("Content-Encoding") == "gzip" {
		var reader *gzip.Reader
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			logger.Error("gzip reader ", "error", err)
			return
		}
		bytes, err = ioutil.ReadAll(reader)
	} else {
		bytes, err = ioutil.ReadAll(resp.Body)
	}

	if resp.StatusCode >= 400 {
		logger.Error("get token error", "error", string(bytes))
		err = errors.New(fmt.Sprintf("login salt server error %d, plz see proxy log", resp.StatusCode))
		return
	}

	logger.Debug("success get token", "token response", string(bytes), "puppet agent info", config.Server)

	loginResp := LoginResp{}
	err = json.Unmarshal(bytes, &loginResp)
	if err != nil {
		logger.Error("fail to unmarshal json", "error", err)
		return
	}

	token = loginResp.Return[0].Token
	expired = time.Unix(int64(loginResp.Return[0].Expire), 0)
	return
}
