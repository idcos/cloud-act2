//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"idcos.io/cloud-act2/service/common"
	"strings"
)

func TestJobExec(t *testing.T) {

	// test string null
	exec := common.ExecHost{
		HostIP: "192.168.1.217",
	}
	fmt.Println(exec.HostID == "")

	// test join
	var errMsg []string
	errMsg = append(errMsg, "username can not be null")
	errMsg = append(errMsg, "username can not be null")
	errMsg = append(errMsg, "username can not be null")

	fmt.Println(strings.Join(errMsg, "\n"))

	fmt.Println(math.Abs(10 / 3))

	str := fmt.Sprintf("%0.0f", 3.33)
	fmt.Println(str)
	value, _ := strconv.ParseInt(str, 10, 8)

	fmt.Println(value)
}

func Test(t *testing.T) {
	fmt.Println(100 / 6)
}
