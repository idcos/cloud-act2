//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/server/common"
	common2 "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/job"
	"idcos.io/cloud-act2/pubsub"
	"idcos.io/cloud-act2/utils/promise"
)

var logger = getLogger()

func JobStdout(w http.ResponseWriter, r *http.Request) {
	logger := getLogger()

	// init web socket server
	server := getServer()

	// add connect
	server.Grader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	c, err := server.Grader.Upgrade(w, r, nil)

	if err != nil {
		errMsg := fmt.Sprintf("web socket fail to Upgrade, %s", err.Error())
		logger.Error(errMsg)
		common.HandleWebSocketError(c, errors.New(errMsg))
		return
	}

	if config.Conf.CacheType != define.Redis {
		common.HandleWebSocketError(c, errors.New("not support when no redis"))
		return
	}

	jobRecordID := r.FormValue("jobRecordID")
	entityID := r.FormValue("entityID")

	// check realTimeOutput; is false, not realTimeOutput
	var record model.JobRecord
	if err := record.GetByID(jobRecordID); err != nil {
		common.HandleWebSocketError(c, err)
		return
	}
	var param common2.ExecParam
	if err := json.Unmarshal([]byte(record.Parameters), &param); err != nil {
		common.HandleWebSocketError(c, err)
		return
	}

	if !param.RealTimeOutput {
		common.HandleWebSocketError(c, errors.New("job record realTime output is false, do not connect web socket "))
		return
	}

	// if job is done return results
	hostResult, err := job.FindHostResultByRecordIDAndEntityID(jobRecordID, entityID)
	if err != nil {
		errMsg := fmt.Sprintf("act2Master fail to find hostResult,recordID: %s , entityID: %s, error: %s", jobRecordID, entityID, err.Error())
		common.HandleWebSocketError(c, errors.New(errMsg))
		return
	}

	if define.Done == hostResult.ExecuteStatus || define.Timeout == hostResult.ExecuteStatus {
		common.HandleWebSocketSuccess(c, hostResult)
		return
	}

	logger.Info("start to subscribe message", "channel", job.GetPublishID(hostResult.TaskID, entityID))

	client, err := pubsub.GetPubSubClient()
	if err != nil {
		logger.Error("get cache client fail", "error", err)
		common.HandleWebSocketError(c, err)
		return
	}
	promise.NewGoPromise(func(chan struct{}) {
		client.Subscribe(job.GetPublishID(hostResult.TaskID, entityID), server.MsgChan)
	}, nil)

	// listen message
	promise.NewGoPromise(func(chan struct{}) {
		listenMsg(server, c)
	}, nil)

}

func listenMsg(server *Server, c *websocket.Conn) {
	logger.Debug("start listen")
	// listen message && close socket
	for {
		select {
		case isDone := <-server.Done:
			if isDone {
				c.Close()
			}
		case msg := <-server.MsgChan:
			logger.Debug("web socket server get message: %s\n", msg)
			err := c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logger.Error("write message to socket error", "error", err.Error())
			}
		case <-server.Interrupt:
			logger.Debug("web socket interrupted, start to close it ")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logger.Error("write message to socket error", "error", err.Error())
			}
			c.Close()
		}
	}
}

type Server struct {
	Grader    websocket.Upgrader
	MsgChan   chan []byte
	Interrupt chan os.Signal
	Done      chan bool
}

var webSocketServer *Server
var once sync.Once

func getServer() *Server {
	if webSocketServer == nil {
		once.Do(Init)
	}
	return webSocketServer
}

func Init() {
	webSocketServer = &Server{
		Grader:    websocket.Upgrader{},
		MsgChan:   make(chan []byte),
		Interrupt: make(chan os.Signal, 1),
		Done:      make(chan bool),
	}
}
