//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package build

import "bytes"

var (
	// Date 编译时间
	Date string
	// Commit git提交ID
	Commit string
	// GitBranch git的分支信息
	GitBranch string
	// GoVersion go version
)

// Version 生成版本信息
func Version(prefix string) string {
	var buf bytes.Buffer
	if prefix != "" {
		buf.WriteString(prefix)
	}
	if Date != "" {
		buf.WriteByte('\n')
		buf.WriteString("\tdate: ")
		buf.WriteString(Date)
	}
	if Commit != "" {
		buf.WriteByte('\n')
		buf.WriteString("\tcommit: ")
		buf.WriteString(Commit)
	}
	if GitBranch != "" {
		buf.WriteByte('\n')
		buf.WriteString("\tbranch: ")
		buf.WriteString(GitBranch)
	}
	return buf.String()
}
