//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"io/ioutil"
	"log"
	"net/http"
)

// SaltResult salt执行结果
func SaltResult(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("result error:" + err.Error())
	}
	log.Println("get result:" + string(bytes))
}
