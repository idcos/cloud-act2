//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import (
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/crypto"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/encoding"
	"idcos.io/cloud-act2/utils/fileutil"
	"idcos.io/cloud-act2/utils/ints"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/utils/promise"

	"regexp"

	"github.com/pkg/errors"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/saltclient"
	serviceCommon "idcos.io/cloud-act2/service/common"
)

//ScriptModule 脚本执行模块
type ScriptModule struct {
	salt *saltclient.SaltClient
}

//NewScriptModule 新建脚本执行模块
func NewScriptModule(salt *saltclient.SaltClient) *ScriptModule {
	return &ScriptModule{
		salt: salt,
	}
}

//Execute 执行
func (s *ScriptModule) Execute(hosts []serviceCommon.ExecHost, params common.ExecScriptParam,
	partitionResult *common.PartitionResult) (string, error) {
	logger := getLogger()

	if len(hosts) == 0 {
		err := errors.New("主机列表不能为空")
		return "", err
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

	// 构建参数列表
	fileName, body, err := s.buildMinionBody(hosts, params)
	if err != nil {
		logger.Error("execution script", "error", err)
		return "", err
	}

	logger.Trace("salt script", "fileName", fileName, "body", dataexchange.ToJsonString(body))

	byts, err := s.salt.MinionExecute(body)
	if err != nil {
		logger.Error("call minion execute", "error", err)
		return "", err
	}

	byts, err = encoding.DecodingTo(byts, params.Encoding)
	if err != nil {
		logger.Error("decoding minion execute", "error", err, "decoding", params.Encoding)
		return "", err
	}

	logger.Debug("minion", "response", string(byts))

	result := saltclient.RunResp{}
	err = json.Unmarshal(byts, &result)
	if err != nil {
		logger.Error("/minions script response to json fail", "error", err)
		return "", err
	}
	jid := result.Return[0].Jid

	// 如果jid为空，则需要特殊处理
	if jid == "" {
		logger.Error("run minion found jid is empty")
		return "", errors.New("not found jid")
	}

	promise.NewGoPromise(func(close chan struct{}) {
		s.getExecResult(hosts, jid, params, fileName, partitionResult, close)
	}, func(message interface{}) {
		logger.Error("get exec result fail", "error", message)
	})
	return jid, nil
}

func getScriptResultResp(resp []byte) (*saltclient.ScriptResultResp, error) {
	result := saltclient.ScriptResultResp{}
	err := yaml.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ScriptModule) getScriptRunResult(jid string, timeout time.Duration) (result *saltclient.ScriptResultResp, isTimeout bool, err error) {
	logger := getLogger()
	byts, err := s.salt.GetJobResult(jid, timeout)
	if err != nil {
		if strings.Index(err.Error(), "timeout") == -1 {
			logger.Error("request /jobs fail", "error", err)
			return nil, false, err
		}
		return nil, true, err
	}
	logger.Trace("get job result", "response body", string(byts))
	if byts == nil {
		return nil, false, nil
	}

	result, err = getScriptResultResp(byts)
	if err != nil {
		logger.Error("/job result to yaml fail", "error", err)
		return
	}

	return result, false, nil
}

func (s *ScriptModule) buildMinionResult(result *saltclient.ScriptResultResp, hostMinion map[string]string, encodingType string) []common.MinionResult {
	logger := getLogger()
	minionResults := make([]common.MinionResult, 0, 1000)
	for minionID, ret := range result.Return[0] {
		minionResult := common.MinionResult{}
		minionResult.HostID = hostMinion[minionID]

		if retValue, ok := ret.(string); ok {
			minionResult.Status = define.Fail
			minionResult.Message = retValue
		} else if minionRet, ok := ret.(map[interface{}]interface{}); ok {
			if retcode, ok := minionRet["retcode"]; !ok {
				minionResult.Status = define.Fail
				minionResult.Message = "unknown minion result"
			} else {
				retcodeValue, err := ints.GetInt(retcode)
				if err != nil {
					minionResult.Status = define.Fail
				} else {
					if retcodeValue == 0 {
						minionResult.Status = define.Success
					} else {
						minionResult.Status = define.Fail
					}
				}

				// 需要将远程的输出结果，进行转码处理
				stdout, err := encoding.DecodingTo([]byte(minionRet["stdout"].(string)), encodingType)
				if err != nil {
					logger.Warn("minion result stdout could not convert to", "encoding", encodingType, "warn", err)
					// 不能转码，用旧的原先的数据
					stdout = []byte(minionRet["stdout"].(string))
				}

				stderr, err := encoding.DecodingTo([]byte(minionRet["stderr"].(string)), encodingType)
				if err != nil {
					logger.Warn("minion result stderr could not convert to", "encoding", encodingType, "warn", err)
					// 不能转码，用旧的原先的数据
					stderr = []byte(minionRet["stderr"].(string))
				}

				minionResult.Stdout = string(stdout)
				minionResult.Stderr = string(stderr)
			}
		} else {
			minionResult.Status = define.Fail
			minionResult.Message = "unknown minion result"
		}
		minionResults = append(minionResults, minionResult)
	}
	return minionResults
}

func (s *ScriptModule) getExecResult(execHosts []serviceCommon.ExecHost, jid string, params common.ExecScriptParam,
	fileName string, partitionResult *common.PartitionResult, close chan struct{}) {

	logger := getLogger().Named(jid)

	timer := listenTimeout(params.Timeout, partitionResult)
	defer taskEnding(fileName, timer, jid)

	res := common.Result{
		Status: define.Success,
	}

	hostMinion := extractEntityIDToHostIDMap(execHosts)
	logger.Info("start to get result", "jid", jid)

	// 使用fib数列的等待时间间隔来处理
	sleepTimeDuration := []time.Duration{
		time.Duration(500), time.Duration(500), time.Duration(1000), time.Duration(1500), time.Duration(2500), time.Duration(4000),
	}
	durationIndex := 0

	for {
		select {
		case <-close:
			logger.Info("process is interrupt")
			res.Status = define.Fail
			res.Message = "job interrupted"
			partitionResult.ResultChan <- res
			partitionResult.Close()
			return
		default:
		}

		if partitionResult.IsClosed() {
			logger.Info("return context has closed")
			return
		}

		result, timeout, err := s.getScriptRunResult(jid, sleepTimeDuration[durationIndex]*time.Millisecond)
		logger.Trace("getScriptRunResult success", "result", fmt.Sprintf("%#v", result), "timeout", timeout, "err", err)
		if timeout {
			logger.Warn("get salt job record timeout")
			continue
		}

		if err != nil {
			res.Message = "get job result, err:" + err.Error()
			res.Status = define.Fail
			partitionResult.ResultChan <- res
			partitionResult.Close()
			break
		}

		if result == nil {
			//time.Sleep(sleepTimeDuration[durationIndex] * time.Millisecond)
			durationIndex = (durationIndex + 1) % len(sleepTimeDuration)
			continue
		}

		minionResults := s.buildMinionResult(result, hostMinion, params.Encoding)

		hostNum := len(minionResults)
		minionResults = filterResultAndAdd(jid, minionResults, params.Timeout)
		res.MinionResults = minionResults

		logger.Debug("getExecResult hostNum and execHostsLen", "hostNum", hostNum, "execHostsLen", len(execHosts))
		isClose := false
		if hostNum >= len(execHosts) {
			isClose = true
		}
		partitionResult.ResultChan <- res
		//common.SendResultAndClose(res, partitionResult, isClose)
		if isClose {
			partitionResult.Close()
			break
		}
	}
}

const (
	pythonHeadRegex = `^#!/usr/.+python`
	pythonHead      = "#!/usr/bin/env python"
)

func addPythonHead(script string) (headScript string) {
	logger := getLogger()

	headScript = script

	lines := strings.Split(script, "\n")
	if len(lines) == 0 {
		return
	}

	headLine := lines[0]
	match, err := regexp.MatchString(pythonHeadRegex, headLine)
	if err != nil {
		logger.Error("python head match fail", "regex", pythonHeadRegex, "headLine", headLine, "error", err)
		return
	}

	if !match {
		headScript = fmt.Sprintf("%s\n%s", pythonHead, script)
	}
	return
}

func (s *ScriptModule) buildMinionBody(execHosts []serviceCommon.ExecHost, params common.ExecScriptParam) (string, *saltclient.MinionsPostBody, error) {
	logger := getLogger()

	//为python脚本做特殊处理
	if params.ScriptType == define.PythonType {
		params.Script = addPythonHead(params.Script)
	}

	// 转换为合适的编码进行保存
	script, err := encoding.EncodingTo([]byte(params.Script), params.Encoding)
	if err != nil {
		logger.Error("encoding script", "error", err)
		return "", nil, err
	}

	// 获取脚本文件名
	fileName, err := fileutil.SaveScriptToFileWithSaltSys(config.Conf.Salt.SYSPath, script, params.ScriptType)

	if err != nil {
		logger.Error("save script file", "error", err)
		return "", nil, err
	}

	source := fmt.Sprintf("salt://%s", fileName)

	var args string
	if a, ok := params.Params["args"]; ok {
		args = a.(string)
	} else {
		logger.Error("params not exits args")
		return "", nil, errors.New("params not exits args")
	}

	var password string
	if len(params.Password) > 0 {
		var err error

		client := crypto.GetClient()
		password, err = client.Decode(params.Password)
		if err != nil {
			logger.Error("unable to decode param password", "err", err)
			return "", nil, err
		}
	}

	kwarg := saltclient.ScriptKwargParam{
		Source:          source,
		Args:            args,
		Timeout:         params.Timeout,
		Password:        password,
		Act2ProxyServer: getAct2ProxyURL(),
		UseVT:           params.RealtimeOutput,
		Env:             params.Env,
	}

	// 给salt-api一个空的RunAs的值，会导致异常
	if len(params.RunAs) > 0 {
		kwarg.RunAs = params.RunAs
	}
	hosts := extractExecHostsToHosts(execHosts)
	body := saltclient.MinionsPostBody{
		Fun:     "cmd.script",
		Tgt:     strings.Join(hosts, ","),
		TgtType: "list",
		Kwarg:   kwarg,
	}

	return fileName, &body, nil
}

func getAct2ProxyURL() string {
	return config.Conf.Act2.ProxyServer + "/api/v1/job/realtime"
}
