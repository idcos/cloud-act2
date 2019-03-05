//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"sync"
)

//SafeChannel 安全通道
type SafeChannel struct{
	ResultChan chan interface{}
	CloseChan chan int
	once sync.Once
}

//NewSafeChan 新建执行结果安全通道
func NewSafeChan(chanLen int)(safeChan *SafeChannel){
	safeChan = &SafeChannel{
		ResultChan: make(chan interface{},chanLen),
		CloseChan: make(chan int),
	}
	return
}

//Close 关闭通道
func (safeChan *SafeChannel) Close(){
	safeChan.once.Do(func(){
		close(safeChan.CloseChan)
	})
}
