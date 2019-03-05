//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"encoding/json"
	"idcos.io/cloud-act2/server/common"
	common2 "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/quickinstall"
	"idcos.io/cloud-act2/utils/promise"
	"io/ioutil"
	"net/http"
)

//QuickInstallMinion 快速安装minion
func QuickInstallMinion(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	info := common2.InstallInfo{}
	err = json.Unmarshal(bytes, &info)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	result := make(chan common2.InstallResp)
	promise.NewGoPromise(func(chan struct{}) {
		quickinstall.QuickInstall(info, result)
	}, nil)

	done := make(chan int)
	promise.NewGoPromise(func(chan struct{}) {
		installResps := make([]common2.InstallResp, 0, 2)
		for installResp := range result {
			if installResp.Error != nil {
				common.HandleError(w, installResp.Error)
				close(done)
				return
			}

			installResps = append(installResps, installResp)
		}
		common.CommonHandleSuccess(w, installResps)
		close(done)
	}, nil)
	<-done
}
