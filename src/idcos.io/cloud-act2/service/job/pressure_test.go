//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package job

import (
	"encoding/json"
	"fmt"
	slave "github.com/dgrr/GoSlaves"
	"github.com/theplant/batchputs"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/utils/generator"
	"sync"
	"testing"
	"time"
)

func TestAddData(t *testing.T) {
	//测试数据的生成
	//初始化数据库
	loadConfig()

	////添加主机数据
	//addHostData()
	//t.Log("add host data finish")
	//hostInfos = hostInfos[:100]
	//addRecordData()
	//t.Log("add record and task finish")
	addHostResultData()
	//t.Log("add host result finish")
}

type hostInfo struct {
	Id      string `gorm:"id"`
	Ip      string `gorm:"ip"`
	ProxyID string `gorm:"proxy_id"`
}

func addHostData() () {

	//2个idc
	idcIDs := []string{generator.GenUUID(), generator.GenUUID()}
	proxyIDs := []string{generator.GenUUID(), generator.GenUUID()}
	idc := model.Act2Idc{
		ID:      idcIDs[0],
		Name:    "上海",
		AddTime: time.Now(),
	}
	err := model.GetDb().Save(&idc).Error
	if err != nil {
		panic(err)
	}

	proxy := model.Act2Proxy{
		ID:        proxyIDs[0],
		LastTime:  time.Now(),
		TwiceTime: time.Now(),
		IdcID:     idcIDs[0],
		Server:    "http://192.168.1.17:5555",
		Type:      define.MasterTypeSalt,
		Status:    define.Running,
		Options:   `{"username":"salt-api","password":"******"}`,
	}

	err = model.GetDb().Save(&proxy).Error
	if err != nil {
		panic(err)
	}

	idc = model.Act2Idc{
		ID:      idcIDs[1],
		Name:    "北京",
		AddTime: time.Now(),
	}
	err = model.GetDb().Save(&idc).Error
	if err != nil {
		panic(err)
	}

	proxy = model.Act2Proxy{
		ID:        proxyIDs[1],
		LastTime:  time.Now(),
		TwiceTime: time.Now(),
		IdcID:     idcIDs[1],
		Server:    "http://192.168.1.17:5555",
		Type:      define.MasterTypeSalt,
		Status:    define.Running,
		Options:   `{"username":"salt-api","password":"******"}`,
	}

	err = model.GetDb().Save(&proxy).Error
	if err != nil {
		panic(err)
	}

	hosts := make([][]interface{}, 10000)
	hostIPs := make([][]interface{}, 10000)
	idcIndex := 0
	for i := 0; i < 10000; i++ {
		hosts[i] = []interface{}{
			generator.GenUUID(),
			idcIDs[idcIndex],
			generator.GenUUID(),
			time.Now(),
			"running",
			"linux",
			proxyIDs[idcIndex],
		}

		hostIPs[i] = []interface{}{
			generator.GenUUID(),
			hosts[i][0],
			nextIP(),
			time.Now(),
		}

		idcIndex++
		if idcIndex > 1 {
			idcIndex = 0
		}
	}
	hostColumns := []string{"id", "idc_id", "entity_id", "add_time", "status", "os_type", "proxy_id"}
	err = batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.Act2Host{}.TableName(), "id", hostColumns, hosts)
	if err != nil {
		panic(err)
	}

	hostIPColumns := []string{"id", "host_id", "ip", "add_time"}
	err = batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.Act2HostIP{}.TableName(), "id", hostIPColumns, hostIPs)
	if err != nil {
		panic(err)
	}
}

var (
	ip1 = 10
	ip2 = 0
	ip3 = 1
	ip4 = 1
)

func nextIP() string {
	ip := fmt.Sprintf("%d.%d.%d.%d", ip1, ip2, ip3, ip4)
	ip4 ++
	if ip4 > 254 {
		ip4 = 1
		ip3 ++
		if ip3 > 254 {
			ip3 = 1
			ip2 ++
		}
	}
	return ip
}

func addRecordData() {
	//测试数据的生成
	//初始化数据库
	loadConfig()

	hostInfos, err := getHostInfos()
	if err != nil {
		panic(err)
	}
	hostIDs := make([]string, len(hostInfos))
	for i, hi := range hostInfos {
		hostIDs[i] = hi.Id
	}
	hostIDsJSON, err := json.Marshal(hostIDs)
	if err != nil {
		panic(err)
	}

	var wait sync.WaitGroup

	for i := 0; i < 13; i ++ {
		wait.Add(1)
		go add1wRecord(string(hostIDsJSON), wait)
	}

	wait.Wait()
}

func getHostInfos() ([]hostInfo, error) {
	hostInfos := []hostInfo{}
	err := model.GetDb().Table("act2_host as host").
		Select("host.id as id,hostIp.id as ip,host.proxy_id as proxy_id").
		Joins("left join act2_host_ip hostIp on hostIp.host_id = host.id").
		Limit(100).Find(&hostInfos).Error
	return hostInfos, err
}

var doingNum = 100

func add1wRecord(hostIDsJSON string, wait sync.WaitGroup) {
	defer wait.Add(-1)
	records := make([][]interface{}, 10000)
	tasks := make([][]interface{}, 10000)

	for i := 0; i < len(records); i ++ {
		executeStatus := define.Done
		if doingNum > 0 {
			executeStatus = define.Doing
		}

		resultStatus := define.Success
		if executeStatus == define.Doing {
			resultStatus = "test"
		}
		if doingNum > 0 {
			doingNum--
		}

		recordID := generator.GenUUID()
		records[i] = []interface{}{
			recordID,
			time.Now(),
			time.Now(),
			executeStatus,
			resultStatus,
			"test",
			define.MasterTypeSalt,
			define.ScriptModule,
			"df -h",
			define.BashType,
			300,
			`{"pattern":"script","script":"ip r","scriptType":"bash","params":{"args":""},"runas":"root","password":"","timeout":300,"env":null,"extendData":"","realTimeOutput":true}`,
			hostIDsJSON,
			"3BB5D781-F4D9-4B9F-9ADD-318425401683",
		}

		tasks[i] = []interface{}{
			generator.GenUUID(),
			time.Now(),
			time.Now(),
			executeStatus,
			resultStatus,
			recordID,
			define.ScriptModule,
			"df -h",
			`{"args":""}`,
			`{"password":"zaq1@WSX","scriptType":"bash","username":"admin"}`,
		}
	}

	for i := 0; i < len(records); i = i + 10 {
		recordColumns := []string{"id", "start_time", "end_time", "execute_status", "result_status", "callback", "provider", "pattern", "script", "script_type", "timeout", "parameters", "hosts", "master_id"}
		err := batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.JobRecord{}.TableName(), "id", recordColumns, records[i:i+10])
		if err != nil {
			panic(err)
		}

		taskColumns := []string{"id", "start_time", "end_time", "execute_status", "result_status", "record_id", "pattern", "script", "params", "options"}
		err = batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.JobTask{}.TableName(), "id", taskColumns, tasks[i:i+10])
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("done")
}

var hostResultColumns = []string{"id", "task_id", "host_id", "proxy_id", "start_time", "end_time", "execute_status", "result_status", "host_ip", "stdout", "stderr", "message"}

func addHostResultData() {
	hostInfos, err := getHostInfos()
	if err != nil {
		panic(err)
	}
	taskIDs := make([]string, 0, 1000000)
	err = model.GetDb().Model(&model.JobTask{}).Pluck("id", &taskIDs).Error
	if err != nil {
		panic(err)
	}

	doingNum = 100

	//初始化线程池
	pool := &slave.SlavePool{
		Work: func(obj interface{}) {
			result := obj.(hostResultData)
			addSingleTaskHostResult(result)
		},
	}
	pool.Open()

	for _, taskID := range taskIDs {
		pool.Serve(hostResultData{
			taskID:    taskID,
			hostInfos: hostInfos,
		})
	}

	for {
		time.Sleep(10 * time.Nanosecond)
	}
}

type hostResultData struct {
	hostInfos []hostInfo
	taskID    string
}

func addSingleTaskHostResult(data hostResultData) {
	hostInfos := data.hostInfos
	taskID := data.taskID
	executeStatus := define.Done
	if doingNum > 0 {
		executeStatus = define.Doing
	}

	resultStatus := define.Success
	if executeStatus == define.Doing {
		resultStatus = "test"
	}
	if doingNum > 0 {
		doingNum--
	}
	hostResults := make([][]interface{}, len(hostInfos))
	for i, hostInfo := range hostInfos {
		hostResults[i] = []interface{}{
			generator.GenUUID(),
			taskID,
			hostInfo.Id,
			hostInfo.ProxyID,
			time.Now(),
			time.Now(),
			executeStatus,
			resultStatus,
			hostInfo.Ip,
			`default via 172.18.0.1 dev eth1 
10.0.1.0/24 dev eth0 proto kernel scope link src 10.0.1.4 
172.18.0.0/16 dev eth1 proto kernel scope link src 172.18.0.7`,
			"",
			"",
		}
	}

	for i := 0; i < len(hostResults); i = i + 10 {
		err := batchputs.Put(model.GetDb().DB(), define.MysqlDriver, model.HostResult{}.TableName(), "id", hostResultColumns, hostResults[i:i+10])
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFindRecordResultsById(b *testing.B) {
	loadConfig()

	b.StartTimer()
	_, err := FindRecordResultsById("000021df-f964-35bc-d347-9db51691672a")
	if err != nil {
		b.Fatal(err)
	}

	b.StopTimer()
}

func BenchmarkFindRecordResultsByPage(b *testing.B) {
	loadConfig()

	b.StartTimer()
	_, err := FindRecordResultsByPage(1, 10)
	if err != nil {
		b.Fatal(err)
	}

	b.StopTimer()
}

func BenchmarkFindRecordResultsCount(b *testing.B) {
	loadConfig()

	b.StartTimer()
	_, err := FindRecordResultsCount()
	if err != nil {
		b.Fatal(err)
	}

	b.StopTimer()
}

func BenchmarkFindHostsByIPs(b *testing.B) {
	loadConfig()

	ips := make([]string, 0, 1000)
	err := model.GetDb().Model(&model.Act2HostIP{}).Limit(1000).Pluck("ip", &ips).Error
	if err != nil {
		b.Fatal(b)
	}

	b.StartTimer()
	hosts, err := model.FindHostIPByIps(ips)
	b.Log(len(hosts))
	b.StopTimer()
}

func BenchmarkFindTasksByRecordID(b *testing.B) {
	loadConfig()

	b.StartTimer()
	_, err := findTasksByRecordID("000021df-f964-35bc-d347-9db51691672a")
	if err != nil {
		panic(err)
	}
	b.StopTimer()
}

func BenchmarkFindUnDoneTasksByRecordID(b *testing.B) {
	loadConfig()

	b.StartTimer()
	_, err := findUnDoneTasksByRecordID("00000039-55a0-6b27-a652-345b4070c2f8")
	if err != nil {
		panic(err)
	}
	b.StopTimer()
}

func BenchmarkFindHostResultsByTaskId(b *testing.B) {
	loadConfig()

	b.StartTimer()
	results, err := findHostResultsByTaskId("ccc08d3c-0ca1-e656-c86f-af8c86bd70dc")
	if err != nil {
		panic(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindHostResultsByRecordID(b *testing.B) {
	loadConfig()

	b.StartTimer()
	results, err := FindHostResultsByRecordID("000021df-f964-35bc-d347-9db51691672a")
	if err != nil {
		panic(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindHostResultsByHostIds(b *testing.B) {
	loadConfig()

	ids := make([]string, 0, 1000)
	err := model.GetDb().Model(&model.Act2Host{}).Limit(1000).Pluck("id", &ids).Error
	if err != nil {
		panic(err)
	}

	b.StartTimer()
	results, err := FindHostResultsByHostIds("ccc08d3c-0ca1-e656-c86f-af8c86bd70dc", ids)
	if err != nil {
		panic(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindHostResultByRecordIDAndEntityID(b *testing.B) {
	loadConfig()

	b.StartTimer()
	_, err := FindHostResultByRecordIDAndEntityID("b2b4285f-de83-5094-67e8-e5e6bfa59610", "2471a93f-239f-3094-135f-4c720be3ad10")
	if err != nil {
		panic(err)
	}
	b.StopTimer()
}

func BenchmarkFindHostInfoByIDs(b *testing.B) {
	loadConfig()

	ids := make([]string, 0, 10000)
	err := model.GetDb().Model(&model.Act2Host{}).Limit(10000).Pluck("id", &ids).Error
	if err != nil {
		panic(err)
	}

	b.StartTimer()
	results, err := findHostInfoByIDs(ids, "salt")
	if err != nil {
		panic(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindHostInfoByHostIDs(b *testing.B) {
	loadConfig()

	ids := make([]string, 0, 10000)
	err := model.GetDb().Model(&model.Act2Host{}).Limit(10000).Pluck("id", &ids).Error
	if err != nil {
		panic(err)
	}

	b.StartTimer()
	results, err := findHostInfoByHostIDs(ids)
	if err != nil {
		panic(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindHostInfoByIDCName(b *testing.B) {
	loadConfig()

	b.StartTimer()
	results, err := FindHostInfoByIDCName("北京")
	if err != nil {
		panic(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindAllIDCHostInfo(b *testing.B) {
	loadConfig()

	b.StartTimer()
	results, err := FindAllIDCHostInfo()
	if err != nil {
		panic(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindHostInfoByIpAndIdc(b *testing.B) {
	loadConfig()

	ips := make([]string, 0, 10000)
	err := model.GetDb().Model(&model.Act2HostIP{}).Limit(10000).Pluck("ip", &ips).Error
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()
	results, err := findHostInfoByIpAndIdc(ips, []string{"北京", "上海"})
	if err != nil {
		b.Fatal(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindJobRecordByIDs(b *testing.B) {
	loadConfig()

	ids := make([]string, 0, 10000)
	err := model.GetDb().Model(&model.JobRecord{}).Limit(10000).Pluck("id", &ids).Error
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()
	results, err := model.FindJobRecordByIDs(ids)
	if err != nil {
		b.Fatal(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindAllMasterIDWithJobRecord(b *testing.B) {
	loadConfig()

	b.StartTimer()
	results, err := model.FindAllMasterIDWithJobRecord()
	if err != nil {
		panic(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindJobTaskByExecuteStatusWithLastDay(b *testing.B) {
	loadConfig()

	b.StartTimer()
	results, err := model.FindJobTaskByExecuteStatusWithLastDay(define.Done)
	if err != nil {
		b.Fatal(err)
	}
	b.StopTimer()

	b.Log(len(results))
}

func BenchmarkFindTimeoutJobs(b *testing.B) {
	loadConfig()

	b.StartTimer()
	results, err := FindTimeoutJobs(300)
	if err != nil {
		b.Fatal(err)
	}
	b.StopTimer()

	b.Log(len(results))
}
