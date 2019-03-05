//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package model

import (
	"bytes"
	"time"

	"github.com/jinzhu/gorm"
)

//Act2APILog api日志
type Act2APILog struct {
	ID            string    `gorm:"column:id;primary_key" json:"id"`
	URL           string    `gorm:"column:url" json:"url"`
	Description   string    `gorm:"column:description" json:"description"`
	Type          string    `gorm:"column:type" json:"type"`
	User          string    `gorm:"column:user" json:"user"`
	Addr          string    `gorm:"column:addr" json:"addr"`
	OperateTime   time.Time `gorm:"column:operate_time" json:"operateTime"`
	TimeConsuming int       `gorm:"column:time_consuming" json:"timeConsuming"`
	RequestParams string    `gorm:"request_params" json:"requestParams"`
	ResponseBody  string    `gorm:"response_body" json:"responseBody"`
}

func (Act2APILog) TableName() string {
	return "act2_api_log"
}

//Save save api log
func (r *Act2APILog) Save() error {
	return GetDb().Save(r).Error
}

//FindApiLogByID 根据id获取api log
func FindApiLogByID(id string) (apiLog Act2APILog, err error) {
	apiLog = Act2APILog{}
	err = GetDb().Where("id = ?", id).Find(&apiLog).Error
	return
}

//ApiLogOption api log的查询选项
type ApiLogOption struct {
	URL                 string              `json:"url"`
	Type                string              `json:"type"`
	User                string              `json:"user"`
	Addr                string              `json:"addr"`
	TimePeriod          TimePeriod          `json:"timePeriod"`
	TimeConsumingPeriod TimeConsumingPeriod `json:"timeConsumingPeriod"`
}

//FindApiLogByPage 获取api log分页数据
func FindApiLogByPage(pageNo int, pageSize int, option ApiLogOption) (apiLogs []Act2APILog, err error) {
	apiLogs = make([]Act2APILog, 0, 100)
	err = packageApiLogOption(GetDb(), option).Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&apiLogs).Error
	return
}

func packageApiLogOption(db *gorm.DB, option ApiLogOption) *gorm.DB {
	where := bytes.NewBufferString("1=1")
	values := make([]interface{}, 0, 6)
	if len(option.URL) > 0 {
		where.WriteString(" and url like ?")
		values = append(values, "%"+option.URL+"%")
	}
	if len(option.Type) > 0 {
		where.WriteString(" and type = ?")
		values = append(values, option.Type)
	}
	if len(option.User) > 0 {
		where.WriteString(" and user like ?")
		values = append(values, "%"+option.User+"%")
	}
	if len(option.Addr) > 0 {
		where.WriteString(" and addr like ?")
		values = append(values, "%"+option.Addr+"%")
	}
	where, values = option.TimePeriod.TimeCondition(where, values, "start_time", "end_time")
	where, values = option.TimeConsumingPeriod.Condition(where, values, "time_consuming")
	if where.Len() > 0 {
		return db.Where(where.String(), values...)
	}
	return db
}
