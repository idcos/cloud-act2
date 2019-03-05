//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package mco

import (
	"bytes"
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/utils/generator"
	"path/filepath"

	"encoding/base64"
	"idcos.io/cloud-act2/utils/cmd"
)

func genRequestId() string {
	return fmt.Sprintf("act2-%s", generator.GenUUID())
}

type Message struct {
	requestId string
	puppetCmd string
}

func NewMessage(puppetCmd string) *Message {
	return &Message{
		requestId: GenRequestId(),
		puppetCmd: puppetCmd,
	}
}

func (m *Message) GetRequestId() string {
	return m.requestId
}

// TODO： 下面的实现方法，效率上面会相对比较低，后续可以考虑用rpc方式来加速
func (m *Message) Marshal() ([]byte, error) {
	msg := map[string]string{
		"cmd":       m.puppetCmd,
		"requestid": m.requestId,
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// 默认ruby的地址
	ruby := config.Conf.Puppet.Ruby
	arg := filepath.Join(config.Conf.ProjectPath, "scripts/gen-msg-content.rb")

	executor := cmd.NewExecutor(ruby, arg)
	executor.Stdin = bytes.NewBuffer(msgData)
	r := executor.Invoke()
	if r.Error != nil {
		return nil, err
	}

	// 转换为byte类型的数组
	var rubyMsg []byte
	err = json.Unmarshal(r.Stdout.Bytes(), &rubyMsg)
	if err != nil {
		return nil, err
	}
	return rubyMsg, nil
}

func GenRequestId() string {
	return genRequestId()
}

// 输出的是一个json的字符串
func ParseMessageBody(body []byte) ([]byte, error) {
	// 将数组信息作为输入转换放入ruby代码中
	bodyStr := base64.StdEncoding.EncodeToString(body)

	ruby := config.Conf.Puppet.Ruby
	arg := filepath.Join(config.Conf.ProjectPath, "scripts/parse-msg-response.rb")
	executor := cmd.NewExecutor(ruby, arg)
	executor.Stdin = bytes.NewBufferString(bodyStr)
	r := executor.Invoke()
	if r.Error != nil {
		return nil, r.Error
	}

	return r.Stdout.Bytes(), nil
}

func GetMcoMessageResult(body []byte) (*McoResult, error) {
	bodyData, err := ParseMessageBody(body)
	if err != nil {
		return nil, err
	}

	var mcoResult McoResult

	err = json.Unmarshal(bodyData, &mcoResult)
	if err != nil {
		return nil, err
	}
	return &mcoResult, nil
}

func GetMessageBodyRequestId(body []byte) (string, error) {
	messageBody, err := GetMcoMessageResult(body)
	if err != nil {
		return "", err
	}

	return messageBody.RequestID, nil
}
