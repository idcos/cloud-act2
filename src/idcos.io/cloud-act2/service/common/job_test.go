//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"encoding/json"
	"fmt"
	"testing"

	"gopkg.in/go-playground/validator.v9"
)

//
//func TestGetLastTime(t *testing.T) {
//	result := ProxyCallbackHostResult{
//		HostID: "341324123",
//		Status: "success",
//		Stdout: "",
//		Stderr: "",
//	}
//
//	validate := validator.New()
//
//	err := validate.Struct(result)
//	if err != nil {
//
//		for _, err := range err.(validator.ValidationErrors) {
//
//			fmt.Println(err.Namespace())
//			fmt.Println(err.Field())
//			fmt.Println(err.StructNamespace()) // can differ when a custom TagNameFunc is registered or
//			fmt.Println(err.StructField())     // by passing alt name to ReportError like below
//			fmt.Println(err.Tag())
//			fmt.Println(err.ActualTag())
//			fmt.Println(err.Kind())
//			fmt.Println(err.Type())
//			fmt.Println(err.Value())
//			fmt.Println(err.Param())
//			fmt.Println()
//		}
//	}
//
//}

func TestGetLastTime2(t *testing.T) {
	value := `{"status":"fail","message":"execute job fail,err:params not exits args","jobRecordId":"5c877f11-7fa3-630c-ac94-1c1feef83806","hostResults":[]}`

	var result ProxyCallBackResult
	json.Unmarshal([]byte(value), &result)

	Validate := validator.New()
	Validate.Struct(result)

}

func TestProxyCallBackResult(t *testing.T) {
	value := `
{
	"status": "success",
	"jobRecordId": "jobxxxxx",
	"hostResults": null,
	"message": "xxxxxxxxxxxxx"
}
`

	r := ProxyCallBackResult{}
	err := json.Unmarshal([]byte(value), &r)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%#v\n", r)
}
