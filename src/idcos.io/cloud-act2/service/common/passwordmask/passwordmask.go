//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package passwordmask

import (
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"strings"
)

//GetPasswordMask 获取密码掩盖字符串
func GetPasswordMask(password string) string {
	logLevel := strings.ToUpper(config.Conf.Logger.LogLevel)
	if logLevel == define.Debug || logLevel == define.Trace {
		return password
	}

	return "******"
}
