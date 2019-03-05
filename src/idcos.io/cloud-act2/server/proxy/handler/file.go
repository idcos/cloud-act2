//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package handler

import (
	"github.com/astaxie/beego/httplib"
	"github.com/jlaffaye/ftp"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// 实现很简单，未考虑各种因素，
func FileReversed(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()
	redirects, ok := r.URL.Query()["redirect"]
	if !ok || len(redirects[0]) < 1 {
		logger.Error("not have redirect key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	redirect := redirects[0]

	var reader io.ReadCloser
	var err error
	if strings.HasPrefix(redirect, "ftp") {
		reader, err = getFtpReader(redirect)
	} else {
		reader, err = getHttpReader(redirect)
	}
	if err != nil {
		logger.Error("get response body fail", "path", redirect, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer reader.Close()

	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, reader)
	if err != nil {
		logger.Error("copy stream from remote server ", "error", err)
		w.WriteHeader(500)
		return
	}

	return
}

func getHttpReader(path string) (reader io.ReadCloser, err error) {
	logger := getLogger()

	fileReq := httplib.Get(path)
	resp, err := fileReq.DoRequest()
	if err != nil {
		logger.Error("file reversed", "error", err)
		return
	}

	return resp.Body, nil
}

func getFtpReader(path string) (reader io.ReadCloser, err error) {
	logger := getLogger()

	urlInfo, err := url.Parse(path)
	if err != nil {
		logger.Error("url parse fail", "ftpUrl", path, "error", err)
		return nil, err
	}
	if urlInfo.Scheme != "ftp" {
		return nil, errors.New("url scheme must be ftp")
	}

	addr := urlInfo.Host
	if len(urlInfo.Port()) > 0 {
		addr = addr + ":" + urlInfo.Port()
	} else {
		addr = addr + ":21"
	}

	conn, err := ftp.Connect(addr)
	if err != nil {
		logger.Error("ftp connect fail", "addr", addr, "error", err)
		return nil, err
	}

	user := urlInfo.User
	if len(user.Username()) > 0 {
		password, _ := user.Password()
		conn.Login(user.Username(), password)
	}

	resp, err := conn.Retr(urlInfo.Path)
	if err != nil {
		logger.Error("get fail by ftp fail", "path", urlInfo.Path, "error", err)
		return nil, err
	}

	return resp, nil
}
