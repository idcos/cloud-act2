//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package flag

import "github.com/spf13/cobra"

type ActionFlag struct {
	HostFile     string
	Type         string
	SrcFile      string
	Args         string
	Username     string
	Password     string
	Port         int
	OsType       string
	IDC          string
	ScriptType   string
	Encoding     string
	Command      string
	Timeout      int
	OutputFormat string
	NoColor      bool
	Verbose      bool
	Async        bool
}

func AddCommonRunFlags(cmd *cobra.Command, actionFlag *ActionFlag) {
	cmd.PersistentFlags().StringVarP(&actionFlag.HostFile, "hostfile", "H", "", "host file information")
	cmd.PersistentFlags().StringVarP(&actionFlag.IDC, "idc", "c", "", "idc information")
	cmd.PersistentFlags().StringVarP(&actionFlag.Type, "type", "t", "salt", "should be one of salt|puppet|ssh")
	cmd.PersistentFlags().StringVarP(&actionFlag.SrcFile, "file", "f", "", "script file")
	cmd.PersistentFlags().StringVarP(&actionFlag.Args, "arg", "a", "", "script run arguments")
	cmd.PersistentFlags().StringVarP(&actionFlag.Username, "username", "u", "", "ssh login username, only used in when type is ssh")
	cmd.PersistentFlags().StringVarP(&actionFlag.Password, "password", "p", "", "ssh login password, only used in when type is ssh")
	cmd.PersistentFlags().IntVarP(&actionFlag.Port, "port", "P", 22, "ssh login  port, only used in when type is ssh")
	cmd.PersistentFlags().StringVarP(&actionFlag.OsType, "osType", "o", "linux", "remote host os type, only used in when type is ssh, should be windows|linux|aix")
	cmd.PersistentFlags().StringVarP(&actionFlag.ScriptType, "scriptType", "s", "shell", "script type, could be: bash,shell(equal bash),python,bat,sls, windows default bat, linux or aix: default bash")
	cmd.PersistentFlags().StringVarP(&actionFlag.Encoding, "encoding", "e", "", "encoding, gb18030,utf-8, windows default gb18030, linux or aix default utf-8")
	cmd.PersistentFlags().StringVarP(&actionFlag.Command, "command", "C", "", "the command will be run on remote host")
	cmd.PersistentFlags().IntVarP(&actionFlag.Timeout, "timeout", "T", 300, "run script or command timeout， 0 no timeout, unit: second")
	cmd.PersistentFlags().StringVarP(&actionFlag.OutputFormat, "output", "", "", "output format, default table, cloud be yaml|json")
	cmd.PersistentFlags().BoolVarP(&actionFlag.NoColor, "nocolor", "", false, "output without color")
	cmd.PersistentFlags().BoolVarP(&actionFlag.Verbose, "verbose", "v", false, "verbose")
	cmd.PersistentFlags().BoolVarP(&actionFlag.Async, "async", "", false, "async")
	return
}
