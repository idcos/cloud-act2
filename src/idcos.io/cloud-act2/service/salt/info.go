//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package salt

import (
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/iputil"
	"strings"

	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/utils/cmd"
	"idcos.io/cloud-act2/utils/system"
)

//Options act2 options
type Options struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//MasterInfo master info
type MasterInfo struct {
	Act2Server   string  `json:"server"`
	Options      Options `json:"options"`
	Status       string  `json:"status"`
	IDC          string  `json:"idc"`
	SerialNumber string  `json:"sn"`
	Type         string  `json:"type"`
}

//MinionInfo minion info
type MinionInfo struct {
	SerialNumber  string   `json:"sn"`
	IPs           []string `json:"ips"`
	Status        string   `json:"status"`
	OsType        string   `json:"os_type"`
	MinionVersion string   `json:"minionVersion"`
}

// 获取salt的状态
func getSaltStatus() (string, error) {
	logger := getLogger()

	pipe := cmd.NewPipe("ps", "-ef")
	pipe.P("grep", "--color=never", "salt-master", "-c")
	// pipe.Pipe("grep", "--color=never", "-v", "grep", "-c")
	buff, err := pipe.Call()
	if err != nil {
		logger.Error("pipe cmd run", "error", err)
		return "", errors.Wrap(err, "salt count error")
	}

	countStr := buff.String()
	countStr = strings.Trim(countStr, "\r\n")
	logger.Debug(fmt.Sprintf("count result %s", countStr))

	count, err := strconv.Atoi(countStr)
	if err != nil {
		logger.Error("ps command could not convert to count", "error", err)
		return "", errors.Wrap(err, "convert count error")
	}
	if count > 1 {
		return "running", nil
	} else {
		return "stopped", nil
	}
}

func getPuppetStatus() (string, error) {
	// puppet master机器是否存活，按照目前的设计，是mco的通道，也就是rabbitmq的服务是否存活
	// TODO: check mcollective_directed的queue是否通更合适
	connection, err := amqp.Dial(config.Conf.Puppet.PuppetDBServer)
	if err != nil {
		return "stopped", nil
	}
	defer connection.Close()
	return "running", nil
}

//GetSaltMasterInfo 获取salt的master信息
func GetSaltMasterInfo(typ string) (*MasterInfo, error) {
	logger := getLogger()

	sn, err := system.GetSystemSN()
	if err != nil {
		logger.Error("get system sn fail", "error", err)
		return nil, err
	}

	status := "stopped"
	if typ == "salt" {
		// 内容还是需要上报给master的
		saltStatus, err := getSaltStatus()
		if err != nil {
			logger.Warn("get salt status fail", "error", err)
		} else {
			status = saltStatus
		}
	} else if typ == "puppet" {
		status, err = getPuppetStatus()
		if err != nil {
			return nil, err
		}
	}

	conf := config.Conf
	return &MasterInfo{
		Act2Server: conf.Act2.ProxyServer,
		Options: Options{
			Username: conf.Salt.Username,
			Password: conf.Salt.Password,
		},
		Status:       status,
		IDC:          conf.IDC,
		SerialNumber: sn,
		Type:         typ,
	}, nil
}

func getIPs(ip4Interface map[string]interface{}) []string {
	logger := getLogger()

	logger.Trace(fmt.Sprintf("ip4 interface %v", ip4Interface))
	var ips []string
	for _, value := range ip4Interface {
		if addrs, ok := value.([]interface{}); ok {
			for _, addr := range addrs {
				ip := addr.(string)
				if !iputil.IsReversedIP(ip) {
					ips = append(ips, ip)
				}
			}
		}
	}
	return ips
}

func buildMinionInfo(minionItem map[string]interface{}) *MinionInfo {
	logger := getLogger()

	uuidName := "idcos_system_id"
	var sn string
	if uuid, ok := minionItem[uuidName]; ok {
		sn = uuid.(string)
	} else {
		logger.Error("could not get system id", "error", dataexchange.ToJsonString(minionItem))
		return nil
	}

	var ips []string
	if ip4Interfaces, ok := minionItem["ip4_interfaces"]; ok {
		if ip4InterfacesValue, ok := ip4Interfaces.(map[string]interface{}); ok {
			ips = getIPs(ip4InterfacesValue)
		}
	} else if ipv4, ok := minionItem["ipv4"]; ok {
		ipv4Strings, ok := ipv4.([]string)
		if ok {
			ips = ipv4Strings
		}
	}

	var kernelStr string
	if kernel, ok := minionItem["kernel"]; ok {
		kernelStr = strings.ToLower(kernel.(string))
	}

	var saltVersionStr string
	if saltVersion, ok := minionItem["saltversion"]; ok {
		saltVersionStr = strings.ToLower(saltVersion.(string))
	}

	minionInfo := MinionInfo{
		SerialNumber:  strings.ToUpper(sn),
		Status:        "running",
		IPs:           ips,
		OsType:        kernelStr,
		MinionVersion: saltVersionStr,
	}

	return &minionInfo

}

//GetMinionInfo get minion info
func GetMinionInfo() ([]*MinionInfo, error) {
	logger := getLogger()
	saltPath := config.Conf.Salt.SaltPath
	saltConfigPath := config.Conf.Salt.SaltConfig

	logger.Debug("minion start", "salt config path", saltConfigPath, "salt path", saltPath)

	executor := cmd.NewExecutor(saltPath, "-c", saltConfigPath, "*", "grains.item",
		"idcos_system_id", "ip4_interfaces", "ipv4", "kernel", "saltversion",
		"--output=json", "--static")
	r := executor.Invoke()

	if r.Error != nil {
		logger.Error("run salt grains.items", "error", r.Error)
		return nil, errors.Wrap(r.Error, "run salt grains.items error")
	}

	logger.Trace("salt minion info", "grains", r.Stdout.String())

	decoder := json.NewDecoder(r.Stdout)
	var minions map[string]interface{}
	err := decoder.Decode(&minions)
	if err != nil {
		logger.Error("json unmarshal", "error", err)
		return nil, errors.Wrap(err, "json unmarshal error")
	}

	var minionInfos []*MinionInfo
	for _, value := range minions {
		minionItem, ok := value.(map[string]interface{})
		if !ok {
			logger.Info("minion item not map", "minionItem", minionItem)
			continue
		}

		minionInfo := buildMinionInfo(minionItem)
		if minionInfo != nil {
			minionInfos = append(minionInfos, minionInfo)
		}
	}

	return minionInfos, nil
}
