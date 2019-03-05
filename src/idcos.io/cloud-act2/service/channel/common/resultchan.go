//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"sync"
)

// PartitionResult 执行结果安全通道
type PartitionResult struct {
	ResultChan chan Result
	CloseChan  chan int
	once       sync.Once
	closed     bool
}

// NewPartitionResult 新建执行结果安全通道
func NewPartitionResult(chanLen int) (pr *PartitionResult) {
	pr = &PartitionResult{
		ResultChan: make(chan Result, chanLen),
		CloseChan:  make(chan int),
	}
	return
}

//Close 关闭通道
func (pr *PartitionResult) Close() {
	pr.once.Do(func() {
		close(pr.CloseChan)
		pr.closed = true
	})
}

func (pr *PartitionResult) IsClosed() bool {
	return pr.closed
}
