//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"idcos.io/cloud-act2/service/common"
)

//ChannelClient channel客户端的接口
type ChannelClient interface {
	Execute(hosts []common.ExecHost, params ExecScriptParam, encodingResultChan *PartitionResult) (jid string, err error)
}
