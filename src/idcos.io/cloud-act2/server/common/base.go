//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"encoding/json"
	"idcos.io/cloud-act2/define"
	"net/http"

	"github.com/gorilla/websocket"

	"io/ioutil"
)

// JSONResult REST API的JSON返回值结构体
type JSONResult struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Content interface{} `json:"content"`
}

//NewJSONResult json result
func NewJSONResult(status string, message string, content interface{}) *JSONResult {
	return &JSONResult{
		Status:  status,
		Message: message,
		Content: content,
	}
}

// NewJSONMapResult 返回JSON返回值结构体指针
func NewJSONMapResult(status string, message string, content map[string]interface{}) *JSONResult {
	return &JSONResult{
		Status:  status,
		Message: message,
		Content: content,
	}
}

// HandleJSONResponse json response process
func HandleJSONResponse(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

//HandleError error response
func HandleError(w http.ResponseWriter, err error) {
	b, _ := json.Marshal(NewJSONResult(define.Error, err.Error(), nil))
	HandleJSONResponse(w, b)
}

//HandleSuccess success response
func HandleSuccess(w http.ResponseWriter, content map[string]interface{}) {
	b, _ := json.Marshal(NewJSONResult(define.Success, "", content))
	HandleJSONResponse(w, b)
}

// CommonHandleSuccess success response
func CommonHandleSuccess(w http.ResponseWriter, content interface{}) {
	b, _ := json.Marshal(NewJSONResult(define.Success, "", content))
	HandleJSONResponse(w, b)
}

func HandleWebSocketSuccess(c *websocket.Conn, v interface{}) {
	b, _ := json.Marshal(NewJSONResult(define.Success, "", v))
	c.WriteMessage(websocket.TextMessage, b)
	c.Close()
}

func HandleWebSocketError(c *websocket.Conn, err error) {
	b, _ := json.Marshal(NewJSONResult(define.Error, err.Error(), nil))
	c.WriteMessage(websocket.TextMessage, b)
	c.Close()
}

//读取json的内容
func ReadJSONRequest(r *http.Request, v interface{}) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}
