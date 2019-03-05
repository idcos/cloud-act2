//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package host

import (
	"fmt"

	"github.com/spf13/cobra"
	client "idcos.io/cloud-act2/cmd/act2ctl/client"
	"idcos.io/cloud-act2/utils"
)

var (
	infoEntityID string
	infoIP       string
	idcName      string
	infoProxyID  string
)

const (
	hostInfoEntityIDUsage = "find host by entityID"
	hostInfoIPUsage       = "find host by ip"
	hostInfoIDCUsage      = "find host by idc"
	hostInfoProxyIDCUsage = "find host by proxyID"
)

func getHostReportCmd() *cobra.Command {
	hostCmd := &cobra.Command{
		Use:   "report",
		Short: "host info report to master",
		Run:   hostReportAction,
	}
	hostCmd.Flags().StringVarP(&idcName, "idc", "c", "", hostInfoIDCUsage)
	return hostCmd
}

func getHostListCmd() *cobra.Command {
	hostCmd := &cobra.Command{
		Use:   "list",
		Short: "list host info",
		Run:   listHostAction,
	}

	hostCmd.Flags().StringVarP(&infoEntityID, "entityID", "e", "", hostInfoEntityIDUsage)
	hostCmd.Flags().StringVarP(&infoIP, "ip", "i", "", hostInfoIPUsage)
	hostCmd.Flags().StringVarP(&idcName, "idc", "c", "", hostInfoIDCUsage)
	hostCmd.Flags().StringVarP(&infoProxyID, "proxyID", "p", "", hostInfoProxyIDCUsage)

	return hostCmd
}

func GetHostCommand() *cobra.Command {
	hostCmd := &cobra.Command{
		Use:   "host",
		Short: "host info",
	}

	hostCmd.AddCommand(getHostReportCmd())
	hostCmd.AddCommand(getHostListCmd())

	return hostCmd
}

func listHostAction(cmd *cobra.Command, args []string) {

	cli := client.GetAct2Client()
	hostInfos, err := cli.GetAllHostInfo(infoEntityID, idcName, infoProxyID, infoIP)
	if err != nil {
		fmt.Println(err)
	}

	var vals [][]interface{}
	title := []string{"EntityID", "IP", "OsType", "Status", "IDC", "ProxyServer", "ProxyStatus", "ProxyType"}
	for _, host := range hostInfos {
		val := []interface{}{host.EntityID, host.IP, host.OsType, host.Status, host.IDC, host.ProxyServer, host.ProxyStatus, host.ProxyType}
		vals = append(vals, val)
	}
	utils.ArrToStdtab(title, vals)
	return
}

func reportHost(idcName string) {
	cli := client.GetAct2Client()
	err := cli.ReportAgain(idcName, "")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("act2ctl grab host success")
	}
}

func hostReportAction(cmd *cobra.Command, args []string) {
	if idcName == "" {
		fmt.Println("should given idc name")
		return
	}

	cli := client.GetAct2Client()
	err := cli.ReportAgain(idcName, "")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("act2ctl notify host to report master success")
	}
}
