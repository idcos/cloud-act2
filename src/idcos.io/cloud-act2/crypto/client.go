//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package crypto

import (
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"sync"
)

type Client interface {
	Encode(plaintext string) string
	Decode(ciphertext string) (string, error)
}

var client Client
var once sync.Once

func GetClient() Client {
	if client == nil {
		once.Do(initClient)
	}
	return client
}

func initClient() {
	switch config.Conf.CryptoType {
	case define.CryptoAEAE:
		fallthrough
	default:
		client, _ = getAeadClient()
	}
}
