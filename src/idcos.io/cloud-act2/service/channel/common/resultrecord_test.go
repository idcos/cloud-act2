//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"sync"
	"testing"
)

func TestLock (*testing.T){
	var wg sync.WaitGroup
	wg.Add(3)
	go func(){
		c := make([]string,0,1)
		for i := 0; i < 100; i++{
			AddRecord(string(i),c,10)
		}

		for i := 0; i < 100; i++{
			RemoveRecord(string(i))
		}
		wg.Add(-1)
	}()
	go func(){
		c := make([]string,0,1)
		for i := 100; i < 200; i++{
			AddRecord(string(i),c,10)
		}

		for i := 100; i < 200; i++{
			RemoveRecord(string(i))
		}
		wg.Add(-1)
	}()
	go func(){
		c := make([]string,0,1)
		for i := 200; i < 300; i++{
			AddRecord(string(i),c,10)
		}

		for i := 200; i < 300; i++{
			RemoveRecord(string(i))
		}
		wg.Add(-1)
	}()
	wg.Wait()
}