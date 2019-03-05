//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package define

const (
	// module types
	FileModule   = "file"
	ScriptModule = "script"
	StateModule  = "state"

	// scriptType  url/text/binary
	UrlType    = "url"
	TextType   = "text"
	BinaryType = "binary"

	BashType   = "bash"
	ShellType  = "shell"
	BatType    = "bat"
	PythonType = "python"
	PerlType   = "perl"
	RubyType   = "ruby"
	StateType  = "sls"

	//user Type
	RootUser  = "root"
	AdminUser = "Administrator"

	//MasterTypeSalt master的类型: salt|puppet|ssh
	MasterTypeSalt   = "salt"
	MasterTypeSSH    = "ssh"
	MasterTypePuppet = "puppet"

	// act2ctl
	DefaultRunAs                 = "root"
	JobExecByIpUri               = "/api/v1/job/ip/exec"
	FindJobRecordByIdUri         = "/api/v1/job/record?id="
	FindHostResultsByJobRecordId = "/api/v1/host/result?jobRecordId="

	// host os type
	Win   = "windows"
	Linux = "linux"
	Aix   = "aix"

	// 网络操作系统
	Cisco   = "cisco"
	Junior  = "junior"
	Huawei  = "huawei"
	H3c     = "h3c"
	Network = "network"

	//cache type
	Redis = "redis"

	//crypto type
	CryptoAEAE = "aead"
)
