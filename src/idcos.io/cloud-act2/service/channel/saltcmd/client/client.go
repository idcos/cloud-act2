//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
//Package client saltstack基于本地命令行的客户端
package client

import (
	"errors"

	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/utils"
)

//SaltCMDClient saltcmd
type SaltCMDClient struct {
}

//NewSaltCMDClient 获取一个saltCMDClient
func NewSaltCMDClient() (client *SaltCMDClient) {
	return &SaltCMDClient{}
}

//Execute 执行脚本
func (client *SaltCMDClient) ExecutionScript(hosts []string, params common.ExecScriptParam,
	callback func(stdout, stderr []byte, err error)) (err error) {
	if len(hosts) == 0 {
		err = errors.New("主机列表不能为空！")
		return
	}

	_, err = utils.SaveScriptToFileWithSaltSys(config.Conf.Salt.SYSPath, []byte(params.Script), params.ScriptType)
	if err != nil {
		return
	}

	return
}

type commandParam struct {
	Source   string
	Args     string
	Runas    string
	Timeout  int
	Password string
	Env      map[string]string
}
