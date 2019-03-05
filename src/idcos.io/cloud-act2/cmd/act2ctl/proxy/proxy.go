//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package proxy

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"idcos.io/cloud-act2/cmd/act2ctl/client"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/utils"
)

var (
	idcName string
)

const (
	proxyUsage = "act2ctl idc proxy [list|del] [option]"
)

//GetProxyCmd get proxy command
func GetProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "proxy commands",
	}

	cmd.AddCommand(getDelProxyCmd())
	cmd.AddCommand(listProxyCmd())

	return cmd
}

func getDelProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del",
		Short: "act2ctl proxy del {proxyId}",
		Long:  "delete proxy",
		Args:  cmdArgValidate,
		RunE:  runDelProxy,
	}

	return cmd
}

func listProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "act2ctl proxy list",
		Long:  "list proxies",
		RunE:  listProxyAction,
	}

	cmd.Flags().StringVarP(&idcName, "idc", "c", "", proxyUsage)

	return cmd
}

func cmdArgValidate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("requires at least one args")
	}
	return nil
}

func runDelProxy(cmd *cobra.Command, args []string) error {
	proxyID := args[0]
	if len(proxyID) == 0 {
		return errors.New("proxy id is required")
	}

	err := client.GetAct2Client().DelProxy(proxyID)
	if err != nil {
		return err
	}

	fmt.Printf("delete proxy %s success\n", proxyID)
	return nil
}

func listProxyAction(cmd *cobra.Command, args []string) error {
	cli := client.GetAct2Client()

	// 查询操作
	var proxies []*common.ProxyInfo
	var err error
	if idcName != "" {
		if proxies, err = cli.IdcProxyList(idcName); err != nil {
			fmt.Println(err.Error())
			return err
		}
	} else {
		if proxies, err = cli.IdcProxyList(""); err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	printProxies(proxies)
	return nil
}

func printProxies(proxies []*common.ProxyInfo) {

	var vals [][]interface{}

	title := []string{"ID", "Server", "Type", "Status", "IdcName"}
	for _, proxy := range proxies {
		val := []interface{}{proxy.ID, proxy.Server, proxy.Type, proxy.Status, proxy.IdcName}
		vals = append(vals, val)
	}
	utils.ArrToStdtab(title, vals)
}
