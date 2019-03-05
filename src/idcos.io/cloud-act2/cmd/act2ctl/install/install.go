//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package install

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"idcos.io/cloud-act2/cmd/act2ctl/aux"
	"idcos.io/cloud-act2/cmd/act2ctl/client"
)

var installFlag InstallFlag

func GetSaltInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "act2ctl install 192.168.1.1,192.168.1.2 -u root -p 111 -o linux -f http://localhost/nfs/salt_install/salt-minion-install.sh -i hangzhou -m 192.168.1.3",
		Long:  "quick install puppet/salt client",
		Args:  cmdArgValidate,
		RunE:  runCmd,
	}

	AddCommonQuickFlags(cmd, &installFlag)
	return cmd
}

func cmdArgValidate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("requires at least one args")
	}

	if len(installFlag.User) == 0 {
		return errors.New("user is required")
	}

	if len(installFlag.Password) == 0 {
		return errors.New("password is required")
	}

	if len(installFlag.File) == 0 {
		return errors.New("file url is required")
	}

	if len(installFlag.IDC) == 0 {
		return errors.New("idc is required")
	}

	if len(installFlag.MasterIP) == 0 {
		return errors.New("masterIp is required")
	}
	return nil
}

func runCmd(cmd *cobra.Command, args []string) error {
	output := aux.NewOutput(os.Stdout, false)
	output.Verbose("start to install, args: (%s), action flag: %#v\n",
		args, installFlag)

	ipList, err := getIpList(args[0])
	if err != nil {
		return err
	}

	err = client.GetAct2Client().QuickInstall(ipList, installFlag.OsType, installFlag.IDC, installFlag.User, installFlag.Password, installFlag.File, installFlag.MasterIP)
	return err
}

func getIpList(ipStr string) (ipList []string, err error) {
	if len(ipStr) == 0 {
		return nil, errors.New("host is required")
	}

	ipList = strings.Split(ipStr, ",")
	return ipList, nil
}
