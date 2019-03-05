//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package model

import (
	"bytes"
	"fmt"
	"time"
)

// Act2Idc table struct
type Act2Idc struct {
	ID      string    `gorm:"column:id;primary_key" json:"id"`
	Name    string    `gorm:"column:name" json:"name"`
	AddTime time.Time `gorm:"column:add_time" json:"addTime"`
}

func (Act2Idc) TableName() string {
	return "act2_idc"
}

func (r *Act2Idc) FindIdcByName(name string) (isFound bool) {
	return globalDb.Where("name = ? ", name).Find(&r).RecordNotFound()
}

func FindFirstIDC() (idc Act2Idc, err error) {
	idc = Act2Idc{}
	err = globalDb.Find(&idc).Limit(1).Error
	return idc, err
}

func GetAllIdcMap() (idcMap map[string]Act2Idc, err error) {
	var idcs []Act2Idc
	idcMap = make(map[string]Act2Idc)
	err = globalDb.Find(&idcs).Error
	if err != nil {
		return
	}

	for _, idc := range idcs {
		idcMap[idc.ID] = idc
	}
	return
}

func GetAllIdcNameMap() (idcMap map[string]Act2Idc, err error) {
	var idcs []Act2Idc
	idcMap = make(map[string]Act2Idc)
	err = globalDb.Find(&idcs).Error
	if err != nil {
		return
	}

	for _, idc := range idcs {
		idcMap[idc.Name] = idc
	}
	return
}

// Act2Proxy table struct
type Act2Proxy struct {
	ID        string    `gorm:"column:id;primary_key"`
	TwiceTime time.Time `gorm:"column:twice_time"`
	LastTime  time.Time `gorm:"column:last_time"`
	IdcID     string    `gorm:"column:idc_id"`
	Server    string    `gorm:"column:server"`
	Type      string    `gorm:"column:type"`
	Status    string    `gorm:"column:status"`
	Options   string    `gorm:"column:options"`
}

// TableName convert
func (Act2Proxy) TableName() string {
	return "act2_proxy"
}

//FindProxyByID 更加id获取proxy
func FindProxyByID(id string) (proxy Act2Proxy, err error) {
	proxy = Act2Proxy{}
	err = GetDb().Model(&Act2Proxy{}).Where("id = ?", id).Find(&proxy).Error
	return
}


func FindProxiesByIdcID(idcID string) ([]Act2Proxy, error) {
	var proxies []Act2Proxy
	err := GetDb().Model(&Act2Proxy{}).Where("idc_id = ?", idcID).Find(&proxies).Error
	return proxies, err
}


func GetAllProxyMap() (proxyMap map[string]Act2Proxy, err error) {
	var proxyes []Act2Proxy

	proxyMap = make(map[string]Act2Proxy)

	err = globalDb.Find(&proxyes).Error
	if err != nil {
		return
	}

	for _, proxy := range proxyes {
		proxyMap[proxy.IdcID] = proxy
	}
	return
}

type proxyIDIDC struct {
	ProxyID string `grom:"column:proxy_id"`
	Act2Idc
}

//GetProxyIDIDCMapByProxyIDs 更加proxyIDs获取proxyID和idc的对应map
func GetProxyIDIDCMapByProxyIDs(proxyIDs []string) (proxyMap map[string]Act2Idc, err error) {
	proxyIDCs := make([]proxyIDIDC, 0, 10)

	err = globalDb.Table("act2_proxy p").
		Select("p.id as proxy_id,i.*").
		Joins("LEFT JOIN act2_idc i ON p.idc_id = i.id").
		Where("p.id in (?)", proxyIDs).Find(&proxyIDCs).Error
	if err != nil {
		return nil, err
	}

	proxyMap = make(map[string]Act2Idc)
	for _, proxyIDC := range proxyIDCs {
		proxyMap[proxyIDC.ProxyID] = proxyIDC.Act2Idc
	}

	return
}

// Save Act2Proxy
func (r *Act2Proxy) Save() error {
	return globalDb.Save(r).Error
}

func (r *Act2Proxy) FindByID(ID string) bool {
	return globalDb.Where("id = ?", ID).Find(r).RecordNotFound()
}

type Act2Host struct {
	ID            string    `gorm:"column:id;primary_key"`
	IdcID         string    `gorm:"column:idc_id"`
	EntityID      string    `gorm:"column:entity_id"`
	ProxyID       string    `gorm:"column:proxy_id"`
	AddTime       time.Time `gorm:"column:add_time"`
	OsType        string    `gorm:"column:os_type"`
	Status        string    `gorm:"column:status"`
	MinionVersion string    `gorm:"column:minion_version"`
}

func GetAct2HostColumns() []string {
	return []string{"id", "idc_id", "proxy_id", "entity_id", "add_time", "os_type", "status", "minion_version"}
}

func FindHostsByEntityIds(systemIds []string) (hosts []Act2Host, err error) {
	err = globalDb.Model(&Act2Host{}).Where("entity_id in (?)", systemIds).Find(&hosts).Error
	return
}

func (r *Act2Host) FindByHostID(hostID string) (notFound bool, err error) {
	db := globalDb.Where("id = ? ", hostID).Find(&r)
	return db.RecordNotFound(), db.Error
}

//FindHostByIDs 根据id列表获取主机列表
func FindHostByIDs(hostIDs []string) (hosts []Act2Host, err error) {
	hosts = make([]Act2Host, 0, 10)
	err = globalDb.Model(&Act2Host{}).Where("id in (?)", hostIDs).Find(&hosts).Error
	return
}

/**
create or update
*/
func (r *Act2Host) Save() error {
	return globalDb.Save(r).Error
}

func (Act2Host) TableName() string {
	return "act2_host"
}

type Act2HostIP struct {
	ID      string    `gorm:"column:id;primary_key" json:"id"`
	HostID  string    `gorm:"column:host_id" json:"hostId"`
	IP      string    `gorm:"column:ip" json:"ip"`
	AddTime time.Time `gorm:"column:add_time" json:"addTime"`
}

func GetAct2HostIPColumns() []string {
	return []string{"id", "host_id", "ip", "add_time"}
}

func (Act2HostIP) TableName() string {
	return "act2_host_ip"
}

/**
create or update
*/
func (r *Act2HostIP) Save() error {
	return globalDb.Save(r).Error
}

func FindByHostID(hostID string, ips *[]Act2HostIP) {
	globalDb.Model(&Act2HostIP{}).Where("host_id = ? ", hostID).Find(&ips)
}

//FindHostsByHostIDs 根据主机id列表获取主机列表
func FindHostsByHostIDs(hostIDs []string) (hosts []Act2Host, err error) {
	hosts = make([]Act2Host, 0, 5)
	if len(hostIDs) == 0 {
		return
	}
	err = globalDb.Model(&Act2Host{}).Where("id in (?)", hostIDs).Find(&hosts).Error
	return
}

func FindHostIPByIps(ips []string) (hostIps []Act2HostIP, err error) {
	hostIps = make([]Act2HostIP, 0, 10)
	if len(ips) == 0 {
		return
	}
	err = globalDb.Model(&Act2HostIP{}).Where("ip in (?)", ips).Find(&hostIps).Error
	return
}

//UpdateHostProxy 更新主机的proxy
func UpdateHostProxyByEntityID(entityID string, proxyID string) (err error) {
	return globalDb.Model(&Act2Host{}).Where("id = ?", entityID).Update("proxy_id", proxyID).Error
}

//HostInfo 主机信息
type HostInfo struct {
	EntityID    string `gorm:"column:entity_id" json:"entityId"`
	IP          string `gorm:"column:ip" json:"ip"`
	OsType      string `gorm:"column:os_type" json:"osType"`
	Status      string `gorm:"column:host_status" json:"status"`
	IDC         string `gorm:"column:idc_name" json:"idc"`
	ProxyServer string `gorm:"column:proxy_server" json:"proxyServer"`
	ProxyStatus string `gorm:"column:proxy_status" json:"proxyStatus"`
	ProxyType   string `gorm:"column:proxy_type" json:"proxyType"`
}

//FindAllHostInfo 查询所有的主机信息
func FindAllHostInfo(entityID, idc, proxyID, ip string) (hostInfoList []HostInfo, err error) {
	hostInfoList = make([]HostInfo, 0, 100)
	db := GetDb().Table("act2_host host").
		Select("host.entity_id as entity_id,host.os_type as os_type,host.status as host_status,ip.ip as ip,idc.name as idc_name,proxy.server as proxy_server," +
			"proxy.status as proxy_status,proxy.type as proxy_type").
		Joins("LEFT JOIN act2_host_ip ip ON host.id = ip.host_id").
		Joins("LEFT JOIN act2_idc idc on host.idc_id = idc.id").
		Joins("LEFT JOIN act2_proxy proxy on host.proxy_id = proxy.id")

	var query bytes.Buffer
	args := make([]interface{}, 0, 4)
	query.WriteString("1=1 ")
	if len(entityID) > 0 {
		queryLike(&query, &args, "host.entity_id", entityID)
	}
	if len(idc) > 0 {
		queryLike(&query, &args, "idc.name", idc)
	}
	if len(proxyID) > 0 {
		queryLike(&query, &args, "proxy.id", proxyID)
	}
	if len(ip) > 0 {
		queryLike(&query, &args, "ip.ip", ip)
	}
	if query.Len() > 0 {
		db = db.Where(query.String(), args...)
	}
	err = db.Find(&hostInfoList).Error
	return
}

func queryLike(query *bytes.Buffer, args *[]interface{}, key string, value string) {
	query.WriteString("and ")
	query.WriteString(key)
	query.WriteString(" like ? ")
	*args = append(*args, fmt.Sprintf("%%%s%%", value))
}