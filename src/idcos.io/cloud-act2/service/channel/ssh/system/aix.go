//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package system

import (
	"golang.org/x/crypto/ssh"
	"bytes"
)

const (
	aixUUIDCmd = `uname -f`
)

// 获取linux下的uuid
type AixSession struct {
	session *ssh.Session
}

// NewLinuxSession session
func NewAixSession(session *ssh.Session) *AixSession {
	return &AixSession{
		session: session,
	}
}

// 注意，aix的system id返回值是一串没有'-'隔开的字符串
func (s *AixSession) SystemID() (string, error) {
	var b bytes.Buffer
	s.session.Stdout = &b

	if err := s.session.Run(aixUUIDCmd); err != nil {
		return "", err
	}

	return b.String(), nil
}
