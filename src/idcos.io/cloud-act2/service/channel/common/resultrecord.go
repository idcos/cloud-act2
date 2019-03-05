//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

import (
	"idcos.io/cloud-act2/utils/stringutil"
	"sync"
	"time"
)

//用于记录已返回结果记录
var returnedRecordMap = make(map[string]recordInfo, 100)

var mutex sync.RWMutex

type recordInfo struct {
	hosts   []string
	expired time.Time
}

//AddRecord 添加记录
func AddRecord(recordID string, hosts []string, timeout int) {
	mutex.RLock()
	info, ok := returnedRecordMap[recordID]
	mutex.RUnlock()
	if ok {
		info.hosts = append(info.hosts, hosts...)
		return
	}

	duration := time.Duration(timeout+10) * time.Second
	info = recordInfo{
		hosts:   hosts,
		expired: time.Now().Add(duration),
	}

	mutex.Lock()
	returnedRecordMap[recordID] = info
	mutex.Unlock()
}

//FilterHosts 过滤已返回的记录
func FilterHosts(recordID string, hosts []string) (filterHosts []string) {
	info, ok := returnedRecordMap[recordID]
	if !ok {
		return hosts
	}

	filterHosts = stringutil.DifferenceStringSlice(hosts, info.hosts)
	return
}

//RemoveRecord 删除记录
func RemoveRecord(recordID string) {
	mutex.Lock()
	delete(returnedRecordMap, recordID)
	mutex.Unlock()
}
