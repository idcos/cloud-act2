//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"sync"
)

var jobInfoMap = map[string]Info{}
var jobMutex sync.RWMutex

//Info 作业信息
type Info struct {
	TaskID string
}

//GetInfoByJid 根据jid获取信息
func GetInfoByJid(jid string) (info Info, ok bool) {
	jobMutex.RLock()
	defer jobMutex.RUnlock()
	info, ok = jobInfoMap[jid]
	return
}

//AddJobInfo 新增数据
func AddJobInfo(jid string, info Info) {
	jobMutex.Lock()
	defer jobMutex.Unlock()
	jobInfoMap[jid] = info
}

//RemoveJobInfo 删除数据
func RemoveJobInfo(key string) {
	jobMutex.Lock()
	defer jobMutex.Unlock()
	delete(jobInfoMap, key)
}
