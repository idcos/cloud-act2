//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package config

import (
	"io/ioutil"
	"runtime"
	"strings"
)

//CommonData 公共数据
type CommonData struct {
	SN string
}

var ComData = &CommonData{}

//LoadCommonData 加载
func LoadCommonData() (err error) {
	filename := ""
	if runtime.GOOS == "linux" {
		filename = "/sys/class/dmi/id/product_uuid"
	} else {
		filename = "/tmp/product_uuid"
	}
	uuid, err := readFile(filename)
	if err != nil {
		return
	}

	ComData.SN = strings.TrimSpace(uuid)
	return
}

func readFile(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)

	if err != nil {
		return "", err
	}
	return string(bytes), err
}
