//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package promise

import (
	"sync"
)

//GoPromise goroutine优化
type GoPromise struct {
	Process func(close chan struct{})
	Catch   func(message interface{})
}

type GoAction struct {
	closeChan chan struct{}
	once      sync.Once
	promise   *GoPromise
}

//NewGoPromise 新建并发业务
func NewGoPromise(process func(close chan struct{}), catch func(message interface{})) *GoAction {
	promise := GoPromise{
		Process: process,
		Catch:   catch,
	}

	action := &GoAction{
		closeChan: make(chan struct{}),
		promise:   &promise,
	}

	run(action)

	return action
}

//Run 运行
func run(a *GoAction) {
	logger := getLogger()

	go func(ga *GoAction) {
		defer func() {
			err := recover()
			if err != nil {
				if ga.promise.Catch != nil {
					ga.promise.Catch(err)
				} else {
					logger.Error("promise panic error", "error", err)
				}
			}

		}()
		ga.promise.Process(a.closeChan)
	}(a)
}

//Close 关闭
func (a *GoAction) Close() {
	a.once.Do(func() {
		close(a.closeChan)
	})
}
