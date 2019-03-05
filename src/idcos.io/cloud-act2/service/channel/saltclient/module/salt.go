//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package module

import "idcos.io/cloud-act2/service/channel/common"

type SaltModuler interface {
	Execute(hosts []string, params common.ExecScriptParam, resultChan chan common.Result) (string, error)
}
