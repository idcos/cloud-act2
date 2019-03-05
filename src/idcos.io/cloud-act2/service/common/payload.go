//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"time"
)

//ParamPayload 作业执行
type ParamPayload struct {
	ExecHosts []HostPayload `json:"execHosts"`
	Param     ExecParam     `json:"param"`
	Provider  string        `json:"provider"`
}

//HostPayload 主机数据
type HostPayload struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	EntityID string `json:"entityId"`
	IdcName  string `json:"idcName"`
	OsType   string `json:"osType,omitempty"`
	Encoding string `json:"encoding,omitempty"` // 系统默认的编码，如果为空，则默认以utf-8值进行处理
}

//JobRunPayload 作业执行
type JobRunPayload struct {
	Info      ParamPayload `json:"info"`
	StartTime time.Time    `json:"startTime"`
	RecordID  string       `json:"recordId"`
}

//JobCancelPayload 作业取消
type JobCancelPayload struct {
	RecordID string `json:"recordId"`
}

//JobDonePayload 作业完成
type JobDonePayload struct {
	RecordID string `json:"recordId"`
	Status   string `json:"status"`
}

//JobDoneStatusPayload 作业成功、失败、超时
type JobDoneStatusPayload struct {
	RecordID string `json:"recordId"`
}

//HostChangeProxyPayload 主机的proxy变更
type HostChangeProxyPayload struct {
	EntityID    string `json:"entityId"`
	ProxyType   string `json:"proxyType"`
	ProxyServer string `json:"proxyServer"`
}

//ProxyPayload proxy
type ProxyPayload struct {
	ProxyID     string `json:"proxyId"`
	ProxyServer string `json:"proxyServer"`
	ProxyType   string `json:"proxyType"`
}
