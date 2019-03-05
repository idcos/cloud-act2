//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package aux

import (
	"errors"
	"fmt"
	"idcos.io/cloud-act2/utils/fileutil"

	"github.com/howeyc/gopass"
)

func IsValidProvider(provider string) bool {
	providers := []string{"salt", "puppet", "ssh"}
	for _, p := range providers {
		if p == provider {
			return true
		}
	}
	return false
}

func ExtractCommand(command string, srcFile string) (string, error) {
	// command 和 srcFile不能共存
	if command == "" && srcFile == "" {
		return "", errors.New("command and src file should not both supply")
	}

	if command == "" {
		script, err := fileutil.ReadFile(srcFile)
		if err != nil {
			return "", err
		}
		command = script
	}

	return command, nil
}

func GetRealOsType(osType string) string {
	if osType == "win" {
		return "windows"
	} else {
		return osType
	}
}

// salt 2018.3.3之后，windows的默认返回值也是utf-8了
func GetEncoding(osType string) string {
	encoding := "utf-8"

	//if !(act2ctl.SaltVersion == "" || act2ctl.SaltVersion == "2018.3.3") {
	//	if osType == define.Win {
	//		encoding = "gb18030"
	//	}
	//}

	return encoding
}

func ReadPassword(msg string) (string, error) {
	fmt.Printf("%s: ", msg)
	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		return "", err
	}

	return string(pass), nil
}
