//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package model

import (
	"bytes"
	"fmt"
	"time"
)

//TimePeriod 时间段
type TimePeriod struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

//Check 校验时间段数据是否存在并合法
func (t TimePeriod) Check() bool {
	if t.StartTime.IsZero() && t.EndTime.IsZero() {
		return false
	}

	if (!t.StartTime.IsZero() && !t.EndTime.IsZero()) && t.EndTime.Before(t.StartTime) {
		return false
	}
	return true
}

// TimeCondition 生成大于等于开始时间且小于等于结束时间的条件表达式
/*
time.IsZero() == true 时会忽略
*/
func (t TimePeriod) TimeCondition(where *bytes.Buffer, values []interface{}, startColumn, endColumn string) (*bytes.Buffer, []interface{}) {
	if t.Check() {
		if !t.StartTime.IsZero() {
			where.WriteString(fmt.Sprintf(" and %s >= ?", startColumn))
			values = append(values, t.StartTime.Format("2006-01-02 15:04:05"))
		}
		if !t.EndTime.IsZero() {
			where.WriteString(fmt.Sprintf(" and %s <= ?", endColumn))
			values = append(values, t.EndTime.Format("2006-01-02 15:04:05"))
		}
	}
	return where, values
}

//TimeConsumingPeriod 耗时段
type TimeConsumingPeriod struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

//Check 校验耗时段数据是否存在并合法
func (t TimeConsumingPeriod) Check() bool {
	if t.Start == 0 && t.End == 0 {
		return false
	}

	if (t.Start > 0 && t.End > 0) && (t.End < t.Start) {
		return false
	}
	return true
}

//Condition 生成耗时段大于等于开始时间段并小于等于结束时间段的条件表达式
/*
condition <= 0 时会忽略
*/
func (t TimeConsumingPeriod) Condition(where *bytes.Buffer, values []interface{}, column string) (*bytes.Buffer, []interface{}) {
	if t.Check() {
		if t.Start > 0 {
			where.WriteString(fmt.Sprintf(" and %s >= ?", column))
			values = append(values, t.Start)
		}
		if t.End > 0 {
			where.WriteString(fmt.Sprintf(" and %s <= ?", column))
			values = append(values, t.End)
		}
	}
	return where, values
}
