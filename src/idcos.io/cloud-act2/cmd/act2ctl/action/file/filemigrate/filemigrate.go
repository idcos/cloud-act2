//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package filemigrate

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"idcos.io/cloud-act2/cmd/act2ctl/aux"
	"idcos.io/cloud-act2/cmd/act2ctl/client"
	"idcos.io/cloud-act2/define"
	common2 "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/complex/common"
	"idcos.io/cloud-act2/service/complex/filemigrate"
)

var flag FileMigrateFlag

const (
	hostInfoRegex = `.+://.+(::.+)?@((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?):.+`
)

func GetFileMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "act2ctl migrate ssh://root::xxx@192.168.1.17:/root/xx.sh|salt://entityId ssh://root::xxx@192.168.1.211:/tmp/xx.sh|salt://entityId --sc 杭州 --tc 北京 -T 300",
		Long:  "Migrate the file of the source host to the target host",
		Args:  cmdArgValidate,
		RunE:  runCmd,
	}

	AddFileMigrateFlags(cmd, &flag)
	return cmd
}

func cmdArgValidate(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("requires at least two args")
	}

	match, err := regexp.MatchString(hostInfoRegex, args[0])
	if err != nil {
		return err
	}
	if !match {
		return errors.New("source host wrong format")
	}

	match, err = regexp.MatchString(hostInfoRegex, args[1])
	if !match {
		return errors.New("target host wrong format")
	}
	return nil
}

func runCmd(cmd *cobra.Command, args []string) error {
	fmt.Printf("start to file migrate, args: (%s), action flag: %#v\n",
		args, flag)

	sourceHost, err := parseHostInfo(args[0])
	if err != nil {
		return err
	}
	if len(sourceHost.password) == 0 {
		pass, err := aux.ReadPassword("source host password")
		if err != nil {
			fmt.Println("read source host password fail")
			return err
		}

		sourceHost.password = pass
	}
	targetHost, err := parseHostInfo(args[1])
	if err != nil {
		return err
	}
	if len(targetHost.password) == 0 {
		pass, err := aux.ReadPassword("target host password")
		if err != nil {
			fmt.Println("read target host password fail")
			return err
		}
		targetHost.password = pass
	}

	info := filemigrate.MasterMigrateInfo{
		SourceHost: common.ComplexHost{
			ExecHost: common2.ExecHost{
				HostIP:   sourceHost.hostIP,
				HostPort: flag.SourcePort,
				EntityID: sourceHost.entityID,
				IdcName:  flag.SourceIDCName,
				OsType:   flag.SourceOSType,
				Encoding: flag.SourceEncoding,
			},
			Username: sourceHost.username,
			Password: sourceHost.password,
			Provider: sourceHost.provider,
		},
		TargetHost: common.ComplexHost{
			ExecHost: common2.ExecHost{
				HostIP:   targetHost.hostIP,
				HostPort: flag.TargetPort,
				EntityID: targetHost.entityID,
				IdcName:  flag.TargetIDCName,
				OsType:   flag.TargetOSType,
				Encoding: flag.TargetEncoding,
			},
			Username: targetHost.username,
			Password: targetHost.password,
			Provider: targetHost.provider,
		},
		TargetFilePath: targetHost.path,
		SourceFilePath: sourceHost.path,
		Timeout:        flag.Timeout,
	}

	_, err = client.GetAct2Client().FileMigrate(info)
	if err != nil {
		return err
	}
	fmt.Println("file migrate success")
	return nil
}

type hostInfo struct {
	provider string
	username string
	password string
	hostIP   string
	path     string
	entityID string
}

func parseHostInfo(arg string) (info hostInfo, err error) {
	providerIndex := strings.Index(arg, "://")
	info.provider = arg[:providerIndex]
	arg = arg[providerIndex+3:]

	switch info.provider {
	case define.MasterTypeSSH:
		userIndex := strings.Index(arg, "@")
		passwordIndex := strings.Index(arg, "::")
		if passwordIndex > 0 {
			info.username = arg[:passwordIndex]
			info.password = arg[passwordIndex+2 : userIndex]
		} else {
			info.username = arg[:userIndex]
		}
		arg = arg[userIndex+1:]

		ipIndex := strings.Index(arg, ":")
		info.hostIP = arg[:ipIndex]
		arg = arg[ipIndex+1:]

		info.path = arg
	case define.MasterTypeSalt:
		info.entityID = arg
	default:
		return hostInfo{}, errors.New("unsupported protocol type")
	}

	return info, nil
}
