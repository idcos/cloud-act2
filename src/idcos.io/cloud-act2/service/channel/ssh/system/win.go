//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package system

import (
	"golang.org/x/crypto/ssh"
	"strings"
	"github.com/pkg/errors"
)

const (
	win8UUIDCmd = `wmic csproduct get uuid`
)

// 获取windows下的uuid
type WindowSession struct {
	session *ssh.Session
}

// NewWindowSession session
func NewWindowSession(session *ssh.Session) *WindowSession {
	return &WindowSession{
		session: session,
	}
}

func (s *WindowSession) SystemID() (string, error) {
	out, err := s.session.CombinedOutput(win8UUIDCmd)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	var newLines []string
	// 移除空行
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			newLines = append(newLines, line)
		}
	}

	// 返回两行数据，第一行是UUID，第二行是具体的值
	if len(lines) < 2 || lines[0] != "UUID" {
		return "", errors.New("error not valid output")
	}

	return lines[1], nil
}
