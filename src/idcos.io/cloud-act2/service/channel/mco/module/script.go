//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import (
	"idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/mco"
	serviceCommon "idcos.io/cloud-act2/service/common"
)

//McoScriptModuleClient mco的脚本执行客户端
type McoScriptModuleClient struct {
	mco *mco.McoClient
}

//NewMcoScriptModuleClient 新建mco的脚本执行客户端
func NewMcoScriptModuleClient(mco *mco.McoClient) (client *McoScriptModuleClient) {
	return &McoScriptModuleClient{
		mco: mco,
	}
}

//Execute 执行
func (client *McoScriptModuleClient) Execute(execHosts []serviceCommon.ExecHost, executeParam common.ExecScriptParam, returnContext *common.PartitionResult) (string, error) {
	return client.mco.Execute(execHosts, executeParam.Script, returnContext)
}
