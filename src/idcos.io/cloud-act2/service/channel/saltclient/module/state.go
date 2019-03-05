//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import (
	"errors"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/fileutil"
	"idcos.io/cloud-act2/utils/maps"
	"strings"
	"time"

	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/utils/debug"
	"idcos.io/cloud-act2/utils/promise"

	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/go-hclog"
	"gopkg.in/yaml.v2"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/saltclient"
	serviceCommon "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/utils/cmd"
)

//StateModule salt state run module
type StateModule struct {
	salt *saltclient.SaltClient
}

//NewStateModule 新建
func NewStateModule(salt *saltclient.SaltClient) *StateModule {
	return &StateModule{
		salt: salt,
	}
}

func loadStateResult(resp []byte) (*saltclient.RunResp, error) {
	logger := getLogger()
	result := saltclient.RunResp{}
	err := yaml.Unmarshal(resp, &result)
	if err != nil {
		logger.Error("unmarshal to yaml object fail", "error", err)
		return nil, err
	}

	return &result, nil
}

//Execute 执行
func (s *StateModule) Execute(hosts []serviceCommon.ExecHost, params common.ExecScriptParam, encodingResultChan *common.PartitionResult) (jid string, err error) {
	logger := getLogger()
	if len(hosts) == 0 {
		err = errors.New("主机列表不能为空")
		return
	}
	/*
	* 如果token已过期，则立刻刷新token
	* 如果即将过期，调用结束后刷新token
	 */
	expiredState := s.salt.CheckTokenExpired()
	if expiredState == define.TokenExpired {
		s.salt.FlushToken()
	} else if expiredState == define.TokenWillExpire {
		defer s.salt.FlushToken()
	}

	body, fileName, err := s.buildMinionBody(hosts, params)
	if err != nil {
		logger.Error("build minion body with state fail", "error", err)
		return
	}

	resp, err := s.salt.MinionExecute(body)
	if err != nil {
		logger.Error("exec minion execute fail")
		return
	}

	logger.Trace("apply state", "response", string(resp))

	result, err := loadStateResult(resp)
	if err != nil {
		logger.Error("apply state error", "error", err)
		return
	}
	jid = result.Return[0].Jid

	// 如果jid为空，则需要特殊处理
	if jid == "" {
		logger.Error("run minion found jid is empty")
		return "", errors.New("not found jid")
	}

	promise.NewGoPromise(func(close chan struct{}) {
		s.getExecResult(hosts, jid, fileName, params.Timeout, encodingResultChan, close)
	}, func(message interface{}) {
		logger.Error("promise panic error", "error", message)
	})

	return
}

func (s *StateModule) buildMinionBody(hosts []serviceCommon.ExecHost, params common.ExecScriptParam) (minionBody *saltclient.MinionsPostBody, fileName string, err error) {
	//调用python脚本替换参数
	args := params.Params["args"].(string)

	script, err := renderScript(params.Script, args)
	if err != nil {
		getLogger().Error("script render fail", "script", script, "args", args)
		return
	}

	//将脚本内容存入文件
	fileName, err = fileutil.SaveScriptToFileWithSaltSys(config.Conf.Salt.SYSPath, []byte(script), params.ScriptType)
	if err != nil {
		getLogger().Error("script save to sys path fail", "params", params)
		return
	}

	kwarg := saltclient.StateKwargParam{
		Mods: fileutil.RemoveFileSuffix(fileName),
	}

	entityIds := make([]string, 0, 1000)
	for _, host := range hosts {
		entityIds = append(entityIds, host.EntityID)
	}

	minionBody = &saltclient.MinionsPostBody{
		Fun:     "state.sls",
		Kwarg:   kwarg,
		Tgt:     strings.Join(entityIds, ","),
		TgtType: "list",
	}

	return
}

func renderScript(script, args string) (resultScript string, err error) {
	resultScript = script
	if len(strings.TrimSpace(args)) == 0 {
		return
	}

	argsMap := argsToMap(args)
	saltState := cmd.NewSaltState()
	bytesBuffer, err := saltState.Render(script, argsMap)
	if err != nil {
		return
	}

	resultScript = bytesBuffer.String()

	if len(strings.TrimSpace(resultScript)) == 0 {
		getLogger().Error("script render result is empty", "script", script, "args", dataexchange.ToJsonString(argsMap))
		err = errors.New("script render result is empty")
	}
	return
}

func argsToMap(args string) (argsMap map[string]interface{}) {
	if len(strings.TrimSpace(args)) == 0 {
		return
	}

	argsMap = make(map[string]interface{})

	itemArr := strings.Split(args, "--")
	for _, item := range itemArr {
		if len(strings.TrimSpace(item)) == 0 {
			continue
		}

		attrArr := strings.Split(item, "=")
		argsMap[strings.TrimSpace(attrArr[0])] = strings.TrimSpace(attrArr[1])
	}

	return
}

func getSimpleStateResponse(resp []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := yaml.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getStateResponse(resp []byte) (*saltclient.StateResultResp, error) {
	result := saltclient.StateResultResp{}
	err := yaml.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

func (s *StateModule) sendFailedMessage(message string, partitionResult *common.PartitionResult) {
	res := common.Result{
		Status:  define.Fail,
		Message: message,
	}
	partitionResult.ResultChan <- res
	partitionResult.Close()
}

func (s *StateModule) buildStateResult(resp map[string]interface{}, hostMinion map[string]string) ([]common.MinionResult, int) {
	var minionResults []common.MinionResult
	var hostNum int

	// TODO: 因为时间紧迫，暂时写成这样
	for _, ret := range resp {
		rets := ret.([]interface{})
		for _, minionResult := range rets {
			var mr common.MinionResult
			minionResultMap, ok := minionResult.(map[interface{}]interface{})
			if ok {
				for minionID, ret := range minionResultMap {
					mr.HostID = hostMinion[minionID.(string)]
					statesResult := ret.(map[interface{}]interface{})

					for _, sr := range statesResult {
						stateResult := sr.(map[interface{}]interface{})
						if r, ok := stateResult["result"]; ok {
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
			} else {
				mr.Status = define.Fail
				if minionResultStr, ok := minionResult.(string); ok {
					mr.Stderr = minionResultStr
				} else {
					mr.Stderr = dataexchange.ToJsonString(minionResult)
				}

			}

			minionResults = append(minionResults, mr)
			hostNum = len(minionResults)
		}
	}

	return minionResults, hostNum
}

func (s *StateModule) buildStateResponse(bytes []byte, jid string, timeout int, hostMinion map[string]string) ([]common.MinionResult, int, error) {
	logger := getLogger()

	result, err := getStateResponse(bytes)
	if err != nil {
		return nil, 0, err
	}

	logger.Trace("get minion result", "minion result", fmt.Sprintf("%#v", result.Info))

	var minionResults []common.MinionResult
	for minionID, ret := range result.Info[0].Result {
		minionResults = parseResult(hostMinion, minionID, ret, minionResults)
	}
	hostNum := len(minionResults)
	minionResults = filterResultAndAdd(jid, minionResults, timeout)
	return minionResults, hostNum, nil
}

func (s *StateModule) isClosed(close chan struct{}, partitionResult *common.PartitionResult) bool {
	logger := getLogger()
	select {
	case <-close:
		logger.Info("process is interrupt")
		s.sendFailedMessage("job interrupted", partitionResult)
		return true
	default:
	}
	return false
}

func (s *StateModule) getMinionResults(byts []byte, jid string, timeout int, hostMinion map[string]string, partitionResult *common.PartitionResult) ([]common.MinionResult, int, error) {
	var hostNum int
	var minionResults []common.MinionResult

	logger := getLogger()

	response := string(byts)
	if strings.HasPrefix(response, "return") {
		resp, err := getSimpleStateResponse(byts)
		if err != nil {
			logger.Error("return get simple state response", "error", err)
			return nil, 0, err
		}

		minionResults, hostNum = s.buildStateResult(resp, hostMinion)

	} else {
		var err error
		minionResults, hostNum, err = s.buildStateResponse(byts, jid, timeout, hostMinion)
		if err != nil {
			logger.Error("/job result to yaml fail", "error", err)
			s.sendFailedMessage("/job result to yaml fail,err:"+err.Error(), partitionResult)
			return nil, 0, err
		}
	}
	return minionResults, hostNum, nil
}

func (s *StateModule) getExecResult(execHosts []serviceCommon.ExecHost, jid, fileName string,
	timeout int, partitionResult *common.PartitionResult, close chan struct{}) {

	defer debug.Recover()

	logger := getLogger()

	timer := listenTimeout(timeout, partitionResult)

	defer taskEnding(fileName, timer, jid)

	res := common.Result{
		Status: define.Success,
	}

	hostMinion := extractEntityIDToHostIDMap(execHosts)

	// 使用fib数列的等待时间间隔来处理
	sleepTimeDuration := []time.Duration{
		time.Duration(500), time.Duration(500), time.Duration(1000), time.Duration(1500), time.Duration(2500), time.Duration(4000),
	}
	durationIndex := 0

	for {
		if s.isClosed(close, partitionResult) || partitionResult.IsClosed() {
			return
		}

		bytes, err := s.salt.GetJobResult(jid, sleepTimeDuration[durationIndex]*time.Millisecond)
		durationIndex = (durationIndex + 1) % len(sleepTimeDuration)
		if err != nil {
			if strings.Index(err.Error(), "timeout") != -1 {
				logger.Error("request /jobs fail", "error", err)
			}
			continue
		}
		logger.Debug("get job result", "response body", string(bytes))

		minionResults, hostNum, err := s.getMinionResults(bytes, jid, timeout, hostMinion, partitionResult)
		if err != nil {
			return
		}

		res.MinionResults = minionResults

		logger.Debug("exec result", "host num", hclog.Fmt("%d", hostNum), "exec hosts", hclog.Fmt("%d", len(execHosts)), "minionResults", fmt.Sprintf("%s", minionResults))
		isClose := false
		if hostNum >= len(execHosts) {
			isClose = true
		}
		partitionResult.ResultChan <- res
		//common.SendResultAndClose(res, returnContext, isClose)
		if isClose {
			partitionResult.Close()
			break
		}
	}
}

func parseResult(hostMinion map[string]string, minionID string, ret saltclient.StateMinionResult, minionResults []common.MinionResult) []common.MinionResult {
	logger := getLogger()

	minionResult := common.MinionResult{}
	minionResult.HostID = hostMinion[minionID]

	smReturnMap, ok := ret.Return.(map[interface{}]interface{})
	if !ok {
		logger.Warn("state return not convert map[interface{}]interface{}", "type", reflect.TypeOf(ret.Return), "return", fmt.Sprintf("%+v", ret.Return))
		minionResult.Status = define.Fail
		minionResult.Stderr = string("unknown state return type")
	} else {
		successSmReturnMap := make(map[string]interface{})
		failSmReturnMap := make(map[string]interface{})

		for model, minionRet := range smReturnMap {
			modelStr := fmt.Sprintf("%v", model)

			if minionRetMap, ok := minionRet.(map[interface{}]interface{}); ok {
				modelResult, ok := minionRetMap["result"]
				if ok && modelResult == true {
					successSmReturnMap[modelStr] = maps.MapSafeToMap(minionRetMap)
				} else {
					failSmReturnMap[modelStr] = maps.MapSafeToMap(minionRetMap)
				}
			} else {
				failSmReturnMap[modelStr] = "unknown result type"
			}
		}

		status := define.Success
		if len(successSmReturnMap) > 0 {
			indentBytes, _ := json.MarshalIndent(successSmReturnMap, "", "\t")
			minionResult.Stdout = string(indentBytes)
		}
		if len(failSmReturnMap) > 0 {
			indentBytes, _ := json.MarshalIndent(failSmReturnMap, "", "\t")
			minionResult.Stderr = string(indentBytes)
			status = define.Fail
		}
		minionResult.Status = status
	}
	minionResults = append(minionResults, minionResult)
	return minionResults
}
