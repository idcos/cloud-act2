//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"net/url"
	"path"
	"strings"
)

//ParseFileUrl 解析文件的url
func ParseFileUrl(script string) (source string, err error) {
	logger := getLogger()

	var urls []string
	err = json.Unmarshal([]byte(script), &urls)
	if err != nil {
		logger.Error("not valid json", "error", err)
		return
	}
	logger.Info("file", "urls", urls)
	if len(urls) < 1 {
		logger.Error("urls should have elements")
		err = errors.New("urls should have elements")
		return
	}
	if config.Conf.Act2.FileReversed {
		source = newProxyUrlAddr(urls[0])
	} else {
		source = urls[0]
	}

	return
}

func newProxyUrlAddr(fileUrl string) string {
	values := url.Values{}
	values.Add("redirect", fileUrl)

	proxyServer := config.Conf.Act2.ProxyServer
	proxyServer = strings.TrimRight(proxyServer, "/") + define.FileReversedUri
	return fmt.Sprintf("%s?%s", proxyServer, values.Encode())
}

//GetFileTarget 获取文件目标位置
func GetFileTarget(params map[string]interface{}) (target string, err error) {
	logger := getLogger()

	target, ok := params["target"].(string)
	if !ok {
		logger.Error("params not exists target args")
		return "", errors.New("params should have target args")
	}
	remoteFileName, ok := params["fileName"].(string)
	if !ok {
		logger.Error("params not exists fileName args")
		return "", errors.New("params should have fileName args")
	}
	target = path.Join(target, remoteFileName)
	return
}
