//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import "idcos.io/cloud-act2/service/common"

type ComplexHost struct {
	common.ExecHost
	Provider string `json:"provider"`
	Username string `json:"username"`
	Password string `json:"password"`
}
