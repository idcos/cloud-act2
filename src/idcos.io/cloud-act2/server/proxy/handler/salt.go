//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"idcos.io/cloud-act2/service/salt"
	"net/http"

	"idcos.io/cloud-act2/server/common"
	serviceCommon "idcos.io/cloud-act2/service/common"
)

//HandleSaltEvent salt event process
func HandleSaltEvent(w http.ResponseWriter, r *http.Request) {
	saltEvent := serviceCommon.SaltEvent{}
	err := common.ReadJSONRequest(r, &saltEvent)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	err = salt.ProcessEventMsg(saltEvent)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.HandleSuccess(w, nil)
}
