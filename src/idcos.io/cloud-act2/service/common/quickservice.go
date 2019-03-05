//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

//InstallInfo 安装信息
type InstallInfo struct {
	Hosts    []ExecHost `json:"hosts"`
	Username string     `json:"username"`
	Password string     `json:"password"`
	Path     string     `json:"path"`
	MasterIP string     `json:"masterIp"`
}

//InstallInfo 安装的返回数据
type InstallResp struct {
	Error        error             `json:"-"`
	Status       string            `json:"status"`
	SuccessHosts []string          `json:"successHosts"`
	FailHosts    []InstallFailHost `json:"failHosts"`
}

//InstallFailHost 安装失败主机和失败信息
type InstallFailHost struct {
	IP      string `json:"ip"`
	Message string `json:"message"`
}
