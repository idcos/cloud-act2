//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package exec

import (
	"context"
	"fmt"
	"idcos.io/cloud-act2/utils/dataexchange"
	"os"
	"time"

	"idcos.io/cloud-act2/cmd/act2ctl/action/flag"
	"idcos.io/cloud-act2/cmd/act2ctl/aux"
	"idcos.io/cloud-act2/cmd/act2ctl/client"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/sdk/goact2"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/utils/promise"
)

func buildHostResults(results []goact2.HostResult) []common.HostResultCallback {
	var resultCallback []common.HostResultCallback

	for _, result := range results {
		resultCallback = append(resultCallback, common.HostResultCallback{
			HostIP:   result.HostIP,
			EntityID: result.EntityID,
			IdcName:  result.IdcName,
			Status:   result.Status,
			Message:  result.Message,
			Stderr:   result.StdErr,
			Stdout:   result.StdOut,
		})
	}
	return resultCallback
}

func getResult(jobRecordID string) (*goact2.JobRecordResult, error) {
	cli := client.GetAct2Client()
	result, err := cli.GetJobRecord(jobRecordID)
	return result, err
}

func getHostResults(jobRecordID string) ([]goact2.HostResult, error) {
	cli := client.GetAct2Client()
	hostResults, err := cli.GetJobRecordHostResults(jobRecordID)
	return hostResults, err
}

type JobExecutor struct {
	output     *aux.Output
	actionFlag flag.ActionFlag
	starTime   time.Time
	endTime    time.Time
}

func NewJobExecutor(output *aux.Output, actionFlag flag.ActionFlag) *JobExecutor {
	return &JobExecutor{
		output:     output,
		actionFlag: actionFlag,
	}
}

func (e *JobExecutor) Verbose(format string, args ...interface{}) {
	e.output.Verbose(format, args...)
}

func (e *JobExecutor) sendToExec(act2CtlParam common.ConfJobIPExecParam) (string, error) {
	cli := client.GetAct2Client()
	jobRecordId, err := cli.ExecIPJob(act2CtlParam)
	return jobRecordId, err
}

func (e *JobExecutor) waitForJobResult(ctx context.Context, jobRecordID string, start time.Time) ([]common.HostResultCallback, error) {
	funcDone := make(chan struct{})
	var execHostResult []goact2.HostResult
	var taskErr error

	promise.NewGoPromise(func(done chan struct{}) {
		task := func() ([]goact2.HostResult, error) {
			for {
				result, err := getResult(jobRecordID)
				if err != nil {
					return nil, err
				}

				if define.Done == result.ExecuteStatus {
					hostResults, err := getHostResults(jobRecordID)
					return hostResults, err
				}

				// 每隔1s查询一次
				select {
				case <-done:
					return nil, nil
				case <-ctx.Done():
					return nil, nil
				default:
					time.Sleep(time.Second)
				}
			}
		}
		execHostResult, taskErr = task()
		funcDone <- struct{}{}
	}, nil)

	// wait for complete
	<-funcDone
	if taskErr != nil {
		fmt.Printf("run error %s\n", taskErr)
		return nil, taskErr
	}

	return buildHostResults(execHostResult), nil
}

func (e *JobExecutor) getStringer(results []common.HostResultCallback, start time.Time) fmt.Stringer {
	var stringer fmt.Stringer
	switch e.actionFlag.OutputFormat {
	case "json":
		stringer = aux.NewJSONOutput(results, start, e.output)
	case "yaml":
		stringer = aux.NewYAMLOutput(results, start, e.output)
	default:
		stringer = aux.NewDefaultOutput(results, start)
	}
	return stringer
}

func (e *JobExecutor) dummyResults(results []common.HostResultCallback, startTime time.Time) {
	stringer := e.getStringer(results, startTime)
	e.Verbose("%s\n", dataexchange.ToJsonString(results))
	fmt.Fprintf(os.Stdout, "%s\n", stringer.String())
}

func (e *JobExecutor) StartAsync(jobParam common.ConfJobIPExecParam) (string, error) {
	jobRecordId, err := e.sendToExec(jobParam)
	if err != nil {
		return "", err
	}
	return jobRecordId, nil
}

func (e *JobExecutor) Start(jobParam common.ConfJobIPExecParam) error {
	startTime := time.Now()
	e.starTime = startTime
	e.output.Verbose("start %s\n", startTime)

	jobRecordId, err := e.sendToExec(jobParam)
	e.output.Verbose("job record %s\n", jobRecordId)
	if err != nil {
		fmt.Println(err)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results, err := e.waitForJobResult(ctx, jobRecordId, startTime)
	if err != nil {
		fmt.Println(err)
		return err
	}

	e.dummyResults(results, startTime)

	return nil
}
