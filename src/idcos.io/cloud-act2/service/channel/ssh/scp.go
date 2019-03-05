//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package ssh

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"idcos.io/cloud-act2/utils/promise"
	"io"
)

// https://blog.byneil.com/%E6%94%B6%E8%97%8F-scp-secure-copy%E5%8D%8F%E8%AE%AE/
// https://github.com/gnicod/goscplib/blob/master/goscplib.go

//Constants
const (
	SCP_PUSH_BEGIN_FILE       = "C"
	SCP_PUSH_BEGIN_FOLDER     = "D"
	SCP_PUSH_BEGIN_END_FOLDER = " 0"
	SCP_PUSH_END_FOLDER       = "E"
	SCP_PUSH_END              = "\x00"
)

type Scp struct {
	session *ssh.Session
}

func NewScp(session *ssh.Session) *Scp {
	return &Scp{
		session: session,
	}
}

func (s *Scp) GetWriter(dest string) (io.Writer, error) {
	logger := getLogger()

	sftpClient, err := s.getClientPipe()
	if err != nil {
		return nil, err
	}
	file, err := sftpClient.Create(dest)
	if err != nil {
		logger.Error("sftp client create new file", "error", err)
		return nil, err
	}
	return file, nil
}

func (s *Scp) copyFile(w io.Writer, reader io.Reader, length int) error {
	source := fmt.Sprintf("%s%04o %d %s\n", SCP_PUSH_BEGIN_FILE, uint32(0644), length, "test")

	w.Write([]byte(source))
	_, err := io.Copy(w, reader)
	if err != nil {
		return err
	}

	w.Write([]byte(SCP_PUSH_END))
	return nil
}

func (s *Scp) PushFile(content []byte, dest string) error {
	logger := getLogger()

	w, err := s.session.StdinPipe()
	if err != nil {
		logger.Error("push file stdin pipe", "error", err)
		return err
	}

	promise.NewGoPromise(func(chan struct{}) {
		defer w.Close()
		s.copyFile(w, bytes.NewBuffer(content), len(content))
	}, nil)

	// 因为服务器地址不一定是linux，所以不能写死为/usr/bin/scp的地址
	cmd := "scp -ptr " + dest
	if err := s.session.Run(cmd); err != nil {
		return err
	}
	return nil
}

func (s *Scp) PushFileReader(reader io.Reader, dest string) error {
	logger := getLogger()

	sftpClient, err := s.getClientPipe()
	if err != nil {
		return err
	}

	file, err := sftpClient.Create(dest)
	if err != nil {
		logger.Error("sftp client create new file", "error", err, "dest", dest)
		return err
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		logger.Error("sftp client copy file", "error", err)
		return err
	}
	return nil
}

func (s *Scp) getClientPipe() (*sftp.Client, error) {
	logger := getLogger()

	if err := s.session.RequestSubsystem("sftp"); err != nil {
		logger.Error("request sub system", "error", err)
		return nil, err
	}
	pw, err := s.session.StdinPipe()
	if err != nil {
		logger.Error("open session stdin pipe", "error", err)
		return nil, err
	}
	pr, err := s.session.StdoutPipe()
	if err != nil {
		logger.Error("open session stdout pipe", "error", err)
		return nil, err
	}

	sftpClient, err := sftp.NewClientPipe(pr, pw, sftp.MaxConcurrentRequestsPerFile(64))
	if err != nil {
		logger.Error("create new sftp client", "error", err)
		return nil, err
	}

	return sftpClient, nil
}

//OpenRemoteFile 打开远程文件
func (s *Scp) OpenRemoteFile(path string) (file *sftp.File, err error) {
	logger := getLogger()

	sftpClient, err := s.getClientPipe()
	if err != nil {
		logger.Error("get client pipe fail", "error", err)
		return nil, err
	}

	file, err = sftpClient.Open(path)
	if err != nil {
		logger.Error("open remote file fail", "path", path, "error", err)
		return nil, err
	}

	return file, nil
}
