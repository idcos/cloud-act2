//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package oom

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/hashicorp/go-hclog"
	"idcos.io/cloud-act2/log"
)

// Writes 'value' to /proc/<pid>/oom_score_adj. PID = 0 means self
// Returns os.ErrNotExist if the `pid` does not exist.
func ApplyOOMScoreAdj(pid int, oomScoreAdj int) error {
	logger := log.L()

	if pid < 0 {
		return fmt.Errorf("invalid PID %d specified for oom_score_adj", pid)
	}

	var pidStr string
	if pid == 0 {
		pidStr = "self"
	} else {
		pidStr = strconv.Itoa(pid)
	}

	maxTries := 2
	oomScoreAdjPath := path.Join("/proc", pidStr, "oom_score_adj")

	// 在文件存在情况下处理oom score adjust
	if _, err := os.Stat(oomScoreAdjPath); err != nil {
		if os.IsNotExist(err) {
			return err
		}
	}

	value := strconv.Itoa(oomScoreAdj)
	logger.Info("attempting to set", "oomScoreAdj", hclog.Fmt("%q", oomScoreAdjPath), "value", hclog.Fmt("%q", value))
	var err error
	for i := 0; i < maxTries; i++ {
		err = ioutil.WriteFile(oomScoreAdjPath, []byte(value), 0700)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Info(fmt.Sprintf("%q does not exist", oomScoreAdjPath))
				return os.ErrNotExist
			}

			logger.Error("write oom score adjust", "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return nil
	}
	if err != nil {
		logger.Error("failed to set", "oomScoreAdj", hclog.Fmt("%q", oomScoreAdjPath), "value", hclog.Fmt("%q", value), "error", err)
	}
	return err
}
