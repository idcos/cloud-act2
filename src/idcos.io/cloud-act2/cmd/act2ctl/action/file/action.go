//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"idcos.io/cloud-act2/utils/encoding"
	"idcos.io/cloud-act2/utils/fileutil"
	"idcos.io/cloud-act2/utils/generator"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"idcos.io/cloud-act2/cmd/act2ctl/action/file/filemigrate"

	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"

	"idcos.io/cloud-act2/cmd/act2ctl/action/exec"
	"idcos.io/cloud-act2/cmd/act2ctl/action/flag"
	"idcos.io/cloud-act2/cmd/act2ctl/aux"
	"idcos.io/cloud-act2/define"
	serviceCommon "idcos.io/cloud-act2/service/common"
)

var (
	actionFlag flag.ActionFlag
)

const (
	runUsage  = "act2ctl run 192.168.1.1,192.168.1.2|-H hostfile -t salt -f /srv/tmp/test.sh|-C 'df -h' [-a arg] [--nocolor] [-output=json|yaml] [-T timeout] [-c idc] [-u username] [-p password] [-P port] [-o osType]"
	fileUsage = "act2ctl file send 192.168.1.1,192.168.1.2|-H hostfile src target -t ssh "
)

func fileSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: fileUsage,
		Long:  `src could be location file or http server file\ndest should be remote host path`,
		Args:  fileArgValidate,
		RunE:  runFile,
	}
	flag.AddCommonRunFlags(cmd, &actionFlag)
	return cmd
}

func GetFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "file operation",
		Long:  `file from one host to other host`,
	}

	cmd.AddCommand(filemigrate.GetFileMigrateCmd())
	cmd.AddCommand(fileSendCmd())
	return cmd
}

func fileArgValidate(cmd *cobra.Command, args []string) error {
	argsLen := len(args)

	if actionFlag.HostFile != "" {
		if argsLen < 2 {
			return errors.New(fmt.Sprintf("file args error, e.g. %s", fileUsage))
		}
	} else {
		if argsLen < 3 {
			return errors.New(fmt.Sprintf("file args error, e.g. %s", fileUsage))
		}
	}

	if !aux.IsValidProvider(actionFlag.Type) {
		return errors.New("not valid type")
	}

	if actionFlag.Type == "ssh" {
		if actionFlag.Username == "" ||
			actionFlag.IDC == "" {
			return errors.New("should given username and idc in ssh mode")
		}
	}

	return nil
}

func getTargetIPs(ips string, actionFlag flag.ActionFlag) ([]string, error) {
	output := aux.NewOutput(os.Stdout, actionFlag.Verbose)
	targetIPs, err := aux.ExtractTargetIPs(ips, actionFlag.HostFile)
	if err != nil {
		output.Printf("target ips extract error")
		return nil, err
	}
	output.Verbose("ips ip address %#v\n", targetIPs)

	targetIPs = aux.DistinctIPs(targetIPs)
	if len(targetIPs) == 0 {
		output.Printf("no ip found\n")
		return nil, errors.New("no ip found")
	}
	return targetIPs, nil
}

func getSSHPassword(typ string, password string) (string, error) {
	var err error

	if typ == "ssh" && password == "" {
		password, err = getPassword()
		if err != nil {
			return "", err
		}
	}

	return password, nil
}

/**
Actions
*/
func runFile(cmd *cobra.Command, args []string) error {
	output := aux.NewOutput(os.Stdout, actionFlag.Verbose)

	output.Verbose("args %v\n", args)

	var ips, srcFile, targetFile string
	if actionFlag.HostFile != "" {
		srcFile = args[0]
		targetFile = args[1]
	} else {
		ips = args[0]
		srcFile = args[1]
		targetFile = args[2]
	}

	srcFile, err := fileutil.ExpandUser(srcFile)
	if err != nil {
		output.Printf("expand user home error")
		return err
	}

	err = checkRunFileParams(ips, srcFile, output)
	if err != nil {
		return err
	}



	var act2CtlParam serviceCommon.ConfJobIPExecParam
	var scriptType string
	var script string

	if strings.HasPrefix(srcFile, "http") || strings.HasPrefix(srcFile, "https") {
		scriptData, _ := json.Marshal([]string{srcFile})
		script = string(scriptData)
		scriptType = "url"
	} else {
		bytes, err := fileutil.ReadFileToBytes(srcFile)
		if err != nil {
			output.Printf("src file read error")
			return err
		}

		script = encoding.DataURIEncode("application/octet-stream", "base64", bytes)
		scriptType = "text"
	}


	if targetFile == "" {
		return errors.New("should given the dest file path")
	}

	params := make(map[string]interface{})
	params["target"] = targetFile
	params["fileName"] = filepath.Base(srcFile)

	typ := strings.TrimSpace(strings.ToLower(actionFlag.Type))
	osType := strings.TrimSpace(strings.ToLower(actionFlag.OsType))
	username := strings.TrimSpace(actionFlag.Username)
	password, err := getSSHPassword(typ, actionFlag.Password)
	if err != nil {
		return err
	}

	targetIPs, err := getTargetIPs(ips, actionFlag)
	if err != nil {
		return err
	}
	execHosts := exec.GetTargetHosts(actionFlag, targetIPs, osType)

	act2CtlParam.ExecHosts = execHosts
	act2CtlParam.ExecParam.Params = params
	act2CtlParam.ExecParam.ScriptType = scriptType
	act2CtlParam.ExecParam.Script = script
	act2CtlParam.ExecParam.Pattern = define.FileModule
	act2CtlParam.ExecParam.RunAs = username
	act2CtlParam.ExecParam.Password = password
	act2CtlParam.ExecParam.Timeout = actionFlag.Timeout
	act2CtlParam.Provider = typ
	act2CtlParam.ExecuteID = generator.GenUUID()
	return exec.StartRun(actionFlag, act2CtlParam)
}

func checkRunFileParams(ips, srcFile string, output *aux.Output) error {
	if !regexp.MustCompile(`^(?:(?:^|,)(?:[0-9]|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])(?:\.(?:[0-9]|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])){3})+$`).MatchString(ips) {
		return errors.New(fmt.Sprintf("file ip args %s error, e.g. %s", ips, fileUsage))
	}

	if !(strings.Contains(srcFile, "http") || strings.Contains(srcFile, "https")) {
		if flag, _ := fileutil.FileExists(srcFile); !flag {
			return errors.New(fmt.Sprintf("source file %s not exists", srcFile))
		}
	}

	if ips == "" && actionFlag.HostFile == "" {
		output.Printf("%s\n", runUsage)
		return errors.New(fmt.Sprintf("file args error, e.g. %s", runUsage))
	}

	return nil
}

func getPassword() (string, error) {
	fmt.Printf("please enter target server password: ")
	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		return "", err
	}

	return string(pass), nil
}
