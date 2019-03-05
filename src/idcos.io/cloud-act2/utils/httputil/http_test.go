//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package httputil

import (
	"testing"
	"idcos.io/cloud-act2/service/common"
	"encoding/json"
	"fmt"
)

func TestHttpPost(t *testing.T) {
	body := `{"jobRecordId":"e4e71733-a7fb-0178-c86c-76ceea4a13f4","executeStatus":"done","resultStatus":"fail","hostResults":[{"entityId":"F88E8415-D66E-4519-9F30-7A5D71024271","hostIp":"10.0.0.14","idcName":"yunji-test-idc","status":"fail","message":"post response status fail,url: http://10.0.0.11:5555/execute error message: invalid character '\u003c' looking for beginning of value","stdout":"","stderr":""},{"entityId":"E2145FD0-4678-4AA2-AAC7-C661035AD1DE","hostIp":"10.0.0.13","idcName":"yunji-test-idc","status":"fail","message":"post response status fail,url: http://10.0.0.11:5555/execute error message: invalid character '\u003c' looking for beginning of value","stdout":"","stderr":""}]}`

	var param common.JobCallbackParam

	json.Unmarshal([]byte(body), param)

	HttpPost("http://10.0.0.124:9080/conf/v2/execute/result", param)

}

func TestHttpPut(t *testing.T) {
	fmt.Println(HttpPut("http://localhost:5555/system/heartbeat",nil))

	//fmt.Println(httplib.Get("localhost:5555/"))
}
