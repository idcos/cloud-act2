//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package timingwheel

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAddTask(t *testing.T) {
	Load()
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	fmt.Println(time.Now())
	AddTask(5, func() {
		fmt.Println("success")
		fmt.Println(time.Now())
		waitGroup.Add(-1)
	})
	waitGroup.Wait()
}
