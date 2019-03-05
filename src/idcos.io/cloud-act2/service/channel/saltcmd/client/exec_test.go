//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package client

import (
	"io"
	"testing"
	"idcos.io/cloud-act2/utils/cmd"
)

func TestCommand(t *testing.T){
	exector := cmd.NewExecutor("ls","-a")

	result := exector.Invoke()
	if(result.Error != nil){
		t.Fatal(result.Error)
	}

	stdout := make([]byte,0,1000)
	for {
		cache := make([]byte,1000)
		len,err := result.Stdout.Read(cache)
		if err != nil {
			if err == io.EOF{
				break
			}
			t.Error(err)
		}
		stdout = append(stdout,cache[:len]...)
	}

	t.Log(stdout)
}