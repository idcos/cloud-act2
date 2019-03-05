//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package mco

import (
	"idcos.io/cloud-act2/utils/generator"
	"idcos.io/cloud-act2/utils/promise"
	"net/url"
	"strings"
	"sync"

	"github.com/streadway/amqp"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/service/channel/common"
	serviceCommon "idcos.io/cloud-act2/service/common"
)

const (
	executeExchange = "mcollective_directed"
	puppetMQOpen    = "puppet mq open channel"
)

type McoRabbitMQClient struct {
	rabbitMQServer string
	replyQueue     string
	routerKey      string
	conn           *amqp.Connection
	once           sync.Once
}

func NewMcoRabbitMQClient(rabbitMQServer string, replyQueue string) *McoRabbitMQClient {
	uri, err := url.Parse(rabbitMQServer)
	if err != nil {
		return nil
	}

	routerKey := strings.Split(uri.Host, ":")[0]

	puppetMQ := &McoRabbitMQClient{
		rabbitMQServer: rabbitMQServer,
		replyQueue:     replyQueue,
		routerKey:      routerKey,
	}
	return puppetMQ
}

func (p *McoRabbitMQClient) Connect() error {
	var err error
	p.once.Do(func() {
		logger := getLogger()

		var conn *amqp.Connection
		conn, err = amqp.Dial(p.rabbitMQServer)
		if err != nil {
			logger.Error("connect rabbitmq", "error", err)
			return
		}

		p.conn = conn
	})

	return err
}

func (p *McoRabbitMQClient) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}

func (p *McoRabbitMQClient) CreateReplyQueue() error {
	logger := getLogger()

	ch, err := p.conn.Channel()
	if err != nil {
		logger.Error(puppetMQOpen, "error", err)
		return err
	}

	err = ch.ExchangeDeclare(executeExchange, amqp.ExchangeDirect, true, false, false, false, nil)
	if err != nil {
		logger.Error("declare exchange", "error", err)
		return err
	}

	q, err := ch.QueueDeclare(p.replyQueue, true, false, false, false, nil)
	if err != nil {
		logger.Error("declare queue", "error", err)
		return err
	}

	err = ch.QueueBind(q.Name, p.routerKey, executeExchange, false, nil)
	if err != nil {
		logger.Error("queue bind", "error", err)
		return err
	}

	return nil

}

func (p *McoRabbitMQClient) CloseReplyQueue() error {
	logger := getLogger()

	ch, err := p.conn.Channel()
	if err != nil {
		logger.Error(puppetMQOpen, "error", err)
		return err
	}

	_, err = ch.QueueDelete(p.replyQueue, true, false, false)
	if err != nil {
		logger.Error("delete reply queue", "error", err)
		return err
	}

	return nil
}

func (p *McoRabbitMQClient) SendMessage(routingKey string, message *Message) error {
	logger := getLogger()

	body, err := message.Marshal()
	if err != nil {
		logger.Error("message marshal", "error", err)
		return err
	}

	return p.Send(routingKey, body)
}

func (p *McoRabbitMQClient) Send(routingKey string, body []byte) error {

	logger := getLogger()

	ch, err := p.conn.Channel()
	if err != nil {
		logger.Error(puppetMQOpen, "error", err)
		return err
	}

	// declare queue
	args := make(amqp.Table)
	// 回调过来的目标的对列名称，这个对列的消息消费方，需要通过body的消息体中的
	args["reply-to"] = p.replyQueue
	args["mc_sender"] = "cloud-act2"
	// 官方expiration的实现是： (msg.ttl + 10) * 1000
	// 官方ttl默认为60s
	args["expiration"] = "70000"

	err = ch.Publish(
		executeExchange, // exchange
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        body,
			Headers:     args,
		})
	if err != nil {
		logger.Error("puppet mq publish message", "error", err)
		return err
	}
	return nil
}

func (p *McoRabbitMQClient) Consume() (<-chan amqp.Delivery, error) {
	logger := getLogger()

	ch, err := p.conn.Channel()
	if err != nil {
		logger.Error(puppetMQOpen, "error", err)
		return nil, err
	}

	msgs, err := ch.Consume(
		p.replyQueue,
		"",
		false,
		false,
		false,
		true,
		nil,
	)
	if err != nil {
		logger.Error("message consume", "error", err)
		return nil, err
	}

	return msgs, nil
}

type McoClient struct {
}

func NewMcoClient() *McoClient {
	return &McoClient{}
}

type MinionResult struct {
	HostID    string
	Status    string
	Stdout    string
	Stderr    string
	Message   string
	requestId string
	done      bool
}

// 等待所有的结果返回，目前没有超时，所以，可能存在，消息不能够全部获取到的情况
func waitResult(client *McoRabbitMQClient, completedCount int, execHosts []serviceCommon.ExecHost, minionResultMap *map[string]MinionResult, done chan bool, close chan struct{}) {
	logger := getLogger()

	msgs, err := client.Consume()
	if err != nil {
		return
	}

	for completedCount < len(execHosts) {
		select {
		case <-close:
			logger.Info("job interrupted")
			done <- true
			return
		case d := <-msgs:
			//content-length是小写
			_, ok := d.Headers["content-length"]
			if !ok {
				break
			}

			mcoResult, err := GetMcoMessageResult(d.Body)
			if err != nil {
				logger.Error("get message body", "error", err)
				break // break select
			}

			r := (*minionResultMap)[mcoResult.RequestID]
			r.Stdout = mcoResult.Body.BodyData.Stdout
			r.Stderr = mcoResult.Body.BodyData.Stderr
			r.Status = mcoResult.Body.StatusMsg
			r.done = true
			(*minionResultMap)[mcoResult.RequestID] = r

			completedCount += 1

		default:
			// check 是否所有的requestMessages都处理完了？

		}
	}
	done <- true
}

func (mco *McoClient) Execute(execHosts []serviceCommon.ExecHost, script string, partitionResult *common.PartitionResult) (string, error) {
	logger := getLogger()

	jid := generator.GenUUID()

	client := NewMcoRabbitMQClient(config.Conf.Puppet.RabbitMQ, jid)
	err := client.Connect()
	if err != nil {
		return "", err
	}
	err = client.CreateReplyQueue()
	if err != nil {
		return "", err
	}

	promise.NewGoPromise(func(close chan struct{}) {
		defer client.CloseReplyQueue()
		defer client.Close()

		var results []MinionResult

		minionResultMap := map[string]MinionResult{}
		completedCount := 0
		for _, execHost := range execHosts {
			// 脚本执行
			message := NewMessage(script)
			r := MinionResult{
				requestId: message.GetRequestId(),
				HostID:    execHost.HostID,
			}

			err := client.SendMessage(execHost.EntityID, message)
			if err != nil {
				r.Message = err.Error()
				completedCount += 1

				// 记录是否发送结束
				r.done = true
			}

			results = append(results, r)
			minionResultMap[message.GetRequestId()] = r
		}

		done := make(chan bool)

		// 等待所有的结果返回，目前没有超时，所以，可能存在，消息不能够全部获取到的情况
		promise.NewGoPromise(func(close chan struct{}) {
			waitResult(client, completedCount, execHosts, &minionResultMap, done, close)
		}, func(message interface{}) {
			logger.Error("promise panic error", "error", message)
		})
		<-done

		var minionResults []common.MinionResult

		for _, mr := range minionResultMap {
			r := common.MinionResult{
				HostID:  mr.HostID,
				Status:  conversionStatus(mr.Status),
				Stdout:  mr.Stdout,
				Stderr:  mr.Stderr,
				Message: mr.Message,
			}

			minionResults = append(minionResults, r)
		}

		// TODO: 此处的状态，可能是failed
		r := common.Result{
			Jid:           jid,
			Status:        "success",
			MinionResults: minionResults,
		}

		partitionResult.ResultChan <- r
		partitionResult.Close()

		//common.SendResultAndClose(r, partitionResult, true)

	}, func(message interface{}) {
		logger.Error("process panic error", "error", message)
	})
	return jid, nil
}

func conversionStatus(puppetStatus string) (status string) {
	if puppetStatus == define.PuppetStatusOK {
		return define.Success
	}
	return define.Fail
}
