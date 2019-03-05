//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package ssh

import (
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/service/channel/common"
	serviceCommon "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/utils/generator"
	"idcos.io/cloud-act2/utils/promise"
	"runtime/debug"
	"sync"
)

type SSHChannel struct {
}

func NewSSHChannel() *SSHChannel {
	return &SSHChannel{}
}

func (c *SSHChannel) hostExecute(params common.ExecScriptParam, minionResult chan common.MinionResult) func(execHost serviceCommon.ExecHost) error {

	logger := getLogger()
	return func(execHost serviceCommon.ExecHost) error {
		r := common.MinionResult{
			HostID: execHost.HostID,
		}

		sshClient := NewSSHClient()
		stdoutBuff, stderrBuff, err := sshClient.Execute(execHost, params)

		if err != nil {
			logger.Error("host execute fail", "error", err)
			r.Status = "fail"
			r.Message = err.Error()
		} else {
			if stdoutBuff != nil {
				r.Stdout = stdoutBuff.String()
			}
			if stderrBuff != nil {
				r.Stderr = stderrBuff.String()
			}

			r.Status = "success"
		}

		minionResult <- r
		return nil
	}
}

//需要存储到cache中
func (c *SSHChannel) Execute(hosts []serviceCommon.ExecHost, params common.ExecScriptParam,
	partitionResult *common.PartitionResult) (string, error) {

	logger := getLogger()

	if params.Pattern == define.StateModule {
		return "", errors.New("unsupported state in ssh provider")
	}

	jid := generator.GenUUID()

	minionResultChan := make(chan common.MinionResult)
	sshExecutor := c.hostExecute(params, minionResultChan)

	promise.NewGoPromise(func(chan struct{}) {
		defer func() {
			logger.Info("ssh execute done", "jid", jid)
			close(minionResultChan)
		}()

		defer func() {
			if err := recover(); err != nil {
				logger.Error("system error occur", "stack", string(debug.Stack()))
			}
		}()

		wg := sync.WaitGroup{}
		wg.Add(len(hosts))

		for _, host := range hosts {
			promise.NewGoPromise(func(chan struct{}) {
				defer wg.Done()

				err := sshExecutor(host)
				if err != nil {
					logger.Error("ssh executor", "error", err)
				}
			}, nil)
		}

		wg.Wait()
	}, nil)

	promise.NewGoPromise(func(close chan struct{}) {
		defer func() {
			logger.Info("minion result done", "jid", jid)
			partitionResult.Close()
		}()

		result := common.Result{
			Jid:    jid,
			Status: "success",
		}

		for r := range minionResultChan {
			result.MinionResults = append(result.MinionResults, r)
		}

		partitionResult.ResultChan <- result
	}, nil)

	return jid, nil
}
