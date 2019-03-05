//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package mco

import (
	"fmt"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/log"
	"idcos.io/cloud-act2/utils/generator"
	"testing"
)

const ip123 = "10.0.0.123"

func TestPuppetMQSend(t *testing.T) {
	err := config.LoadConfig("/usr/yunji/cloud-act2/conf/cloud-act2-proxy.yaml")
	if err != nil {
		t.Errorf("file distribution %s", err)
		return
	}

	if config.Conf.IsMaster() {
		log.InitLogger("master")
	} else {
		log.InitLogger("proxy")
	}

	message := NewMessage("uptime")

	jid := generator.GenUUID()

	mq := NewMcoRabbitMQClient(config.Conf.Puppet.RabbitMQ, jid)
	err = mq.Connect()
	if err != nil {
		t.Errorf("puppet mq %s", err)
		return
	}
	defer mq.Close()
	fmt.Printf("mq connect %#v\n", mq)

	err = mq.CreateReplyQueue()
	if err != nil {
		t.Errorf("puppet reply queue %s", err)
		return
	}
	defer mq.CloseReplyQueue()

	mq.SendMessage(ip123, message)

	done := make(chan bool)

	go func() {

		deliveries, err := mq.Consume()
		if err != nil {
			fmt.Printf("consume message error %s\n", err)
		}

		for delivery := range deliveries {
			fmt.Printf("delivery %#v\n", delivery)

			result, err := GetMcoMessageResult(delivery.Body)
			if err != nil {
				fmt.Printf("get mco message error %s\n", err)
				done <- true
			}
			fmt.Printf("result %#v\n", result)

			done <- true
		}
	}()

	<-done

}
