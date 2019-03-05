//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import "sync"

//ExecScriptParam 通用执行入参
type ExecScriptParam struct {
	// pattern可以为script|file|state
	Pattern        string
	Script         string
	ScriptType     string
	Encoding       string
	Params         map[string]interface{}
	RunAs          string
	Timeout        int
	Password       string
	Env            map[string]string
	ExtendData     interface{}
	RealtimeOutput bool
}

//Minion minion的结构体
type Minion struct {
	ID     string
	Status string
}

//Result 执行结果
type Result struct {
	Jid           string         `json:"-"`
	Status        string         `json:"status"`
	Message       string         `json:"message"`
	MinionResults []MinionResult `json:"minions"`
	TaskID        string         `json:"taskId"`
	CallbackURL   string         `json:"-"`
	RetryCount    uint           `json:"-"`
}

//ReturnContext 用于结果返回的上下文
type ReturnContext struct {
	TaskID      string
	CallbackURL string
	CloseChan   chan int
	once        sync.Once
	mutex       sync.Mutex
}

//Close 关闭通道
func (ctx *ReturnContext) Close() {
	ctx.once.Do(func() {
		close(ctx.CloseChan)
	})
}

//IsClosed 检查通道是否关闭，即任务是否已经结束
func (ctx *ReturnContext) CheckClose() (close bool) {
	select {
	case <-ctx.CloseChan:
		close = true
	default:
	}
	return
}

//MinionResult 主机执行结果
type MinionResult struct {
	HostID  string `json:"hostId"`
	Status  string `json:"status"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Message string `json:"message"`
}
