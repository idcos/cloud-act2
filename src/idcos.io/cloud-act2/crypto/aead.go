//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package crypto

import (
	"crypto/cipher"
	"encoding/base64"
	"golang.org/x/crypto/chacha20poly1305"
	"idcos.io/cloud-act2/config"
	"sync"
)

type aeadClient struct {
	aead  cipher.AEAD
	nonce []byte
	once  sync.Once
}

func (c *aeadClient) Encode(plaintext string) string {
	//乱码字符串放入数据库等系统时会导致byte变化,需要将byte数组base64编码
	bytes := c.aead.Seal(nil, c.nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(bytes)
}

func (c *aeadClient) Decode(ciphertext string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	bytes, err := c.aead.Open(nil, c.nonce, cipherBytes, nil)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func getAeadClient() (*aeadClient, error) {
	logger := getLogger()

	keyStr := config.Conf.CryptoKey
	bytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		logger.Error("read crypto key fail", "error", err)
		return nil, err
	}

	aead, _ := chacha20poly1305.New(bytes)
	return &aeadClient{
		aead:  aead,
		nonce: bytes[:chacha20poly1305.NonceSize],
	}, nil
}
