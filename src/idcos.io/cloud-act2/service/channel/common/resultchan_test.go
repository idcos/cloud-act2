//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDoubleClose(*testing.T) {
	safeChan := NewPartitionResult(2)
	safeChan.Close()
	safeChan.Close()
}

func TestListenClose(*testing.T) {
	safeChan := NewPartitionResult(2)
	wait := sync.WaitGroup{}

	wait.Add(1)

	go func(sc *PartitionResult) {
		for {
			select {
			case <-sc.CloseChan:
				fmt.Println("通道已关闭")
				wait.Done()
				return
			default:
			}

			fmt.Println("------")
			time.Sleep(1 * time.Second)
		}
	}(safeChan)

	go func(sc *PartitionResult) {
		time.Sleep(5 * time.Second)
		sc.Close()
	}(safeChan)

	wait.Wait()
}
