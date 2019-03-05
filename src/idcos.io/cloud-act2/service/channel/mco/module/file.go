//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import (
	"idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/channel/mco"
	serviceCommon "idcos.io/cloud-act2/service/common"
)

//McoFileModuleClient mco的文件下发客户端
type McoFileModuleClient struct {
	mco *mco.McoClient
}

//NewMcoFileModuleClient 新建mco的文件下发客户端
func NewMcoFileModuleClient(mco *mco.McoClient) (client *McoFileModuleClient) {
	return &McoFileModuleClient{
		mco: mco,
	}
}

//Execute 执行
func (client *McoFileModuleClient) Execute(execHosts []serviceCommon.ExecHost, executeParam common.ExecScriptParam, returnContext *common.PartitionResult) (string, error) {
	source, err := common.ParseFileUrl(executeParam.Script)
	if err != nil {
		return "", nil
	}

	target, err := common.GetFileTarget(executeParam.Params)
	if err != nil {
		return "", nil
	}

	script := `puppet resource file "nodeserver"  path="` + target + `" source="` + source + `"  mode=755 owner=root`
	return client.mco.Execute(execHosts, script, returnContext)
}
