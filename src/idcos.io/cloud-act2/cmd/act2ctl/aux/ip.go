//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package aux

import (
	"errors"
	"fmt"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/utils/fileutil"
	"strings"
)

func DistinctIPs(ips []string) []string {
	if len(ips) == 0 {
		return nil
	}

	// remove duplicate ip
	ipMap := make(map[string]common.NullStruct)
	for _, ip := range ips {
		ipMap[ip] = common.NullStruct{}
	}

	var results []string
	for ip := range ipMap {
		results = append(results, ip)
	}
	return results
}

func ExtractTargetIPs(target string, hostFile string) ([]string, error) {
	if target == "" && hostFile == "" {
		return nil, errors.New(fmt.Sprintf("must set one of ip list or host file"))
	}

	var targetIPs []string
	if target == "" {
		if content, err := fileutil.ReadFile(hostFile); err != nil {
			return nil, err
		} else {
			targetIPs = strings.Split(content, "\n")
			// 移除空格和空的行
			var ips []string
			for _, ip := range targetIPs {
				ip = strings.TrimSpace(ip)
				if ip == "" {
					continue
				}
				ips = append(ips, ip)
			}
			targetIPs = ips
		}
	} else {
		targetIPs = strings.Split(target, ",")
	}

	targetIPs = DistinctIPs(targetIPs)

	if len(targetIPs) == 0 {
		return nil, errors.New("no ip found")
	}

	return targetIPs, nil
}
