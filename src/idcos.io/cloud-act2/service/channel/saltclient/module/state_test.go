//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"idcos.io/cloud-act2/define"
	"testing"

	"idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/saltclient"
	serviceCommon "idcos.io/cloud-act2/service/common"
)

const args = "--user=zjr --password=******"

func TestStateExec(t *testing.T) {
	config := saltclient.Config{
		Server:   "https://192.168.1.11:8001",
		Username: "saltapi",
		Password: "******",
	}
	client, err := saltclient.NewSaltClient(config)
	if err != nil {
		t.Fatal(err)
	}

	saltModuler := NewStateModule(client)

	hosts := make([]serviceCommon.ExecHost, 0, 2)
	hosts = append(hosts, serviceCommon.ExecHost{EntityID: "6C12A913-756C-4A6B-B149-35E6351BA939"})
	hosts = append(hosts, serviceCommon.ExecHost{EntityID: "0BBFE987-D1A3-4365-89E8-0FAF7F90F5D7"})

	execParams := make(map[string]interface{})
	execParams["args"] = args

	param := common.ExecScriptParam{
		Pattern: "salt",
		Script: `zjr:
  user.present:
    - name: {{user}}
    - hash_password: True
    - password: {{password}}`,
		ScriptType: "sls",
		Params:     execParams,
		Timeout:    300,
	}

	safeChan := &common.ReturnContext{}

	result := common.PartitionResult{}

	jid, err := saltModuler.Execute(hosts, param, &result)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("jid:" + jid)

	go func() {
		for {
			select {
			case <-safeChan.CloseChan:
				goto loop
			default:
			}
		}
	loop:
		fmt.Println("test done")
	}()

	<-safeChan.CloseChan
}

func TestStateLoader(t *testing.T) {
	data := `info:
- Arguments: []
  Function: state.sls
  Minions:
  - 0BBFE987-D1A3-4365-89E8-0FAF7F90F5D7
  - 6C12A913-756C-4A6B-B149-35E6351BA939
  Result:
    0BBFE987-D1A3-4365-89E8-0FAF7F90F5D7:
      out: highstate
      retcode: 0
      return: &id001
        user_|-useradd_|-zlong_|-present:
          __id__: useradd
          __run_num__: 0
          __sls__: 5db9fbb1-c7f0-4edb-bd8d-6f244375e55f
          changes:
            lstchg: 17814
            passwd: XXX-REDACTED-XXX
          comment: Updated user zlong
          duration: 210.433
          name: zlong
          result: true
          start_time: '14:53:21.325455'
      success: true
    6C12A913-756C-4A6B-B149-35E6351BA939:
      out: highstate
      retcode: 0
      return: &id002
        user_|-useradd_|-zlong_|-present:
          __id__: useradd
          __run_num__: 0
          __sls__: 5db9fbb1-c7f0-4edb-bd8d-6f244375e55f
          changes: {}
          comment: User zlong is present and up to date
          duration: 50.927
          name: zlong
          result: true
          start_time: '14:53:20.734556'
      success: true
  StartTime: 2018, Oct 10 14:53:19.687708
  Target: unknown-target
  Target-type: list
  User: root
  jid: '20181010145319687708'
return:
- 0BBFE987-D1A3-4365-89E8-0FAF7F90F5D7: *id001
  6C12A913-756C-4A6B-B149-35E6351BA939: *id002
`

	resp, err := getStateResponse([]byte(data))
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("resp %#v", resp)
}


func TestStateResponse(t *testing.T) {
	r := `
return:
- 58506540-E9D1-439B-B2B0-8CD706CCEA9B:
    user_|-user_add_|-zlong11_|-present:
      __id__: user_add
      __run_num__: 0
      __sls__: afa29e4e-adc9-453b-9d9e-b2865c4747e9
      changes: {}
      comment: User zlong11 is present and up to date
      duration: 33.275
      name: zlong11
      result: true
      start_time: "01:15:32.395389"
`

	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(r), &result)
	if err != nil {
		t.Errorf("error %s", err)
	}

	var minionResults []common.MinionResult


	for _, ret := range result {
		rets := ret.([]interface{})
		for _, minionResult := range rets {
			var mr common.MinionResult
			for minionID, ret := range minionResult.(map[interface{}]interface{}) {
				mr.HostID = minionID.(string)
				statesResult := ret.(map[interface{}]interface{})

				for _, sr := range statesResult {
					stateResult := sr.(map[interface{}]interface{})
					if r, ok := stateResult["result"]; ok {
						fmt.Println(r)
						if r.(bool) {
							mr.Status = define.Success
						} else {
							mr.Status = define.Fail
						}
					} else {
						mr.Status = define.Fail
					}
				}


			}

			//hostNum = len(minionResults)
			//minionResults = filterResultAndAdd(jid, minionResults, timeout)
			//
			minionResults = append(minionResults, mr)
		}
	}
	fmt.Println(minionResults)
}