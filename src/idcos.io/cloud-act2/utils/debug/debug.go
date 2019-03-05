//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package debug

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"idcos.io/cloud-act2/log"
)

// SetupDumpStackTrap setups signal trap to dump stack.
func SetupDumpStackTrap() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR2)
	go func() {
		for range c {
			DumpStacks()
		}
	}()
}

func GetDumpStacksBytes() []byte {
	var (
		buf       []byte
		stackSize int
	)
	bufferLen := 16384
	for stackSize == len(buf) {
		buf = make([]byte, bufferLen)
		stackSize = runtime.Stack(buf, true)
		bufferLen *= 2
	}
	buf = buf[:stackSize]
	return buf
}

func Recover() {
	if r := recover(); r != nil {
		v := fmt.Sprintf("%s\n", string(GetDumpStacksBytes()))
		log.L().Named("recover").Error("panic", "stack", v)
	}
}

// DumpStacks dumps the runtime stack.
func DumpStacks() {
	buf := GetDumpStacksBytes()
	// Note that if the daemon is started with a less-verbose log-level than "info" (the default), the goroutine
	// traces won't show up in the log.
	// logrus.Infof("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
	writeDumpStackToFile(buf)
}

func writeDumpStackToFile(buf []byte) {
	dir, err := ioutil.TempDir("", "cloudact2")
	if err != nil {
		return
	}

	pid := os.Getegid()
	now := time.Now()

	filename := fmt.Sprintf("%v-%v", pid, now.Format("2006-01-02-15-04-05"))
	path := filepath.Join(dir, filename)

	err = ioutil.WriteFile(path, buf, os.ModePerm)
	if err != nil {
		return
	}
}
