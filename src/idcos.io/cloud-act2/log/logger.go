//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package log

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"idcos.io/cloud-act2/config"
	"io"
	"log/syslog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

var (
	mutex       sync.Mutex
	prevent     sync.Once
	logFile     *os.File
	logger      = hclog.NewNullLogger()
	logNotifies []chan struct{}
)

// 打开日志文件
func open(filename string) (*os.File, error) {
	path := filepath.Dir(filename)
	_, err := os.Lstat(path)
	if err != nil {
		pathErr := err.(*os.PathError)
		if pathErr.Err.Error() == "no such file or directory" {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}

func openFileLogger(name string) hclog.Logger {
	mutex.Lock()
	defer mutex.Unlock()

	// 先将旧的保存下来，close必须放到所有的步骤之后
	oldLogFile := logFile
	defer func() {
		if oldLogFile != nil {
			oldLogFile.Close()
		}
	}()

	cnfLog := config.Conf.Logger
	file, err := open(cnfLog.LogFile)
	if err != nil {
		fmt.Printf("open log file %s, error %v\n", cnfLog.LogFile, err)
		return nil
	}

	logFile = file
	l := getLogger(name, file)

	// 通知外部所有所有需要修改logger的地方
	for _, notify := range logNotifies {
		notify <- struct{}{}
	}

	return l
}

func initFileLog(name string) {
	logger = openFileLogger(name)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)

	reload := func() {
		logger.Debug("get log reload signal")
		logger = openFileLogger(name)
	}

	go func() {
		for range c {
			reload()
		}
	}()

}

func getLogger(name string, writer io.Writer) hclog.Logger {
	cnfLog := config.Conf.Logger
	logger = hclog.New(&hclog.LoggerOptions{
		Name:            name,
		Level:           hclog.LevelFromString(cnfLog.LogLevel),
		Output:          writer,
		IncludeLocation: true,
		TimeFormat:      cnfLog.LogDateFormat,
	})
	return logger
}

func InitLogger(name string) hclog.Logger {
	prevent.Do(func() {
		cnfLog := config.Conf.Logger
		if cnfLog.Facility == "" || cnfLog.Facility == "file" {
			initFileLog(name)
		} else if cnfLog.Facility == "syslog" {
			writer, err := syslog.New(syslog.LOG_DEBUG|syslog.LOG_LOCAL6, config.Conf.GetName())
			if err != nil {
				fmt.Printf("open syslog error %s", err)
			}
			logger = getLogger(name, writer)
		} else if cnfLog.Facility == "rsyslog" {
			if cnfLog.LogProtocol != "tcp" && cnfLog.LogProtocol != "udp" {
				fmt.Printf("unknown log protocol, exit\n")
				return
			}
			writer, err := syslog.Dial(cnfLog.LogProtocol, cnfLog.LogServer, syslog.LOG_DEBUG|syslog.LOG_LOCAL6,
				config.Conf.GetName())
			if err != nil {
				fmt.Printf("open syslog error %s", err)
			}

			logger = getLogger(name, writer)
		}

	})

	return logger
}

func L() hclog.Logger {
	return logger
}

func Writer() io.Writer {
	mutex.Lock()
	defer mutex.Unlock()
	return logFile
}

// 必须在程序初始化的时候添加
func AddLogNotify(log chan struct{}) {
	logNotifies = append(logNotifies, log)
}
