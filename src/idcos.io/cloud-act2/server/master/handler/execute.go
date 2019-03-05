//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"net/http"

	"idcos.io/cloud-act2/server/common"
)

//Execute 可以执行的入口
func Execute(w http.ResponseWriter, _ *http.Request) {
	common.HandleSuccess(w, nil)
}
