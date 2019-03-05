//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package middleware

import (
	"bytes"
	"fmt"
	"idcos.io/cloud-act2/crypto"
	act2Log "idcos.io/cloud-act2/log"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/utils/generator"
	"idcos.io/cloud-act2/utils/stringutil"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var methodList = []string{http.MethodPut, http.MethodDelete, http.MethodPost}

// 需要被脱敏的api信息

type DesensitizingEntry struct {
	// url路径
	Path string
	// 方法
	Method string
	// 需要被脱敏的字符串列表，使用key表示
	Desensitizing []string
}

// 请求信息的脱敏结构
var desensitizingEntries = []DesensitizingEntry{{
	Path:          "/api/v1/job/id/exec",
	Method:        "POST",
	Desensitizing: []string{"password"},
}, {
	Path:          "/api/v1/job/ip/exec",
	Method:        "POST",
	Desensitizing: []string{"password"},
}, {
	Path:          "/api/v1/register",
	Method:        "POST",
	Desensitizing: []string{"password"},
},
}

func ApiLogBefore(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !stringutil.StringScliceContains(methodList, r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		logger := act2Log.L()
		userObj := r.Context().Value("user")
		user := ""
		if userObj != nil {
			userName, ok := userObj.(string)
			if ok {
				user = userName
			}
		}
		byts, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(byts))
		if err != nil {
			logger.Error("read request body fail", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(byts))
		requestParams := ""
		if len(byts) > 0 {
			requestParams = string(byts)
		}

		apiLog := model.Act2APILog{
			ID:            generator.GenUUID(),
			URL:           r.URL.RequestURI(),
			Type:          r.Method,
			User:          user,
			Addr:          r.Host,
			OperateTime:   time.Now(),
			RequestParams: requestParams,
		}

		responser := &Responser{
			w:   w,
			buf: bytes.NewBufferString(""),
		}
		upResponser, ok := w.(*Responser)
		if ok {
			responser = upResponser
		}

		next.ServeHTTP(responser, r)

		apiLogAfter(apiLog, r.URL, responser.buf)

	}
	return http.HandlerFunc(fn)
}

func findDesensitizingEntry(path string, method string) *DesensitizingEntry {
	for _, desensitizingEntry := range desensitizingEntries {
		if desensitizingEntry.Method == method && desensitizingEntry.Path == path {
			return &desensitizingEntry
		}
	}
	return nil
}

func desensitizingData(entry *DesensitizingEntry, data string) string {
	client := crypto.GetClient()

	// TODO: 下面的方法，对空格不具备免疫，用jq是可能是最好的方式
	//desensitizationParam 脱敏参数信息 desValues是需要脱敏的字段列表
	desensitizationParam := func(param, desValues string) string {
		res := param
		if desValues == "" || !strings.Contains(param, ",") {
			return param
		}

		for _, item := range strings.Split(param, ",") {
			if !strings.Contains(item, ":") {
				continue
			}

			k := strings.Split(item, ":")[0]
			v := strings.Split(item, ":")[1]

			if k == "" || v == "" {
				continue
			}

			k = strings.Trim(strings.Trim(k, "["), "{")
			v = strings.Trim(strings.Trim(v, "]"), "}")

			for _, des := range strings.Split(desValues, ",") {
				if strings.Trim(k, "\"") == des {
					coreV := strings.Trim(v, `"`)
					newData := client.Encode(coreV)
					res = strings.Replace(res, v, fmt.Sprintf(`"%s"`, newData), -1)
				}
			}

		}
		return res
	}

	for _, desensitizing := range entry.Desensitizing {
		data = desensitizationParam(data, desensitizing)
	}
	return data
}

func apiLogAfter(apiLog model.Act2APILog, url *url.URL, buffer *bytes.Buffer) {
	logger := act2Log.L()

	responseBody := buffer.String()

	entry := findDesensitizingEntry(url.Path, apiLog.Type)
	if entry != nil {
		apiLog.RequestParams = desensitizingData(entry, apiLog.RequestParams)
	}

	apiLog.TimeConsuming = int(time.Now().Sub(apiLog.OperateTime).Seconds())
	apiLog.ResponseBody = responseBody
	err := apiLog.Save()
	if err != nil {
		logger.Error("save api log fail", "error", err)
	}
}
