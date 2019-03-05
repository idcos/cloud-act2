//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import (
	"testing"

	"idcos.io/cloud-act2/utils"
	"gopkg.in/yaml.v2"
	"fmt"
)

func TestScriptResultResp(t *testing.T) {
	data := []byte(`
info:
- Arguments:
  - __kwarg__: true
    args: ''
    env: null
    password: ''
    runas: root
    source: salt://e0ba15e9-de67-4bd7-8e74-ec42ca9daa56.sh
    timeout: 300
  Function: cmd.script
  Minions:
  - E2145FD0-4678-4AA2-AAC7-C661035AD1DE
  - F88E8415-D66E-4519-9F30-7A5D71024271
  Result:
    E2145FD0-4678-4AA2-AAC7-C661035AD1DE:
      return: &id001
        pid: 32724
        retcode: 0
        stderr: ''
        stdout: "default via 10.0.0.1 dev eth0 \n10.0.0.0/8 dev eth0 proto kernel\
          \ scope link src 10.0.0.13 \n169.254.0.0/16 dev eth0 scope link metric 1002"
    F88E8415-D66E-4519-9F30-7A5D71024271:
      return: &id002
        pid: 20742
        retcode: 0
        stderr: ''
        stdout: !!binary |
          DQpDOlxXaW5kb3dzXHN5c3RlbTMyXGNvbmZpZ1xzeXN0ZW1wcm9maWxlPmlwY29uZmlnDQoNCldp
          bmRvd3MgSVAgQ29uZmlndXJhdGlvbg0KDQoNCkV0aGVybmV0IGFkYXB0ZXIgsb612MGsvdMgMzoN
          Cg0KICAgQ29ubmVjdGlvbi1zcGVjaWZpYyBETlMgU3VmZml4ICAuIDogDQogICBMaW5rLWxvY2Fs
          IElQdjYgQWRkcmVzcyAuIC4gLiAuIC4gOiBmZTgwOjpjMDQ5OjZkMGQ6MzkyNDozZDZlJTE2DQog
          ICBJUHY0IEFkZHJlc3MuIC4gLiAuIC4gLiAuIC4gLiAuIC4gOiAxMC4wLjEwLjExMA0KICAgU3Vi
          bmV0IE1hc2sgLiAuIC4gLiAuIC4gLiAuIC4gLiAuIDogMjU1LjAuMC4wDQogICBEZWZhdWx0IEdh
          dGV3YXkgLiAuIC4gLiAuIC4gLiAuIC4gOiANCg0KRXRoZXJuZXQgYWRhcHRlciBWUE4gLSBWUE4g
          Q2xpZW50Og0KDQogICBNZWRpYSBTdGF0ZSAuIC4gLiAuIC4gLiAuIC4gLiAuIC4gOiBNZWRpYSBk
          aXNjb25uZWN0ZWQNCiAgIENvbm5lY3Rpb24tc3BlY2lmaWMgRE5TIFN1ZmZpeCAgLiA6IA0KDQpU
          dW5uZWwgYWRhcHRlciBpc2F0YXAue0NDM0ZEOTE2LUQ3NTMtNDE2Ni04MkFELTM3QkZFMkEzNUIy
          MX06DQoNCiAgIE1lZGlhIFN0YXRlIC4gLiAuIC4gLiAuIC4gLiAuIC4gLiA6IE1lZGlhIGRpc2Nv
          bm5lY3RlZA0KICAgQ29ubmVjdGlvbi1zcGVjaWZpYyBETlMgU3VmZml4ICAuIDogDQoNClR1bm5l
          bCBhZGFwdGVyIGlzYXRhcC57M0IwMEYyRDMtMEIwOC00MkY1LTkxNUMtNzhFQTRGRUQxRkZBfToN
          Cg0KICAgTWVkaWEgU3RhdGUgLiAuIC4gLiAuIC4gLiAuIC4gLiAuIDogTWVkaWEgZGlzY29ubmVj
          dGVkDQogICBDb25uZWN0aW9uLXNwZWNpZmljIEROUyBTdWZmaXggIC4gOiANCg0KQzpcV2luZG93
          c1xzeXN0ZW0zMlxjb25maWdcc3lzdGVtcHJvZmlsZT5kaXINCiBWb2x1bWUgaW4gZHJpdmUgQyBo
          YXMgbm8gbGFiZWwuDQogVm9sdW1lIFNlcmlhbCBOdW1iZXIgaXMgQUVCMi1EMTgwDQoNCiBEaXJl
          Y3Rvcnkgb2YgQzpcV2luZG93c1xzeXN0ZW0zMlxjb25maWdcc3lzdGVtcHJvZmlsZQ0KDQoyMDE4
          LzA2LzExICAxMzoyNCAgICA8RElSPiAgICAgICAgICAuDQoyMDE4LzA2LzExICAxMzoyNCAgICA8
          RElSPiAgICAgICAgICAuLg0KMjAxOC8wNi8xMSAgMTM6MjQgICAgPERJUj4gICAgICAgICAgRG9j
          dW1lbnRzDQogICAgICAgICAgICAgICAwIEZpbGUocykgICAgICAgICAgICAgIDAgYnl0ZXMNCiAg
          ICAgICAgICAgICAgIDMgRGlyKHMpICAxMiwzNjAsODU5LDY0OCBieXRlcyBmcmVl
  StartTime: 2018, Sep 28 15:18:43.707108
  Target: F88E8415-D66E-4519-9F30-7A5D71024271,E2145FD0-4678-4AA2-AAC7-C661035AD1DE
  Target-type: list
  User: salt-api
  jid: '20180928151843707108'
return:
- E2145FD0-4678-4AA2-AAC7-C661035AD1DE: *id001
  F88E8415-D66E-4519-9F30-7A5D71024271: *id002
`)

	result, err := getScriptResultResp(data)
	if err != nil {
		t.Error(err)
		return
	}

	mapData := result.Return[0]
	minionRet := mapData["F88E8415-D66E-4519-9F30-7A5D71024271"].(map[interface{}]interface{})

	stdout, err := utils.DecodingTo([]byte(minionRet["stdout"].(string)), "gb18030")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(stdout))

	t.Logf("result length %v", len(result.Return))

	if len(result.Return) != 2 {
		t.Errorf("result return not equal")
		return
	}

}

func TestSLSScriptResultResp(t *testing.T) {
	data := `info:
- Arguments: []
  Function: state.sls
  Minions:
  - 6C12A913-756C-4A6B-B149-35E6351BA939
  Result:
    6C12A913-756C-4A6B-B149-35E6351BA939:
      out: highstate
      retcode: 0
      return: &id001
        user_|-useradd_|-zlong_|-present:
          __id__: useradd
          __run_num__: 0
          __sls__: 6b6e1931-ab29-4b71-af16-1a298522b189
          changes:
            lstchg: 17814
            passwd: XXX-REDACTED-XXX
          comment: Updated user zlong
          duration: 167.497
          name: zlong
          result: true
          start_time: '11:46:03.322594'
      success: true
  StartTime: 2018, Oct 10 11:46:01.872686
  Target: unknown-target
  Target-type: list
  User: root
  jid: '20181010114601872686'
return:
- 6C12A913-756C-4A6B-B149-35E6351BA939: *id001`

	var v map[string]interface{}
	err := yaml.Unmarshal([]byte(data), &v)
	//result, err := getScriptResultResp([]byte(data))
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(v)

	ret := v["return"]
	t.Log(ret)

}

func TestAddPythonHead1(t *testing.T) {
	script := "print('hello world')"
	newScript := addPythonHead(script)
	if newScript != fmt.Sprintf("%s\n%s", pythonHead, script) {
		t.Error("add head fail")
	}
}

func TestAddPythonHead2(t *testing.T) {
	script := "#!/usr/bin/python\nprint('hello world')"
	newScript := addPythonHead(script)
	if script != newScript {
		t.Error("add head fail")
	}
}
