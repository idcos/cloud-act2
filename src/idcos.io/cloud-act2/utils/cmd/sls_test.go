//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package cmd

import (
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/model"
	"testing"
	"fmt"
)

func loadConfig() {
	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")
	model.OpenConn(config.Conf)
}

func TestSaltState_Render(t *testing.T) {
	loadConfig()

	state := NewSaltState()
	template := `userlock:
  user.present:
    - name: '{{name}}'
    - expire: '{{expire_days}}'
    - inactdays: {{days}} 
`
	buff, err := state.Render(template, map[string]interface{}{
		"name":        "damon",
		"expire_days": "5",
		"days":        "day",
	})

	if err != nil {
		t.Error(err)
	}

	fmt.Println(buff.String())
}
