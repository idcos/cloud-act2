//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package auth

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"idcos.io/cloud-act2/log"
	"idcos.io/cloud-act2/utils/httputil"
)

type SSHAuth interface {
	Auth(host string, port string, user string) (ssh.AuthMethod, error)
}

type NativeAuth struct {
	password string
}

func NewNativeAuth(password string) *NativeAuth {
	return &NativeAuth{
		password: password,
	}
}

func (n *NativeAuth) Auth(host string, port string, user string) (ssh.AuthMethod, error) {
	return ssh.Password(n.password), nil
}

type HttpServer struct {
	server string
}

func (h *HttpServer) request(host string, port string, user string) ([]byte, error) {
	body := map[string]interface{}{
		"host": host,
		"port": port,
		"user": user,
	}
	resp, err := httputil.HttpPost(h.server, body)
	if err != nil {
		return nil, err
	}
	resp = bytes.TrimSpace(resp)

	return resp, nil
}

type HttpAuth struct {
	s HttpServer
}

func NewHttpAuth(server string) *HttpAuth {
	return &HttpAuth{
		s: HttpServer{
			server: server,
		},
	}
}

func (h *HttpAuth) Auth(host string, port string, user string) (ssh.AuthMethod, error) {
	resp, err := h.s.request(host, port, user)
	if err != nil {
		return nil, err
	}

	if bytes.HasPrefix(resp, []byte("{")) {
		var result struct {
			Content string `json:"content"`
			Status  string `json:"status"`
		}

		err := json.Unmarshal(resp, &result)
		if err != nil {
			return nil, err
		}

		if result.Status == "success" || result.Status == "true" {
			return ssh.Password(string(resp)), nil
		} else {
			return nil, errors.New("not found valid password")
		}

	} else {
		return ssh.Password(string(resp)), nil
	}
}

type HttpKeyAuth struct {
	s HttpServer
}

func NewHttpKeyAuth(server string) *HttpKeyAuth {
	return &HttpKeyAuth{
		s: HttpServer{
			server: server,
		},
	}
}

func (h *HttpKeyAuth) Auth(host string, port string, user string) (ssh.AuthMethod, error) {
	logger := log.L().Named("auth")
	resp, err := h.s.request(host, port, user)
	if err != nil {
		logger.Error("request host", "error", err)
		return nil, err
	}

	var der string

	if bytes.HasPrefix(resp, []byte("{")) {
		var result struct {
			Content string `json:"content"`
			Status  string `json:"status"`
		}

		err := json.Unmarshal(resp, &result)
		if err != nil {
			return nil, err
		}

		if result.Status == "success" || result.Status == "true" {
			der = result.Content
		} else {
			return nil, errors.New("not found valid password")
		}

	} else {
		der = string(resp)
	}

	logger.Trace("ssh auth key", "private key", der)

	key, err := ssh.ParsePrivateKey([]byte(der))
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(key), nil
}
