//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

// ProxyJobExecParam 接收执行参数
type ProxyJobExecParam struct {
	ExecHosts []ExecHost `json:"execHosts" validate:"required"`
	ExecParam ExecParam  `json:"execParam" validate:"required"`
	Callback  string     `json:"callback" validate:"required"`
	TaskID    string     `json:"taskId" validate:"required"`
	Provider  string     `json:"provider" validate:"required"`
}

type SyncJobExecParam struct {
	ExecHosts []ExecHost `json:"execHosts" validate:"required"`
	ExecParam ExecParam  `json:"execParam" validate:"required"`
	Provider  string     `json:"provider" validate:"required"`
}

type ConfJobIDExecParam struct {
	EntityIDs []string  `json:"entityIds" validate:"required"`
	ExecParam ExecParam `json:"execParam" validate:"required"`
	Provider  string    `json:"provider" validate:"required"`
	Callback  string    `json:"callback"`
	ExecuteID string    `json:"executeId" validate:"required"`
}

// ConfJobIPExecParam 作业执行参数
type ConfJobIPExecParam struct {
	ExecHosts   []ExecHost `json:"execHosts" validate:"required"`
	ExecParam   ExecParam  `json:"execParam" validate:"required"`
	Provider    string     `json:"provider" validate:"required"` // provider可以为salt|puppet|openssh
	Callback    string     `json:"callback"`
	ExecuteID   string     `json:"executeId" validate:"required"`
}

/**
执行主机信息
*/
type ExecHost struct {
	HostIP   string `json:"hostIp"`
	HostPort int    `json:"hostPort"`
	EntityID string `json:"entityId"`
	HostID   string `json:"hostId"`
	IdcName  string `json:"idcName,omitempty"`
	OsType   string `json:"osType,omitempty"`
	Encoding string `json:"encoding,omitempty"` // 系统默认的编码，如果为空，则默认以utf-8值进行处理
	ProxyID  string `json:"proxyId"`
}

// IsWindows check
func (e ExecHost) IsWindows() bool {
	return e.OsType == "windows"
}

type JobCallbackParam struct {
	JobRecordID   string               `json:"jobRecordId"`
	ExecuteID     string               `json:"executeId"`
	ExecuteStatus string               `json:"executeStatus"`
	ResultStatus  string               `json:"resultStatus"`
	HostResults   []HostResultCallback `json:"hostResults"`
}

type HostResultCallback struct {
	EntityID string `json:"entityId"`
	HostIP   string `json:"hostIp"`
	IdcName  string `json:"idcName"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Time     string `json:"time"`
}

//ExecParam 执行参数
type ExecParam struct {
	// 模块名称，支持 script：脚本执行, salt.state：状态应用, file：文件下发
	// new in 2018年08月23日
	Pattern string `json:"pattern" validate:"required"`

	// 依据模块名称进行解释
	// pattern为script时，script为脚本内容
	// pattern为salt.state时，script为salt的state内容
	// pattern为file时，script为文件内容或url数组列表
	Script string `json:"script"`
	// 依据pattern进行解释
	// pattern为script时，scriptType为shell, bat, python
	// pattern为file时，scriptType为url或者text

	ScriptType string                 `json:"scriptType" validate:"required"`
	Params     map[string]interface{} `json:"params"`
	RunAs      string                 `json:"runas,omitempty"`
	Password   string                 `json:"password"`
	Timeout    int                    `json:"timeout" validate:"required"`
	Env        map[string]string      `json:"env"`
	ExtendData interface{}            `json:"extendData"`
	// 是否实时输出，像巡检任务、定时任务则不需要实时输出
	RealTimeOutput bool `json:"realTimeOutput"`
}

type ProxyResultContent struct {
	TaskID      string                    `json:"taskId" validate:"required"`
	ProxyID     string                    `json:"proxyId"`
	MasterSend  bool                      `json:"masterSend"`
	HostResults []ProxyCallbackHostResult `json:"hostResults" validate:"required"`
	StartTime   string                    `json:"startTime"`
	EndTime     string                    `json:"endTime"`
}

//ProxyCallBackResult proxy执行结果
type ProxyCallBackResult struct {
	Status  string             `json:"status" validate:"required"`
	Message string             `json:"message"`
	Content ProxyResultContent `json:"content" validate:"required"`
	ErrChan chan error         `json:"-"`
}

//ProxyCallbackHostResult proxy的主机执行结果
type ProxyCallbackHostResult struct {
	HostID string `json:"hostId" validate:"required"`
	Status string `json:"status" validate:"required"`
	Stdout string `json:"stdout" validate:"required"`
	Stderr string `json:"stderr" validate:"required"`
	// 机器上的错误信息，如这一台机器的超时信息等
	Message string `json:"message" validate:"required"`
}

//JobCancelForm 作业取消
type JobCancelForm struct {
	JobRecordID string `json:"jobRecordId"`
}

//RealTimeForm 实时结果
type RealTimeForm struct {
	Type     string  `json:"type"`
	Jid      string  `json:"jid"`
	EntityID string  `json:"entityId"`
	Now      float64 `json:"now"`
	Stdout   string  `json:"stdout"`
	Stderr   string  `json:"stderr"`
}

//RealTimeRedisForm redis中实时结果
type RealTimeRedisForm struct {
	Now    float64 `json:"now"`
	Stdout string  `json:"stdout"`
	Stderr string  `json:"stderr"`
}
