//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package record

import (
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/utils/dataexchange"
	"os"

	"github.com/spf13/cobra"
	"idcos.io/cloud-act2/cmd/act2ctl/aux"
	"idcos.io/cloud-act2/cmd/act2ctl/client"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/utils"
)

const (
	recordListUsage     = "act2ctl record list -n 1 -p 10 -v"
	recordIDUsage       = "act2ctl record  get jobRecordID -v"
	recordCancelUsage   = "act2ctl record  cancel jobRecordID -v"
	recordIDFormatUsage = "job format, t: table, c: column, default: t"
)

var (
	isAll    bool   // idc列表
	recordID string // 作业记录
)

var (
	infoEntityID string
	infoIP       string
	infoIDCName  string
	infoProxyID  string
	pageNum      int
	pageSize     int
	verbose      bool
	jobFormat    string
)

//GetRecordCmd Record命令
func GetRecordCmd() *cobra.Command {
	recordCmd := &cobra.Command{
		Use:   "record",
		Short: "job record of act2",
	}

	recordCmd.AddCommand(recordListCmd())
	recordCmd.AddCommand(recordGetCmd())

	return recordCmd
}

//recordListCmd 分页查询任务列表
func recordListCmd() *cobra.Command {
	rListCmd := &cobra.Command{
		Use:   "list",
		Short: "-l",

		Run: recordListAction,
	}

	rListCmd.Flags().IntVarP(&pageNum, "pageNum", "n", 1, recordListUsage)
	rListCmd.Flags().IntVarP(&pageSize, "pageSize", "p", 10, recordListUsage)
	//rListCmd.Flags().BoolVarP(&verbose, "verboase", "v", false, recordListUsage)

	return rListCmd
}

//recordGetCmd 根据ID获取任务信息
func recordGetCmd() *cobra.Command {
	rIDCmd := &cobra.Command{
		Use: "get",
		Run: recordIDAction,
	}
	rIDCmd.Flags().StringVarP(&jobFormat, "format", "f", "t", recordIDFormatUsage)
	//rIDCmd.Flags().BoolVarP(&verbose, "verboase", "v", false, recordListUsage)
	return rIDCmd
}

//recordIDAction 根据ID获取任务信息Action
func recordIDAction(cmd *cobra.Command, args []string) {
	output := aux.NewOutput(os.Stdout, verbose)
	cli := client.GetAct2Client()

	if len(args) <= 0 {
		// 打印信息
		fmt.Printf("should given job record. usage actctl record get jobRecordId,%s\n", recordIDUsage)
	}

	recordID := args[0]
	records, err := cli.FindRecordResultByID(recordID)
	if err != nil {
		output.Verbose("FindRecordResultByID  error; %s\n", err.Error())
	}

	if jobFormat == "column" || jobFormat == "c" {
		for _, record := range records {
			printSingleRecord(record)
			fmt.Println("\n")
		}
		return
	}
	printRecords(records)
}

//printSingleRecord 打印单条任务信息
func printSingleRecord(record model.RecordResult) {
	fmt.Printf("JobRecordID: %s\n", record.JobRecordID)
	fmt.Printf("EntityID: %s\n", record.EntityID)
	fmt.Printf("hostIP: %s\n", record.HostIP)
	fmt.Printf("startTime: %s\n", record.StartTime)
	fmt.Printf("endTime: %s\n", record.EndTime)
	fmt.Printf("status: %s\n", record.Status)
	fmt.Printf("stdout: %s\n", record.StdOut)
	fmt.Printf("stderr: %s\n", record.StdErr)
}

//recordListAction 任务列表Action
func recordListAction(cmd *cobra.Command, args []string) {
	output := aux.NewOutput(os.Stdout, verbose)

	cli := client.GetAct2Client()

	var records []model.RecordResult

	pagination, err := cli.RecordList(pageNum, pageSize)

	if err != nil {
		fmt.Println(err)
	}

	output.Verbose("FindRecordResults  success; %s\n", dataexchange.ToJsonString(pagination))

	if err := json.Unmarshal([]byte(dataexchange.ToJsonString(pagination.List)), &records); err != nil {
		fmt.Printf("convert records err, %s", err.Error())
	}

	output.Verbose("convert record result success; %s\n", dataexchange.ToJsonString(records))
	printRecords(records)
}

//printRecords 打印多条任务列表
func printRecords(records []model.RecordResult) {
	header := []string{"recordID", "entityID", "hostIP", "startTime", "endTime", "status"}
	var vals [][]interface{}
	for _, record := range records {
		vals = append(vals, []interface{}{
			record.JobRecordID,
			record.EntityID,
			record.HostIP,
			record.StartTime,
			record.EndTime,
			record.Status,
		})

	}

	utils.ArrToStdtab(header, vals)

}
