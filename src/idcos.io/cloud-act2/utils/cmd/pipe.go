//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package cmd

import (
	"os/exec"
	"io"
	"bytes"
	"fmt"
)

type pipeData struct {
	name   string
	cmd    *exec.Cmd
	reader *io.PipeReader // reader是前面一个留下的
	writer *io.PipeWriter // writer是自己的
	prev   *pipeData
	next   *pipeData
}

type pipe struct {
	head *pipeData
	last *pipeData
}

func NewPipe(name string, args ...string) *pipe {
	cmd := exec.Command(name, args...)
	data := &pipeData{
		name:   fmt.Sprintf("%s %v", name, args),
		cmd:    cmd,
		reader: nil,
		writer: nil,
		prev:   nil,
		next:   nil,
	}

	p := pipe{
		head: data,
		last: data,
	}
	return &p
}

func (p *pipe) P(name string, args ...string) *pipe {
	// 这个管道是上一个节点的输出和下一个节点的输入
	reader, writer := io.Pipe()

	p.last.cmd.Stdout = writer
	p.last.writer = writer

	cmd := exec.Command(name, args...)
	cmd.Stdin = reader

	data := &pipeData{
		name:   fmt.Sprintf("%s %v", name, args),
		cmd:    cmd,
		reader: reader,
		writer: nil,
		prev:   p.last,
		next:   nil,
	}

	p.last.next = data
	p.last = data

	return p
}

func (p *pipe) CallWithInput(stdin *bytes.Buffer) (*bytes.Buffer, error) {
	p.head.cmd.Stdin = stdin
	return p.Call()
}

// 管道有两个状态，执行中，以及当前的进程执行结束
// 为了方便控制，需要以执行中(Start)或者执行结束(Wait)的其中一个状态作为控制点
// 现在实现以执行结束作为控制点
// 从当前节点running转到当前wait的时候，前一个节点的writer需要关闭
// 下一个节点的reader，这开启中，等到下一次running并等待结束之前，需要把这次的reader给关闭
func (p *pipe) Call() (*bytes.Buffer, error) {
	var buff bytes.Buffer
	p.last.cmd.Stdout = &buff

	// 第一个启动
	h := p.head
	err := h.cmd.Start()
	if err != nil {
		return nil, err
	}

	for ; h.next != nil; h = h.next {
		// 开启下一个节点
		if h.next != nil {
			h.next.cmd.Start()
		}

		// 当前节点从running转到wait
		h.cmd.Wait()

		// 上一次的输入关闭
		if h.prev != nil {
			if h.prev.reader != nil {
				h.prev.reader.Close()
			}
		}

		// 当前节点需要关闭输出
		h.writer.Close()
	}

	// 最后一个节点
	if h != nil {
		h.cmd.Wait()
		// 上一次的输入关闭
		if h.prev != nil {
			if h.prev.reader != nil {
				h.prev.reader.Close()
			}
		}
	}

	p.head = nil
	p.last = nil

	return &buff, nil
}
