//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package idc

import (
	"fmt"

	"github.com/spf13/cobra"
	"idcos.io/cloud-act2/cmd/act2ctl/client"
	"idcos.io/cloud-act2/utils"
)

const (
	idcUsage   = "act2ctl idc [-hl]"
	hostUsage  = "act2ctl idc host [-lgc]"
	proxyUsage = "act2ctl idc proxy [-ldc]"
)

const (
	hostInfoEntityIDUsage = "find host by entityID"
	hostInfoIPUsage       = "find host by ip"
	hostInfoIDCUsage      = "find host by idc"
	hostInfoProxyIDCUsage = "find host by proxyID"
)

var (
	isAll   bool   // idc列表
	idcName string // idc名称
	isGrab  bool   //是否抓取
	proxyID string // proxyID
)

var (
	infoEntityID string
	infoIP       string
	infoIDCName  string
	infoProxyID  string
)

// idc 下面的命令
func getIDCommand() *cobra.Command {
	idcCmd := &cobra.Command{
		Use: "list",
		Run: idcListAction,
	}

	return idcCmd
}

func GetIDCCmd() *cobra.Command {
	idcCmd := &cobra.Command{
		Use:   "idc",
		Short: "idc information",
	}

	idcCmd.AddCommand(getIDCommand())
	return idcCmd
}

func idcListAction(cmd *cobra.Command, args []string) {
	cli := client.GetAct2Client()
	idcList, err := cli.IdcList()
	if err != nil {
		fmt.Println(err)
	}

	title := []string{"ID", "Name", "AddTime"}
	var vals [][]interface{}
	for _, idc := range idcList {
		val := []interface{}{idc.ID, idc.Name, idc.AddTime}
		vals = append(vals, val)
	}
	utils.ArrToStdtab(title, vals)
}
