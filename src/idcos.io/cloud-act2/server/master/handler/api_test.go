//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/redis"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/generator"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/astaxie/beego/httplib"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/log"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/server/common"
	serviceCommon "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/host"
	"idcos.io/cloud-act2/utils"
)

func TestUpdateEntityIdByHostId(t *testing.T) {
	postRequest := httplib.Put("http://localhost:6868/api/v1/host/entity")
	postRequest.Param("hostId", "23143243214").Param("entityId", "34134123")
	fmt.Println(postRequest)
	req, err := postRequest.String()

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(req)
}

// 并发测试 上报
func TestRegister(t *testing.T) {

	err := sqlInit()
	if err != nil {
		t.Fatal(err)
	}
	log.InitLogger("master")

	sn := generator.GenUUID()
	idc := "test" + generator.GenUUID()

	body := `{"master":{"idc":"` + idc + `","options":{"password":"******","username":"salt-api"},"server":"http://192.168.1.17:5555","sn":"` + sn + `","status":"running","type":"salt"},"minions":[{"ips":["192.168.1.217"],"sn":"6C12A913-756C-4A6B-B149-35E6351BA939","status":"running"},{"ips":["192.168.1.218"],"sn":"0BBFE987-D1A3-4365-89E8-0FAF7F90F5D7","status":"running"}]}`

	var reg serviceCommon.RegParam

	err = json.Unmarshal([]byte(body), &reg)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(dataexchange.ToJsonString(reg))

	for i := 0; i < 100; i++ {
		err = host.Register(reg)
		if err != nil {
			t.Fatal(err)
		}
	}

	defer func() {
		model.GetDb().Delete(&model.Act2Idc{}, "name = ?", idc)
		model.GetDb().Delete(&model.Act2Proxy{}, "id = ?", sn)
		model.GetDb().Where("host_id in (?)",model.GetDb().Model(&model.Act2Host{}).Select("id").Where("proxy_id = ?",sn).QueryExpr()).Delete(&model.Act2HostIP{})
		model.GetDb().Delete(&model.Act2Host{}, "proxy_id = ?", sn)
	}()

	count := 0
	err = model.GetDb().Model(&model.Act2Idc{}).Where("name = ?",idc).Count(&count).Error
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("idc save fail")
	}

	err = model.GetDb().Model(&model.Act2HostIP{}).Where("host_id in (?)",model.GetDb().Model(&model.Act2Host{}).Select("id").Where("proxy_id = ?",sn).QueryExpr()).Count(&count).Error
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("host ip save fail")
	}

	err = model.GetDb().Model(&model.Act2Host{}).Where("proxy_id = ?", sn).Count(&count).Error
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("host save fail")
	}

	err = model.GetDb().Model(&model.Act2Proxy{}).Where("id = ?",sn).Count(&count).Error
	if count != 1 {
		t.Fatal("proxy save fail")
	}
}

func TestIdcHosts(t *testing.T) {
	req := httplib.Get("http://localhost:6868/api/v1/idc/host?idc=杭州")
	str, err := req.String()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(str)
}

func sqlInit() error {
	configPath := "/usr/yunji/cloud-act2/etc/cloud-act2.yaml"
	err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	err = model.OpenConn(config.Conf)
	if err != nil {
		return err
	}
	return nil
}

func TestAllIDCHosts(t *testing.T) {
	// sql init
	err := sqlInit()
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/idc/host/all", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AllIDCHosts)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var jsonResult common.JSONResult
	if err := json.Unmarshal(rr.Body.Bytes(), &jsonResult); err != nil {
		t.Errorf("unmarshal to json data, error occur %s", err)
	}

	if jsonResult.Status == define.Success {
		t.Errorf("handler returned wrong status : got %v want %v",
			jsonResult.Status, define.Success)
	}

	t.Log(jsonResult.Content)
}

func TestExecByHostIPs(t *testing.T) {
	execParam := serviceCommon.ConfJobIPExecParam{
		ExecHosts: []serviceCommon.ExecHost{
			{
				HostIP:   "192.168.1.17",
				HostPort: 22,
				EntityID: "",
				HostID:   "",
				IdcName:  "",
				OsType:   "linux",
				Encoding: "utf-8",
			},
		},
		ExecParam: serviceCommon.ExecParam{
			Pattern:    "script",
			Script:     "echo hello world",
			ScriptType: "bash",
			Params: map[string]interface{}{
				"args": "",
			},
			RunAs:      "root",
			Password:   "******",
			Timeout:    100,
			ExtendData: map[string]interface{}{},
		},
		Provider:  "salt",
		Callback:  "",
		ExecuteID: generator.GenUUID(),
	}
	bodyData, err := json.Marshal(&execParam)
	body := bytes.NewBuffer(bodyData)

	//创建一个请求
	req, err := http.NewRequest("POST", "/api/v1/job/exec", body)
	if err != nil {
		t.Fatal(err)
	}

	// 我们创建一个 ResponseRecorder (which satisfies http.ResponseWriter)来记录响应
	rr := httptest.NewRecorder()

	//直接使用HealthCheckHandler，传入参数rr,req
	ExecByHostIPs(rr, req)

	// 检测返回的状态码
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestHostResult(t *testing.T) {
	value := `{"status":"success","message":"","content":{"taskId":"d8724f47-84e3-f9bb-b468-b30682e1e970",
	"hostResults":[{"hostId":"fc962e58-9e08-914d-5eff-fcac7b9af641","status":"success","stdout":"","stderr":"","message":""}]}}`

	// sql init
	err := sqlInit()
	if err != nil {
		t.Fatal(err)
	}

	utils.InitValidate()
	err = redis.InitRedisClient(config.Conf.Redis)
	if err != nil {
		t.Fatal(err)
	}

	body := bytes.NewBufferString(value)

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/api/v1/host/result/callback", body)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HostResult)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var jsonResult common.JSONResult
	if err := json.Unmarshal(rr.Body.Bytes(), &jsonResult); err != nil {
		t.Errorf("unmarshal to json data, error occur %s", err)
	}

	if jsonResult.Status == define.Success {
		t.Errorf("handler returned wrong status : got %v want %v",
			jsonResult.Status, define.Success)
	}
}
