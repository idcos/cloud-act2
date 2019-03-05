//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"testing"
	"time"
)

func TestScp(t *testing.T) {
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("******"),
		},
		HostKeyCallback: func(string, net.Addr, ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Duration(2) * time.Second,
	}

	port := "22"
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", "10.0.0.11", port), config)
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		t.Error(err)
		return
	}
	defer session.Close()

	scp := NewScp(session)
	err = scp.PushFile([]byte("hello中文\n"), "/tmp/test22.sh")
	if err != nil {
		fmt.Printf("push file %v\n", err)
		t.Error(err)
	}

}
