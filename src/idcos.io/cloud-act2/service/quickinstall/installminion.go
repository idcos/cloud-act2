//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package quickinstall

import (
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"idcos.io/cloud-act2/crypto"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/service/job"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/generator"
	"idcos.io/cloud-act2/utils/promise"
)

const (
	linuxPath  = "/tmp"
	windowPath = "C:\\"
)

func sendErrorToChan(resp chan common.InstallResp, err error) {
	resp <- common.InstallResp{
		Error: err,
	}
}

func QuickInstall(info common.InstallInfo, resp chan common.InstallResp) {
	logger := getLogger()

	// 对密码进行混淆处理
	if len(info.Password) > 0 {
		client := crypto.GetClient()
		info.Password = client.Encode(info.Password)
	}

	//下发安装包
	recordID, target, err := sendPackage(info)
	if err != nil {
		logger.Error("send package fail", "error", err)
		sendErrorToChan(resp, err)
		return
	}

	//监听任务结束
	done := make(chan model.JobRecord)
	promise.NewGoPromise(func(chan struct{}) {
		waitForRecordDone(recordID, done)
	}, nil)
	record := <-done

	//获取主机的成功、失败列表
	successHosts, failHosts, err := getExecResult(record, info.Hosts)
	if err != nil {
		logger.Error("get host exec result fail", "error", err)
		sendErrorToChan(resp, err)
		return
	}

	//返回安装包下发成功的信息
	resp <- common.InstallResp{
		Status:       "send-package",
		SuccessHosts: execHostToIPList(successHosts),
		FailHosts:    failHosts,
	}

	//如果下发文件成功的主机为空，执行结束
	if len(successHosts) == 0 {
		close(resp)
		return
	}

	//执行安装程序
	recordID, err = runPackage(successHosts, checkIsWindows(info.Hosts[0].OsType), target, info.MasterIP, info.Username, info.Password)
	if err != nil {
		logger.Error("run package fail", "error", err)
		sendErrorToChan(resp, err)
		return
	}

	//监听任务结束
	promise.NewGoPromise(func(chan struct{}) {
		waitForRecordDone(recordID, done)
	}, nil)
	record = <-done

	successHosts, failHosts, err = getExecResult(record, successHosts)
	if err != nil {
		logger.Error("get host exec result fail", "error", err)
		sendErrorToChan(resp, err)
		return
	}

	//返回安装成功的信息
	resp <- common.InstallResp{
		Status:       "install-package",
		SuccessHosts: execHostToIPList(successHosts),
		FailHosts:    failHosts,
	}

	//关闭通道
	close(resp)

	logger.Info("install minion success", "list", dataexchange.ToJsonString(successHosts))
}

func execHostToIPList(execHosts []common.ExecHost) []string {
	ips := make([]string, len(execHosts))
	for i, execHost := range execHosts {
		ips[i] = execHost.HostIP
	}
	return ips
}

func checkIsWindows(osType string) bool {
	if osType == define.Win {
		return true
	}
	return false
}

func runPackage(hosts []common.ExecHost, isWindows bool, targetPath string, masterIP string, username string, password string) (string, error) {
	//script
	script := ""
	scriptType := ""
	if isWindows {
		script = targetPath + " /S /master=" + masterIP
		scriptType = define.BatType
	} else {
		script = "sh " + targetPath + " " + masterIP
		scriptType = define.ShellType
	}

	execParam := common.ConfJobIPExecParam{
		ExecHosts: hosts,
		Provider:  define.MasterTypeSSH,
		ExecuteID: generator.GenUUID(),
		ExecParam: common.ExecParam{
			Pattern:    define.ScriptModule,
			Script:     script,
			ScriptType: scriptType,
			Timeout:    300,
			RunAs:      username,
			Password:   password,
		},
	}

	recordID, err := job.ProcessAndExecByIP("system", execParam)
	if err != nil {
		return "", err
	}

	return recordID, nil
}

func getExecResult(record model.JobRecord, hosts []common.ExecHost) (successHosts []common.ExecHost, failHosts []common.InstallFailHost, err error) {
	logger := getLogger()
	if record.ResultStatus == define.Success {
		return hosts, nil, nil
	}

	if record.ResultStatus == define.Timeout {
		return nil, execHostsToInstalFialHosts(hosts, "timeout"), nil
	}

	hostResults, err := job.FindHostResultsByRecordID(record.ID)
	if err != nil {
		logger.Error("find host result fail", "recordID", record.ID, "error", err)
		return nil, nil, err
	}

	successHosts = make([]common.ExecHost, 0, 10)
	failHosts = make([]common.InstallFailHost, 0, 10)
	for _, hostResult := range hostResults {
		execHost, err := getExecHostBySameIP(hosts, hostResult.HostIP)
		if err != nil {
			continue
		}
		if hostResult.ResultStatus == define.Success {
			successHosts = append(successHosts, execHost)
		} else {
			msg := hostResult.Stderr
			if len(msg) == 0 {
				msg = hostResult.Message
			}
			failHosts = append(failHosts, common.InstallFailHost{
				IP:      hostResult.HostIP,
				Message: msg,
			})
		}
	}

	return successHosts, failHosts, nil
}

func execHostsToInstalFialHosts(execHosts []common.ExecHost, message string) []common.InstallFailHost {
	failHosts := make([]common.InstallFailHost, len(execHosts))
	for i, execHost := range execHosts {
		failHosts[i] = common.InstallFailHost{
			IP:      execHost.HostIP,
			Message: message,
		}
	}
	return failHosts
}

func getExecHostBySameIP(hosts []common.ExecHost, ip string) (common.ExecHost, error) {
	for _, host := range hosts {
		if host.HostIP == ip {
			return host, nil
		}
	}
	return common.ExecHost{}, errors.New("not found host by ip,abnormal data")
}

func waitForRecordDone(recordID string, done chan<- model.JobRecord) {
	logger := getLogger()
	for {
		record := &model.JobRecord{}
		err := record.GetByID(recordID)
		if err != nil {
			logger.Error("query record fail", "id", recordID, "error", err)
		}

		if record.ExecuteStatus == define.Done {
			done <- *record
			break
		}

		time.Sleep(2 * time.Second)
	}
}

func sendPackage(info common.InstallInfo) (string, string, error) {
	//target
	fileSuffix := filepath.Ext(info.Path)
	fileName := "install" + fileSuffix
	target := ""
	if fileSuffix == ".sh" {
		target = linuxPath
	} else {
		target = windowPath
	}

	//urls
	urls := []string{info.Path}

	execParam := common.ConfJobIPExecParam{
		ExecHosts: info.Hosts,
		Provider:  define.MasterTypeSSH,
		ExecuteID: generator.GenUUID(),
		ExecParam: common.ExecParam{
			Pattern: define.FileModule,
			Params: map[string]interface{}{
				"target":   target,
				"fileName": fileName,
			},
			ScriptType: define.UrlType,
			Script:     dataexchange.ToJsonString(urls),
			RunAs:      info.Username,
			Password:   info.Password,
			Timeout:    300,
		},
	}

	recordID, err := job.ProcessAndExecByIP("system", execParam)
	if err != nil {
		return "", "", err
	}
	return recordID, target + "/" + fileName, nil
}
