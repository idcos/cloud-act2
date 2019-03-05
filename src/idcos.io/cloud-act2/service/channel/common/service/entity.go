//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package service

//ChannelConfig 通道配置
type ChannelConfig struct {
	Server   string
	Type     string
	Option   string
}

type saltOption struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}
