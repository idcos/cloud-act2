//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package mco

type McoResultBodyData struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Success  bool   `json:"success"`
	ExitCode int    `json:"exitcode"`
}

type McoResultBody struct {
	StatusCode int               `json:"statuscode"`
	StatusMsg  string            `json:"statusmsg"`
	BodyData   McoResultBodyData `json:"data"`
}

type McoResult struct {
	SenderID    string        `json:"senderid"`
	RequestID   string        `json:"requestid"`
	SenderAgent string        `json:"senderagent"`
	MsgTime     int           `json:"msgtime"`
	Body        McoResultBody `json:"body"`
	TTL         int           `json:"ttl"`
	Hash        string        `json:"hash"`
}
