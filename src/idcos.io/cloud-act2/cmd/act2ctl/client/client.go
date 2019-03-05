//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package client

import (
	config2 "idcos.io/cloud-act2/cmd/act2ctl/config"
	"idcos.io/cloud-act2/sdk/goact2"
)


func GetAct2Client() *goact2.Client {
	act2ctl := config2.GetAct2ctl()
	client := goact2.NewClient(*act2ctl)
	return client
}
