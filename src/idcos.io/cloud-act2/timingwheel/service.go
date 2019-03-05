//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package timingwheel

import (
	"time"

	"github.com/RussellLuo/timingwheel"
)

var wheel *timingwheel.TimingWheel

//Load 初始化
func Load() {
	tw := timingwheel.NewTimingWheel(1*time.Second, 1000)
	wheel = tw
	wheel.Start()
}

func Unload() {
	if wheel != nil {
		wheel.Stop()
	}
}

//AddTask 添加任务
func AddTask(second int, after func()) (timer *timingwheel.Timer) {
	return wheel.AfterFunc(time.Duration(second)*time.Second, after)
}
