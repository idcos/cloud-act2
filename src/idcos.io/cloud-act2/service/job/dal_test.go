//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"fmt"
	"testing"
)

func TestFindAllIDCHostInfo(t *testing.T) {
	loadConfig()

	strs := []string{"a42a5f55-63c5-2c5e-6fdb-9e938e6b6823", "e6228a11-e069-b8ff-9ee2-bbd23e32c011"}

	hostInfos, err := findHostInfoByHostIDs(strs)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(len(hostInfos))
}