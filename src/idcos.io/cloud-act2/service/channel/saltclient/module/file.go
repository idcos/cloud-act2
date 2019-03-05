//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import (
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/fileutil"
	"idcos.io/cloud-act2/utils/promise"
	"strings"
	"time"

	"github.com/RussellLuo/timingwheel"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/saltclient"
	serviceCommon "idcos.io/cloud-act2/service/common"
)

type FileModule struct {
	salt *saltclient.SaltClient
}

func NewFileModule(salt *saltclient.SaltClient) *FileModule {
	return &FileModule{
		salt: salt,
	}
}

func (s *FileModule) Execute(execHosts []serviceCommon.ExecHost, params common.ExecScriptParam,
	partitionResult *common.PartitionResult) (jid string, err error) {

	logger := getLogger()
	if len(execHosts) == 0 {
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

	// 构建参数列表
	fileName, body, err := s.buildMinionBody(execHosts, params)
	if err != nil {
		logger.Error("execution script", "error", err)
		return "", err
	}

	logger.Debug("request info", "filename", fileName, "minionBody", dataexchange.ToJsonString(body))

	byts, err := s.salt.MinionExecute(body)
	if err != nil {
		logger.Error("call minion execute", "error", err)
		return "", err
	}

	logger.Debug("minion", "response", string(byts))

	result := saltclient.RunResp{}
	err = json.Unmarshal(byts, &result)
	if err != nil {
		logger.Error("/minions file response to json fail", "error", err)
		return "", err
	}

	logger.Debug("file response value", "result", result)

	if len(result.Return) == 0 {
		logger.Error("run minion error")
		return "", errors.New("run minion error, may parameters error")
	}

	jid = result.Return[0].Jid

	// 如果jid为空，则需要特殊处理
	if jid == "" {
		logger.Error("run minion found jid is empty")
		return "", errors.New("not found jid")
	}

	promise.NewGoPromise(func(close chan struct{}) {
		s.getExecResult(execHosts, jid, params.Timeout, fileName, partitionResult, close)
	}, func(message interface{}) {
		logger.Error("promise panic error", "error", message)
	})

	return jid, nil
}

func getFileResponse(resp []byte) (*saltclient.FileResultResp, error) {
	result := saltclient.FileResultResp{}
	err := yaml.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func taskEnding(fileName string, timer *timingwheel.Timer, jid string) {
	logger := getLogger()

	err := fileutil.RemoveFile(config.Conf.Salt.SYSPath, fileName)
	if err != nil {
		logger.Error("delete temp script file fail", "fileName", fileName, "syspath", config.Conf.Salt.SYSPath)
	}

	timer.Stop()
	removeResultRecord(jid)
}

func (s *FileModule) getExecResult(execHosts []serviceCommon.ExecHost, jid string, timeout int, fileName string, partitionResult *common.PartitionResult, close chan struct{}) {
	logger := getLogger()

	timer := listenTimeout(timeout, partitionResult)

	defer taskEnding(fileName, timer, jid)

	res := common.Result{
		Status: define.Success,
	}

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
			return
		}

		byts, err := s.salt.GetJobResult(jid, sleepTimeDuration[durationIndex]*time.Millisecond)
		durationIndex = (durationIndex + 1) % len(sleepTimeDuration)
		if err != nil {
			if strings.Index(err.Error(), "timeout") != -1 {
				logger.Error("request /jobs fail", "error", err)
			}
			continue
		}
		logger.Debug("get job result", "response body", string(byts))

		hostNum := processFileResult(byts, &res, execHosts, jid, timeout)

		isClose := false
		if hostNum >= len(execHosts) {
			isClose = true
		}
		partitionResult.ResultChan <- res
		if isClose {
			partitionResult.Close()
			break
		}
	}
}

func processFileResult(byts []byte, res *common.Result, execHosts []serviceCommon.ExecHost, jid string, timeout int) (hostNum int) {
	logger := getLogger()

	result, err := getFileResponse(byts)
	if err != nil {
		logger.Error("/job result to yaml fail", "error", err)
		res.Message = "/job result to yaml fail,err:" + err.Error()
		res.Status = define.Fail
		return
	}

	hostMinion := extractEntityIDToHostIDMap(execHosts)

	var minionResults []common.MinionResult
	for _, minionInfo := range result.Return {
		for minionID := range minionInfo {
			minionResult := common.MinionResult{}
			minionResult.HostID = hostMinion[minionID]
			minionResult.Status = define.Success

			minionResults = append(minionResults, minionResult)
		}

	}
	hostNum = len(minionResults)
	minionResults = filterResultAndAdd(jid, minionResults, timeout)
	res.MinionResults = minionResults
	return
}

func (s *FileModule) buildMinionBody(execHosts []serviceCommon.ExecHost, params common.ExecScriptParam) (string, *saltclient.MinionsPostBody, error) {
	logger := getLogger()

	var source, fileName string
	var saltFunc string

	switch params.ScriptType {
	case "url":
		tempSource, tempSaltFuncm, err := parseFileParamByURL(params)
		if err != nil {
			return "", nil, err
		}

		source = tempSource
		saltFunc = tempSaltFuncm
		break
	case "conf":
		fallthrough
	case "text":
		tempFileName, tempSource, tempSaltFunc, err := parseFileParamByText(params)
		if err != nil {
			return "", nil, err
		}

		fileName = tempFileName
		source = tempSource
		saltFunc = tempSaltFunc
		break
	default:
		logger.Error("unsupported script type of file distribution")
		return "", nil, errors.New("unsupported script type of file distribution")
	}

	logger.Info("source info", "source", source)

	target, err := common.GetFileTarget(params.Params)
	if err != nil {
		return "", nil, err
	}

	kwarg := saltclient.FileKwargParam{
		Dest:     target,
		Source:   source,
		MakeDirs: true,
	}
	hosts := extractExecHostsToHosts(execHosts)

	body := saltclient.MinionsPostBody{
		Fun:     saltFunc,
		Tgt:     strings.Join(hosts, ","),
		TgtType: "list",
		Kwarg:   kwarg,
	}

	return fileName, &body, nil
}

func parseFileParamByURL(params common.ExecScriptParam) (source, saltFunc string, err error) {
	source, err = common.ParseFileUrl(params.Script)
	if err != nil {
		return
	}

	saltFunc = "cp.get_url"
	return
}

func parseFileParamByText(params common.ExecScriptParam) (fileName, source, saltFunc string, err error) {
	logger := getLogger()
	script, err := fileutil.GetDataURI([]byte(params.Script))
	if err != nil {
		logger.Error("get script file content", "error", err)
		return
	}

	fileName, err = fileutil.SaveScriptToFileWithSaltSys(config.Conf.Salt.SYSPath, script, "conf")
	if err != nil {
		logger.Error("save script file", "error", err)
		return
	}

	source = fmt.Sprintf("salt://%s", fileName)
	saltFunc = "cp.get_file"
	return
}
