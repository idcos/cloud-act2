//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"encoding/json"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/server/common"
	serviceCommon "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/webhook"
	"io/ioutil"
	"net/http"
)

//AddWebHook 添加web钩子
func AddWebHook(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	form := serviceCommon.WebHookAddForm{}
	err = json.Unmarshal(bytes, &form)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	err = checkWebHookParams(form.Event, form.Uri)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	err = webhook.AddWebHook(form)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, "add success")
}

//UpdateWebHook 更新web钩子
func UpdateWebHook(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	form := serviceCommon.WebHookUpdateForm{}
	err = json.Unmarshal(bytes, &form)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	err = checkWebHookParams(form.Event, form.Uri)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if len(form.ID) == 0 {
		common.HandleError(w, errors.New("id is required"))
		return
	}

	err = webhook.UpdateWebHook(form)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, "update success")
}

func checkWebHookParams(event, uri string) (err error) {
	if len(event) == 0 {
		return errors.New("event is required")
	}

	if len(uri) == 0 {
		return errors.New("uri is required")
	}

	return nil
}

//DeleteWebHook 删除web钩子
func DeleteWebHook(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		common.HandleError(w, errors.New("id is required"))
		return
	}

	err := webhook.DeleteByID(id)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.CommonHandleSuccess(w, "delete success")
}
