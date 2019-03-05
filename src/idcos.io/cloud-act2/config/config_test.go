//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package config

import (
	"testing"
	"fmt"
)

func TestConfigLoad(t *testing.T) {
	filename := "/usr/yunji/cloud-act2/etc/cloud-act2.yaml"
	err := LoadConfig(filename)
	if err != nil {
		t.Errorf("load config error: %s", err)
	}
	fmt.Printf("%v", Conf)
}
