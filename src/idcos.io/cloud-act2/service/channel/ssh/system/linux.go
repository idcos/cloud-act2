//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package system

import (
	"golang.org/x/crypto/ssh"
	"bytes"
	"strings"
)

const (
	linuxUUIDCmd = `cat /sys/class/dmi/id/product_uuid`
)

// 获取linux下的uuid
type LinuxSession struct {
	session *ssh.Session
}

// NewLinuxSession session
func NewLinuxSession(session *ssh.Session) *LinuxSession {
	return &LinuxSession{
		session: session,
	}
}

func (s *LinuxSession) SystemID() (string, error) {
	var b bytes.Buffer
	s.session.Stdout = &b

	if err := s.session.Run(linuxUUIDCmd); err != nil {
		return "", err
	}

	return strings.TrimRight(b.String(), "\r\n"), nil
}
