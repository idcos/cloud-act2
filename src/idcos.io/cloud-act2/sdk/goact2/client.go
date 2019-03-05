//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package goact2

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/httputil"

	"idcos.io/cloud-act2/service/complex/filemigrate"

	"github.com/astaxie/beego/httplib"
	"idcos.io/cloud-act2/define"

	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/service/common"

	"github.com/pkg/errors"
	"idcos.io/cloud-act2/model"
)

type Client struct {
	act2Ctl config.Act2Ctl
}

func NewClient(act2Ctl config.Act2Ctl) *Client {
	client := &Client{
		act2Ctl: act2Ctl,
	}

	client.connect()
	return client
}

func (c *Client) getRequestURL(path string) string {
	cluster := strings.TrimRight(c.act2Ctl.Cluster, "/")
	return cluster + path
}

func (c *Client) Auth() func(req *httplib.BeegoHTTPRequest) *httplib.BeegoHTTPRequest {
	if c.act2Ctl.AuthType == "basic" {
		return func(req *httplib.BeegoHTTPRequest) *httplib.BeegoHTTPRequest {
			return req.SetBasicAuth(c.act2Ctl.Username, c.act2Ctl.Password)
		}
	}
	return func(req *httplib.BeegoHTTPRequest) *httplib.BeegoHTTPRequest {
		return req
	}
}

func (c *Client) connect() error {
	return nil
}

func (c *Client) get(path string) ([]byte, error) {
	url := c.getRequestURL(path)
	return httputil.HttpGet(url, c.Auth())
}

func (c *Client) getQuery(path string, q map[string]string) ([]byte, error) {
	url := c.getRequestURL(path)
	request := httputil.NewDefaultGetRequest(url)
	for key, value := range q {
		request = request.Param(key, value)
	}
	request = httputil.SetRequestOptions(request, c.Auth())

	return request.Bytes()
}

func (c *Client) post(path string, body []byte) ([]byte, error) {
	url := c.getRequestURL(path)
	return httputil.HttpPost(url, body, c.Auth())
}

func (c *Client) delete(path string, body []byte) ([]byte, error) {
	url := c.getRequestURL(path)
	return httputil.HttpDelete(url, body, c.Auth())
}

func (c *Client) DelProxy(id string) (err error) {
	_, err = c.delete("/api/v1/idc/proxy?id="+id, nil)
	return err
}

func (c *Client) IdcList() ([]*model.Act2Idc, error) {
	logger := getLogger()
	byts, err := c.get("/api/v1/idc/all")
	if err != nil {
		logger.Error("request proxies", "error", err)
		return nil, err
	}

	var result struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Content []*model.Act2Idc `json:"content"`
	}

	err = json.Unmarshal(byts, &result)
	if err != nil {
		logger.Error("unmarshal proxies result", "error", err)
		return nil, err
	}
	if result.Status != define.Success {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func (c *Client) IdcProxyList(idc string) ([]*common.ProxyInfo, error) {

	logger := getLogger()
	var bytes []byte
	var err error
	if idc != "" {
		bytes, err = c.get(fmt.Sprintf("/api/v1/idc/proxy?idc=%s", idc))
	} else {
		bytes, err = c.get("/api/v1/idc/proxies")
	}

	if err != nil {
		logger.Error("request proxies", "error", err)
		return nil, err
	}

	var result struct {
		Status  string              `json:"status"`
		Message string              `json:"message"`
		Content []*common.ProxyInfo `json:"content"`
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		logger.Error("unmarshal proxies result", "error", err)
		return nil, err
	}

	return result.Content, nil
}
func (c *Client) IDCHosts(idc string) ([]common.HostInfo, error) {
	var byts []byte
	var err error

	if idc == "" {
		byts, err = c.get("/api/v1/idc/host/all")
	} else {
		byts, err = c.getQuery("/api/v1/idc/host", map[string]string{"idc": idc})
	}

	if err != nil {
		return nil, err
	}

	var result struct {
		Status  string            `json:"status"`
		Message string            `json:"message"`
		Content []common.HostInfo `json:"content"`
	}

	err = json.Unmarshal(byts, &result)
	if err != nil {
		return nil, err
	}

	return result.Content, nil
}

func (c *Client) RecordList(pageNum int, pagesize int) (*common.Pagination, error) {
	byts, err := c.getQuery("/api/v1/job/record/page", map[string]string{
		"pageNo":   fmt.Sprintf("%d", pageNum),
		"pageSize": fmt.Sprintf("%d", pagesize),
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Status  string             `json:"status"`
		Message string             `json:"message"`
		Content *common.Pagination `json:"content"`
	}

	err = json.Unmarshal(byts, &result)
	if err != nil {
		return nil, err
	}

	return result.Content, nil
}

func (c *Client) FindRecordResultByID(id string) ([]model.RecordResult, error) {
	byts, err := c.getQuery("/api/v1/job/record/result", map[string]string{
		"id": id,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Status  string               `json:"status"`
		Message string               `json:"message"`
		Content []model.RecordResult `json:"content"`
	}

	err = json.Unmarshal(byts, &result)
	if err != nil {
		return nil, err
	}

	return result.Content, nil
}

func (c *Client) GetAllHostInfo(entityID, idc, proxyID, ip string) (hostInfoList []model.HostInfo, err error) {
	v := url.Values{}
	if len(entityID) > 0 {
		v.Add("entityId", entityID)
	}
	if len(idc) > 0 {
		v.Add("idc", idc)
	}
	if len(proxyID) > 0 {
		v.Add("proxyId", proxyID)
	}
	if len(ip) > 0 {
		v.Add("ip", ip)
	}

	params := v.Encode()

	var url string
	if len(params) > 0 {
		url = c.getRequestURL("/api/v1/host/all/info?" + params)
	} else {
		url = c.getRequestURL("/api/v1/host/all/info")
	}

	byts, err := httputil.HttpGet(url, c.Auth())
	if err != nil {
		return nil, err
	}

	var result struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Content []model.HostInfo `json:"content"`
	}

	err = json.Unmarshal(byts, &result)
	if err != nil {
		return nil, err
	}

	if result.Status != define.Success {
		return nil, errors.New(result.Message)
	}
	return result.Content, nil
}

func (c *Client) ExecIPJob(param common.ConfJobIPExecParam) (string, error) {
	url := c.getRequestURL("/api/v1/job/ip/exec")
	byts, err := json.Marshal(param)
	if err != nil {
		return "", err
	}

	byts, err = httputil.HttpPost(url, byts, c.Auth())
	if err != nil {
		return "", err
	}

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Content string `json:"content"`
	}

	err = json.Unmarshal(byts, &result)
	if err != nil {
		return "", err
	}

	if result.Status != define.Success {
		return "", errors.New(result.Message)
	}

	return result.Content, nil
}

func (c *Client) GetJobRecord(jobRecordID string) (*JobRecordResult, error) {
	query := url.Values{}
	query.Add("id", jobRecordID)

	url := c.getRequestURL("/api/v1/job/record?" + query.Encode())

	byts, err := httputil.HttpGet(url, c.Auth())
	if err != nil {
		return nil, err
	}

	var result struct {
		Status  string          `json:"status"`
		Message string          `json:"message"`
		Content JobRecordResult `json:"content"`
	}

	err = json.Unmarshal(byts, &result)
	if err != nil {
		return nil, err
	}

	if result.Status != define.Success {
		return nil, errors.New(result.Message)
	}

	return &result.Content, nil
}

func (c *Client) GetJobRecordHostResults(jobRecordID string) ([]HostResult, error) {
	query := url.Values{}
	query.Add("jobRecordId", jobRecordID)
	url := c.getRequestURL("/api/v1/host/result?" + query.Encode())

	byts, err := httputil.HttpGet(url, c.Auth())
	if err != nil {
		return nil, err
	}

	var result struct {
		Status  string       `json:"status"`
		Message string       `json:"message"`
		Content []HostResult `json:"content"`
	}

	err = json.Unmarshal(byts, &result)
	if err != nil {
		return nil, err
	}

	if result.Status != define.Success {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func (c *Client) ReportAgain(IDC string, entityID string) error {
	query := url.Values{}
	query.Add("entityId", entityID)
	query.Add("idc", IDC)

	url := c.getRequestURL("/api/v1/system/heartbeat?" + query.Encode())

	byts, err := httputil.HttpGet(url, c.Auth())
	if err != nil {
		return err
	}

	var result struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Content []interface{} `json:"content"`
	}

	err = json.Unmarshal(byts, &result)
	if err != nil {
		return err
	}

	if result.Status != define.Success {
		return errors.New(result.Message)
	}

	return nil

}

func (c *Client) QuickInstall(ipList []string, osType, idc, username, password, path, masterIP string) error {
	hosts := make([]common.ExecHost, len(ipList))
	for i, ip := range ipList {
		hosts[i] = common.ExecHost{
			HostIP:  ip,
			OsType:  osType,
			IdcName: idc,
		}
	}

	param := common.InstallInfo{
		Hosts:    hosts,
		Username: username,
		Password: password,
		Path:     path,
		MasterIP: masterIP,
	}

	url := c.getRequestURL("/api/v1/quick/install")
	body, err := json.Marshal(param)
	if err != nil {
		return err
	}
	httpConfig := func(req *httplib.BeegoHTTPRequest) *httplib.BeegoHTTPRequest {
		return req.SetTimeout(3*time.Second, 2*time.Hour).Header("Content-Type", "application/json")
	}
	bytes, err := httputil.HttpPost(url, body, c.Auth(), httpConfig)
	if err != nil {
		return err
	}

	var result struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Content []interface{} `json:"content"`
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	if result.Status != define.Success {
		return errors.New(result.Message)
	}

	return nil
}

func (c *Client) FileMigrate(info filemigrate.MasterMigrateInfo) (string, error) {
	url := c.getRequestURL("/api/v1/complex/file/migrate")
	body, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	bytes, err := httputil.HttpPost(url, body, c.Auth())
	if err != nil {
		return "", err
	}

	var result struct {
		Status  string      `json:"status"`
		Message string      `json:"message"`
		Content interface{} `json:"content"`
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return "", err
	}

	if result.Status != define.Success {
		return "", errors.New(result.Message)
	}

	return dataexchange.ToJsonString(result.Content), err
}
