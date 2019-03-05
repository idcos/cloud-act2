//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package system

import (
	"io/ioutil"
	"runtime"
	"strings"

	"idcos.io/cloud-act2/log"
)

//GetSystemSN mac will not running
//mac系统，运行下面的内容到/tmp/product_uuid中
// system_profiler SPHardwareDataType | awk '/UUID/ { print tolower($3); }'
func GetSystemSN() (string, error) {
	logger := log.L()
	var filename string
	if runtime.GOOS == "linux" {
		filename = "/sys/class/dmi/id/product_uuid"
	} else {
		filename = "/tmp/product_uuid"
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Info("in mac system, you should use `system_profiler SPHardwareDataType | awk '/UUID/ { print tolower($3); }' > /tmp/product_uuid` before run system")
		logger.Error("read product uuid", "error", err)
		return "", nil
	}
	return strings.ToUpper(strings.TrimSpace(string(bytes))), nil
}
