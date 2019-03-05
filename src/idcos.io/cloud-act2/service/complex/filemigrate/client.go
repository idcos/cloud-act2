//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package filemigrate

import (
	"fmt"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/crypto"
	"idcos.io/cloud-act2/define"
	common2 "idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/common/service"
	"idcos.io/cloud-act2/service/channel/ssh"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/utils/dataexchange"
	"io"
)

//GetRemoteStream 获取远程主机文件的读取流
func GetRemoteStream(info MigrateInfo) (reader io.Reader, err error) {
	switch info.SourceHost.Provider {
	case define.MasterTypeSSH:
		return getStreamBySSH(info)
	default:
		return nil, errors.New("unknown provider type")
	}
}

func getStreamBySSH(info MigrateInfo) (reader io.Reader, err error) {
	logger := getLogger()

	sshClient := ssh.NewSSHClient()

	client := crypto.GetClient()
	var password string
	if len(info.SourceHost.Password) > 0 {
		var err error
		password, err = client.Decode(info.SourceHost.Password)
		if err != nil {
			return nil, err
		}
	}

	session, err := sshClient.GetSession(info.SourceHost.ExecHost, info.SourceHost.Username, password)
	if err != nil {
		logger.Error("get ssh executor client fail", "error", err)
		return nil, err
	}

	scpClient := ssh.NewScp(session)
	file, err := scpClient.OpenRemoteFile(info.SourceFilePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func PushFile(info MigrateInfo) (result common2.Result, timeout bool, err error) {
	logger := getLogger()

	switch info.TargetHost.Provider {
	case define.MasterTypeSSH:
		return pushFileBySSH(info)
	case define.MasterTypeSalt:
		return pushFileBySalt(info)
	default:
		logger.Error("unsupported protocol type")
		return common2.Result{}, false, errors.New("unsupported protocol type")
	}
}

func pushFileBySalt(info MigrateInfo) (result common2.Result, timeout bool, err error) {
	logger := getLogger()

	// 获取channelClient
	channelClient, err := service.GetClient(info.TargetHost.Provider, define.ScriptModule)
	if err != nil {
		logger.Error("get channel client fail", "provider", info.TargetHost.Provider, "error", err)
		return common2.Result{}, false, err
	}

	urlPath := fmt.Sprintf("%s/api/v1/complex/file/migrate/download", config.Conf.Act2.ProxyServer)
	script := dataexchange.ToJsonString([]string{urlPath})

	execHosts := []common.ExecHost{info.TargetHost.ExecHost}

	resultChan := common2.NewPartitionResult(2)

	encoding := "utf-8"
	if len(info.TargetHost.Encoding) > 0 {
		encoding = info.TargetHost.Encoding
	}
	channelClient.Execute(execHosts, common2.ExecScriptParam{
		Pattern:    define.FileModule,
		Script:     script,
		ScriptType: define.UrlType,
		RunAs:      info.TargetHost.Username,
		Timeout:    info.Timeout,
		Password:   info.TargetHost.Password,
		Encoding:   encoding,
	}, resultChan)

	result, timeout = waitResult(resultChan)
	return result, timeout, nil
}

func pushFileBySSH(info MigrateInfo) (result common2.Result, timeout bool, err error) {
	logger := getLogger()

	sshClient := ssh.NewSSHClient()
	var password string
	if len(info.TargetHost.Password) > 0 {
		var err error
		client := crypto.GetClient()
		password, err = client.Decode(info.TargetHost.Password)
		if err != nil {
			return common2.Result{}, false, err
		}
	}

	session, err := sshClient.GetSession(info.TargetHost.ExecHost, info.TargetHost.Username, password)
	if err != nil {
		logger.Error("get ssh executor client fail", "error", err)
		return common2.Result{}, false, err
	}

	scpClient := ssh.NewScp(session)
	if info.MasterTransfer {
		err = streamCopy(scpClient, info)
		if err != nil {
			return common2.Result{}, false, err
		}

		return common2.Result{
			Status: define.Success,
			MinionResults: []common2.MinionResult{
				{
					HostID: info.TargetHost.HostID,
					Status: define.Success,
				},
			},
		}, false, nil
	}
	reader, err := scpClient.GetWriter(info.TargetFilePath)
	if err != nil {
		return common2.Result{}, false, err
	}
	err = GetFileByMaster(info, reader)
	if err != nil {
		return common2.Result{}, false, err
	}
	return common2.Result{
		Status: define.Success,
		MinionResults: []common2.MinionResult{
			{
				HostID: info.TargetHost.HostID,
				Status: define.Success,
			},
		},
	}, false, nil
}

func streamCopy(scpClient *ssh.Scp, info MigrateInfo) error {
	reader, err := GetRemoteStream(info)
	if err != nil {
		return err
	}
	err = scpClient.PushFileReader(reader, info.TargetFilePath)
	if err != nil {
		return err
	}
	return nil
}

func waitResult(resultChan *common2.PartitionResult) (result common2.Result, timeout bool) {
	resultList := make([]common2.Result, 0, 1)
	for {
		select {
		case results := <-resultChan.ResultChan:
			resultList = append(resultList, results)
		case <-resultChan.CloseChan:
			if len(resultChan.ResultChan) == 0 {
				goto loop
			}
		}
	}

loop:
	if len(resultList) == 0 {
		return common2.Result{}, true
	} else {
		return resultList[0], false
	}
}
