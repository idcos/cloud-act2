//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"time"

	"idcos.io/cloud-act2/model"
)

type NullStruct struct {
}

// Register 上报数据传输对象
type RegParam struct {
	Master Master   `json:"master"`
	Minion []Minion `json:"minions"`
}
type Master struct {
	Sn      string  `json:"sn"`
	Server  string  `json:"server"`
	Status  string  `json:"status"`
	Idc     string  `json:"idc"`
	Type    string  `json:"type"`
	Options Options `json:"options"`
}

type Options struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Minion struct {
	Sn            string   `json:"sn"`
	IPs           []string `json:"ips"`
	Status        string   `json:"status"`
	OsType        string   `json:"os_type"`
	MinionVersion string   `json:"minionVersion"`
}

type HostInfo struct {
	HostID        string `gorm:"column:hostId" json:"hostId"`
	HostIP        string `gorm:"column:hostIp" json:"hostIp"`
	IdcName       string `gorm:"column:idcName" json:"idcName"`
	IdcID         string `gorm:"column:idcId" json:"idcId"`
	EntityID      string `gorm:"column:entityId" json:"entityId"`
	OsType        string `gorm:"column:osType" json:"osType"`
	ProxyID       string `gorm:"column:proxyId" json:"proxyId"`
	MinionVersion string `gorm:"column:minionVersion" json:"minionVersion"`
}

type ProxyInfo struct {
	ID      string `gorm:"column:id" json:"id"`
	Server  string `gorm:"column:server" json:"server"`
	Type    string `gorm:"column:type" json:"type"`
	IdcName string `gorm:"column:idcName" json:"idcName"`
	IdcID   string `gorm:"column:idcId" json:"idcId"`
	Status  string `gorm:"column:status" json:"status"`
}

type HostIdcStatInfo struct {
	IdcID string `gorm:"column:idcId"`
	Count int    `gorm:"column:count"`
}

type HostProxyStatInfo struct {
	ProxyID string `gorm:"column:proxyId"`
	Count   int    `gorm:"column:count"`
}

//GroupInfo 组信息
type GroupInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

//UpdateNameInfo 修改名称的信息
type UpdateNameInfo struct {
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

//AssignIPHostGroup ip绑定主机组对象
type IPsHostGroup struct {
	IPs       []string `json:"ips"`
	GroupName string   `json:"groupName"`
}

//CmdCommandGroup 命令命令组对象
type CmdCommandGroup struct {
	Cmds      []string `json:"cmds"`
	GroupName string   `json:"groupName"`
}

//HostGroupCommandGroup 主机组命令组对象
type HostGroupCommandGroup struct {
	HostGroupNames   []string `json:"hostGroupNames"`
	CommandGroupName string   `json:"commandGroupName"`
}

//GetLastTime 获取上一次执行时间
func GetLastTime(master Master) time.Time {
	var exitMaster model.Act2Proxy
	twiceTime := time.Now()

	exitMaster.FindByID(master.Sn)

	if exitMaster.ID != "" {
		twiceTime = exitMaster.LastTime
	}
	return twiceTime
}

//HostProxyChangeInfo 主机的proxy变更信息
type HostProxyChangeInfo struct {
	ProxyID  string `json:"proxyId"`
	EntityID string `json:"entityId"`
}

//HostInfoCondition 主机信息的查询条件
type HostInfoCondition struct {
	IDC      string `json:"idc"`
	EntityID string `json:"entity_id"`
	IP       string `json:"ip"`
	ProxyID  string `json:"proxyId"`
}
