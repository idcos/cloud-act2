//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package cmd

import (
	"errors"
	"idcos.io/cloud-act2/utils/fileutil"
	"idcos.io/cloud-act2/utils/generator"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"idcos.io/cloud-act2/cmd/act2ctl/action/exec"
	"idcos.io/cloud-act2/cmd/act2ctl/action/flag"
	"idcos.io/cloud-act2/cmd/act2ctl/aux"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/service/common"
)

var (
	actionFlag flag.ActionFlag
)

func GetRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "act2ctl run 192.168.1.1,192.168.1.2|-H hostfile -t salt -f /srv/tmp/test.sh|-C 'df -h' [-a arg] [--nocolor] [-output=json|yaml] [-T timeout] [-c idc] [-u username] [-p password] [-P port] [-o osType]",
		Long:  `run command to the remote machine, support salt,ssh protocol`,
		Args:  cmdArgValidate,
		RunE:  runCmd,
	}

	flag.AddCommonRunFlags(cmd, &actionFlag)
	return cmd
}

func cmdArgValidate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("requires at least one args")
	}

	if !aux.IsValidProvider(actionFlag.Type) {
		return errors.New("not valid type")
	}

	return nil
}

func runCmd(cmd *cobra.Command, args []string) error {
	output := aux.NewOutput(os.Stdout, actionFlag.Verbose)
	output.Verbose("start to run action, args: (%s), action flag: %#v\n",
		args, actionFlag)

	target := args[0]

	targetIPs, err := aux.ExtractTargetIPs(target, actionFlag.HostFile)
	if err != nil {
		return err
	}

	srcFile, err := fileutil.ExpandUser(actionFlag.SrcFile)
	if err != nil {
		return err
	}

	output.Verbose("run action target ips: %s\n", targetIPs)
	output.Verbose("start to check file , srcFile: %s\n", srcFile)

	var act2CtlParam common.ConfJobIPExecParam

	command, err := aux.ExtractCommand(actionFlag.Command, srcFile)
	if err != nil {
		return err
	}

	typ := strings.TrimSpace(strings.ToLower(actionFlag.Type))
	osType := strings.TrimSpace(strings.ToLower(actionFlag.OsType))

	params := make(map[string]interface{})
	params["args"] = actionFlag.Args

	osType = aux.GetRealOsType(osType)
	execHosts := exec.GetTargetHosts(actionFlag, targetIPs, osType)
	scriptType := exec.GetScriptType(actionFlag, osType)

	act2CtlParam.ExecHosts = execHosts
	act2CtlParam.Callback = ""

	pattern := define.ScriptModule
	if define.StateType == scriptType {
		pattern = define.StateModule
	}
	username := exec.GetUsername(actionFlag, osType)
	password, err := exec.GetPassword(actionFlag, typ)
	if err != nil {
		return err
	}

	act2CtlParam.ExecParam.ScriptType = scriptType
	act2CtlParam.ExecParam.Params = params
	act2CtlParam.ExecParam.Script = command
	act2CtlParam.ExecParam.Pattern = pattern
	act2CtlParam.ExecParam.RunAs = username
	act2CtlParam.ExecParam.Password = password
	act2CtlParam.ExecParam.Timeout = actionFlag.Timeout
	act2CtlParam.ExecParam.RealTimeOutput = false
	act2CtlParam.Provider = typ
	act2CtlParam.ExecuteID = generator.GenUUID()

	return exec.StartRun(actionFlag, act2CtlParam)
}
