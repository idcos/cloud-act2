//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package model

import "time"

//Act2WebHook web钩子对象
type Act2WebHook struct {
	ID            string    `gorm:"id;primary_key" json:"id"`
	Event         string    `gorm:"event" json:"event"`
	Uri           string    `gorm:"uri" json:"uri"`
	Token         string    `gorm:"token" json:"token"`
	CustomHeaders string    `gorm:"custom_headers" json:"customHeaders"`
	CustomData    string    `gorm:"custom_data" json:"customData"`
	AddTime       time.Time `gorm:"add_time" json:"addTime"`
	UpdateTime    time.Time `gorm:"update_time" json:"updateTime"`
}

//Act2WebHookRecord web钩子记录
type Act2WebHookRecord struct {
	ID        string    `gorm:"id;primary_key" json:"id"`
	WebHookID string    `gorm:"web_hook_id" json:"webHookId"`
	SendTime  time.Time `gorm:"send_time" json:"sendTime"`
	Event     string    `gorm:"event" json:"event"`
	Request   string    `gorm:"request" json:"request"`
	Response  string    `gorm:"response" json:"response"`
}

//TableName 数据库表名
func (Act2WebHook) TableName() string {
	return "act2_web_hook"
}

//TableName 数据库表名
func (Act2WebHookRecord) TableName() string {
	return "act2_web_hook_record"
}

//Save 保存
func (r *Act2WebHook) Save() error {
	return globalDb.Save(r).Error
}

//Save 保存
func (r *Act2WebHookRecord) Save() error {
	return globalDb.Save(r).Error
}

//FindWebHookByID 查询web钩子
func FindWebHookByID(id string) (webHook Act2WebHook, err error) {
	webHook = Act2WebHook{}
	err = globalDb.Model(&Act2WebHook{}).Where("id = ?", id).Find(&webHook).Error
	return
}

//FindWebHookInfoByEvent
func FindWebHookInfoByEvent(event string) (webHookInfoList []Act2WebHook, err error) {
	webHookInfoList = make([]Act2WebHook, 0, 10)
	err = globalDb.Where("event = ?", event).Find(&webHookInfoList).Error
	return
}
