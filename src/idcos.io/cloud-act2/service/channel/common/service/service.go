//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
//Package service channel对外的统一服务
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"idcos.io/cloud-act2/utils/dataexchange"
	"sync"
	"time"

	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/service/common/report"
	"idcos.io/cloud-act2/utils/debug"
	"idcos.io/cloud-act2/utils/promise"

	"github.com/jinzhu/copier"
	globalConfig "idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/mco"
	mcoModule "idcos.io/cloud-act2/service/channel/mco/module"
	"idcos.io/cloud-act2/service/channel/saltclient"
	"idcos.io/cloud-act2/service/channel/saltclient/module"
	"idcos.io/cloud-act2/service/channel/ssh"
	serviceCommon "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/job"
)

var clientMap = make(map[string]common.ChannelClient)

func GetClientByTypeFromCache(typ, moduleName string) (client common.ChannelClient, found bool) {
	logger := getLogger()

	key := fmt.Sprintf("%s_%s", typ, moduleName)
	client, ok := clientMap[key]
	if !ok {
		found = false
		return
	}

	found = true
	logger.Debug("get channel client by cache")
	return
}

func CreateClientAndSaveCache(config ChannelConfig, moduleName string) (client common.ChannelClient, err error) {
	logger := getLogger()

	client, err = produceClientByConfig(config, moduleName)
	if err != nil {
		logger.Error("fail to producerClientByConfig", "error", err)
		return
	}

	key := fmt.Sprintf("%s_%s", config.Type, moduleName)
	clientMap[key] = client
	logger.Debug("create channel client success", "moduleName", moduleName)
	return
}

func produceClientByConfig(config ChannelConfig, moduleName string) (client common.ChannelClient, err error) {
	switch config.Type {
	case define.MasterTypeSalt:
		client, err = produceSaltModule(config, moduleName)
		return
	case define.MasterTypeSSH:
		sshChannel, err := produceSSHClient()
		if err != nil {
			return nil, err
		}
		return sshChannel, nil
	case define.MasterTypePuppet:
		mcoClient, err := produceMcoModule(config, moduleName)
		if err != nil {
			return nil, err
		}
		return mcoClient, nil
	default:
		err = errors.New("未知的channel类型,type:" + config.Type)
	}
	return
}

func produceSaltModule(config ChannelConfig, moduleName string) (client common.ChannelClient, err error) {
	salt, err := produceSaltClient(config)
	if err != nil {
		return nil, err
	}

	switch moduleName {
	case define.ScriptModule:
		client = module.NewScriptModule(salt)
	case define.FileModule:
		client = module.NewFileModule(salt)
	case define.StateModule:
		client = module.NewStateModule(salt)
	default:
		err = fmt.Errorf("unknown module name: %s", moduleName)
	}
	return
}

func produceMcoModule(config ChannelConfig, moduleName string) (client common.ChannelClient, err error) {
	mcoClient, err := produceMcoClient()
	if err != nil {
		return nil, err
	}

	switch moduleName {
	case define.ScriptModule:
		client = mcoModule.NewMcoScriptModuleClient(mcoClient)
	case define.FileModule:
		client = mcoModule.NewMcoFileModuleClient(mcoClient)
	default:
		err = fmt.Errorf("unknown module name: %s", moduleName)
	}
	return
}

func produceSaltClient(config ChannelConfig) (client *saltclient.SaltClient, err error) {
	//解析option获取username和password
	if len(config.Option) == 0 {
		err = errors.New("option项为空！saltclient无法获取用户名和密码")
		return
	}
	option := saltOption{}
	err = json.Unmarshal([]byte(config.Option), &option)
	if err != nil {
		err = errors.New("option格式有误，解析失败，option:" + config.Option + ",err:" + err.Error())
		return
	}

	//创建saltClient
	saltConfig := saltclient.Config{
		Server:   config.Server,
		Username: option.UserName,
		Password: option.Password,
	}
	client, err = saltclient.NewSaltClient(saltConfig)
	return
}

func produceSSHClient() (*ssh.SSHChannel, error) {
	sshChannel := ssh.NewSSHChannel()
	return sshChannel, nil
}

func produceMcoClient() (*mco.McoClient, error) {
	return mco.NewMcoClient(), nil
}

func groupByEncodingHosts(execHosts []serviceCommon.ExecHost) map[string][]serviceCommon.ExecHost {
	encodingHosts := map[string][]serviceCommon.ExecHost{}
	for _, execHost := range execHosts {
		encoding := execHost.Encoding
		if hosts, ok := encodingHosts[encoding]; ok {
			hosts = append(hosts, execHost)
			encodingHosts[encoding] = hosts
		} else {
			var hosts []serviceCommon.ExecHost
			hosts = append(hosts, execHost)
			encodingHosts[encoding] = hosts
		}
	}
	return encodingHosts
}

func waitForResult(partitionResult *common.PartitionResult, taskID string, jid string, taskResultChan chan common.Result) {
	logger := getLogger()

	defer func() {
		logger.Debug("job exec done", "taskID", taskID, "jid", jid)
	}()

	for {
		select {
		// closeChan被close之后，此处一定会收到消息，所以下面的代码会在close之后被多次执行
		case <-partitionResult.CloseChan:
			logger.Info("")
			if len(partitionResult.ResultChan) == 0 {
				return
			}
		default:
			time.Sleep(time.Duration(10) * time.Microsecond)
		}

		// 需要分开处理，ResultChan和CloseChan可能会同时达到
		// 此时如果CloseChan数据被先处理，则执行的时候，就会丢失数据
		select {
		case results := <-partitionResult.ResultChan:
			logger.Debug("callback job exec result", "jid", jid, "results", results)
			// processResult(callbackURL, jobRecordID, results)

			taskResultChan <- results
		case <-partitionResult.CloseChan:
			if len(partitionResult.ResultChan) == 0 {
				return
			}
		}
	}
}

func buildFailMinionResult(execHosts []serviceCommon.ExecHost, message string) []common.MinionResult {
	var minionResults []common.MinionResult
	for _, execHost := range execHosts {
		minionResult := common.MinionResult{
			HostID:  execHost.HostID,
			Status:  define.Fail,
			Message: message,
		}
		minionResults = append(minionResults, minionResult)
	}
	return minionResults
}

func getExecuteResult(taskResultChan chan common.Result, returnContext *common.ReturnContext) {
	defer debug.Recover()

	logger := getLogger()

	status := define.Success
	var message string
	var minionResults []common.MinionResult
	for result := range taskResultChan {
		// 获取到所有结果后，发送给上层调用方
		if result.Status != define.Success {
			status = define.Fail
			message = result.Message
		}

		logger.Debug("receive task message", "result", fmt.Sprintf("%v", result))

		minionResults = append(minionResults, result.MinionResults...)
	}

	// TODO: need trace id for log information
	logger.Info("execute host complete, combine the result")

	r := common.Result{
		Status:        status,
		Message:       message,
		MinionResults: minionResults,
	}
	common.SendResultAndClose(r, returnContext, true)
}

func getConfigByProvider(provider string) ChannelConfig {
	var server, option string
	if provider == define.MasterTypeSalt {
		server = globalConfig.Conf.Salt.Server
		option = fmt.Sprintf(`{"username":"%s","password":"%s"}`, globalConfig.Conf.Salt.Username, globalConfig.Conf.Salt.Password)
	} else if provider == define.MasterTypePuppet {
		server = globalConfig.Conf.Puppet.RabbitMQ
		option = globalConfig.Conf.Puppet.ReplyQueue
	}
	return ChannelConfig{
		Server: server,
		Type:   provider,
		Option: option,
	}
}

//GetClient 获取通道client
func GetClient(provider string, module string) (common.ChannelClient, error) {
	// 获取或创建channelClient
	channelClient, found := GetClientByTypeFromCache(provider, module)
	if !found {
		var err error
		channelClient, err = CreateClientAndSaveCache(getConfigByProvider(provider), module)
		if err != nil {
			return nil, err
		}
	}
	return channelClient, nil
}

func encodingProcess(taskResultChan chan common.Result, provider string, executeParam common.ExecScriptParam,
	execHosts []serviceCommon.ExecHost, taskID string, returnContext *common.ReturnContext) {

	defer debug.Recover()

	logger := getLogger()

	defer close(taskResultChan)

	var channelClient common.ChannelClient
	var err error

	// 获取或创建channelClient
	channelClient, err = GetClient(provider, executeParam.Pattern)
	if err != nil {
		logger.Error("create channel client fail", "error", err)
		result := common.Result{
			Status:  define.Fail,
			Message: err.Error(),
		}
		taskResultChan <- result
		return
	}

	// 对于同一类型的编码进行处理
	encodingHosts := groupByEncodingHosts(execHosts)

	logger.Debug("exec encoding host", "hosts", fmt.Sprintf("%#v", encodingHosts))

	wg := sync.WaitGroup{}
	wg.Add(len(encodingHosts))
	for encoding, encodingExecHosts := range encodingHosts {
		promise.NewGoPromise(func(chan struct{}) {
			defer debug.Recover()

			defer func() {
				wg.Done()
				logger.Info("exec running done", "encoding", encoding)
			}()

			var callExecParam common.ExecScriptParam
			copier.Copy(&callExecParam, &executeParam)
			callExecParam.Encoding = encoding

			partitionResult := common.NewPartitionResult(2)

			//超时时间修正
			timeout := executeParam.Timeout - globalConfig.Conf.Act2.TimeoutCorrection
			//如果修改后时间小于等于0，则超时时间为原超时时间的一半
			if timeout <= 0 {
				timeout = executeParam.Timeout / 2
			}
			executeParam.Timeout = timeout

			logger.Debug("exec param", "timeout", fmt.Sprintf("%v", timeout))

			if logger.IsTrace() {
				logger.Trace("start channel execute", "encoding", encoding, "encodingExecHosts",
					dataexchange.ToJsonString(encodingExecHosts), "execParam", dataexchange.ToJsonString(callExecParam))
			}

			// 执行
			jid, err := channelClient.Execute(encodingExecHosts, callExecParam, partitionResult)
			if err != nil {
				logger.Error("channel client execute script fail", "error", err)
				result := common.Result{
					Jid:           jid,
					Status:        define.Fail,
					Message:       err.Error(),
					MinionResults: buildFailMinionResult(encodingExecHosts, err.Error()),
				}
				taskResultChan <- result
				return
			}

			logger.Info("salt job already start", "encoding", encoding, "taskID", taskID, "jid", jid)

			//为实时输出记录jid对应的作业信息
			if executeParam.RealtimeOutput {
				job.AddJobInfo(jid, job.Info{TaskID: taskID})
			}
			// 获取结果
			waitForResult(partitionResult, taskID, jid, taskResultChan)

		}, nil)
	}
	wg.Wait()
}

//Execute channel层对外的统一执行接口
func Execute(provider string, taskID string, execHosts []serviceCommon.ExecHost, executeParam common.ExecScriptParam, returnContext *common.ReturnContext) {
	taskResultChan := make(chan common.Result, len(execHosts))

	promise.NewGoPromise(func(chan struct{}) {
		getExecuteResult(taskResultChan, returnContext)
	}, nil)

	promise.NewGoPromise(func(chan struct{}) {
		encodingProcess(taskResultChan, provider, executeParam, execHosts, taskID, returnContext)
	}, nil)

	//统计doing
	recorder := report.GetRecorder(globalConfig.ComData.SN)
	err := recorder.AddDoing()
	if err != nil {
		getLogger().Warn("add record doing", "error", err)
	}
}
