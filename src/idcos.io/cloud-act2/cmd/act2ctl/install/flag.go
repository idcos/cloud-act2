//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package install

import "github.com/spf13/cobra"

type InstallFlag struct {
	User     string
	Password string
	OsType   string
	File     string
	IDC      string
	MasterIP string
}

func AddCommonQuickFlags(cmd *cobra.Command, quickFlag *InstallFlag) {
	cmd.PersistentFlags().StringVarP(&quickFlag.User, "user", "u", "", "ssh user")
	cmd.PersistentFlags().StringVarP(&quickFlag.Password, "password", "p", "", "ssh password")
	cmd.PersistentFlags().StringVarP(&quickFlag.File, "file", "f", "", "file url")
	cmd.PersistentFlags().StringVarP(&quickFlag.IDC, "idc", "i", "", "idc name")
	cmd.PersistentFlags().StringVarP(&quickFlag.MasterIP, "masterIp", "m", "", "master ip")
	cmd.PersistentFlags().StringVarP(&quickFlag.OsType, "osType", "o", "linux", "os type")
	return
}
