//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package exec

import (
	"fmt"
	"os"
	"strings"

	"idcos.io/cloud-act2/cmd/act2ctl/action/flag"
	"idcos.io/cloud-act2/cmd/act2ctl/aux"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/service/common"
)

func GetTargetHosts(actionFlag flag.ActionFlag, targetIPs []string, osType string) []common.ExecHost {
	var execHosts []common.ExecHost
	for _, ip := range targetIPs {
		execHosts = append(execHosts, common.ExecHost{
			HostID:   "",
			EntityID: "",
			HostIP:   ip,
			HostPort: actionFlag.Port,
			OsType:   osType,
			IdcName:  actionFlag.IDC,
			Encoding: aux.GetEncoding(osType),
		})
	}
	return execHosts
}

func GetScriptType(actionFlag flag.ActionFlag, osType string) string {
	scriptType := strings.TrimSpace(strings.ToLower(actionFlag.ScriptType))
	if len(scriptType) == 0 {
		scriptType = define.BashType
	}
	if scriptType == define.BashType && osType == define.Win {
		scriptType = define.BatType
	}
	return scriptType
}

func GetUsername(actionFlag flag.ActionFlag, osType string) string {
	username := strings.TrimSpace(actionFlag.Username)
	// 如果没有设置用户名，则使用默认用户名，windows: administrator, 其他：root
	if username == "" {
		username = define.RootUser
		if osType == define.Win {
			username = define.AdminUser
		}
	}

	return username
}

func GetPassword(actionFlag flag.ActionFlag, typ string) (string, error) {
	password := actionFlag.Password
	if typ == "ssh" && password == "" {
		pass, err := aux.ReadPassword("target host password")
		if err != nil {
			return "", err
		}
		password = pass
	}
	return password, nil
}

func StartRun(actionFlag flag.ActionFlag, act2CtlParam common.ConfJobIPExecParam) error {
	output := aux.NewOutput(os.Stdout, actionFlag.Verbose)
	output.Verbose("%#v\n", act2CtlParam)

	jobExecutor := NewJobExecutor(output, actionFlag)
	if actionFlag.Async {
		jobRecordID, err := jobExecutor.StartAsync(act2CtlParam)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", jobRecordID)
		return nil
	} else {
		output.Verbose("start sync job\n")
		return jobExecutor.Start(act2CtlParam)
	}
}
