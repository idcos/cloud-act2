//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package goact2

import "time"

type JobRecordResult struct {
	ID            string    `json:"id"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	ExecuteStatus string    `json:"executeStatus"`
	ResultStatus  string    `json:"resultStatus"`
	Callback      string    `json:"callback"`
	Hosts         string    `json:"hosts"`
	Provider      string    `json:"provider"`
	Script        string    `json:"script"`
	ScriptType    string    `json:"scriptType"`
	Pattern       string    `json:"pattern"`
	Timeout       int       `json:"timeout"`
	Parameters    string    `json:"parameters"`
	MasterID      string    `json:"masterId"`
	User          string    `json:"user"` // 外部调用的用户名
	ExecuteID     string    `json:"executeId"`
}

type HostResult struct {
	JobRecordID string `json:"recordId"`
	HostIP      string `json:"hostIp"`
	HostID      string `json:"hostId"`
	EntityID    string `json:"entityId"`
	OsType      string `json:"osType"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Status      string `json:"status"`
	IdcName     string `json:"idcName"`
	StdOut      string `json:"stdout"`
	StdErr      string `json:"stderr"`
	Message     string `json:"message"`
}
