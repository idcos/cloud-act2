//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package iputil

import (
	"net"
)

func Rfc1918Private(ip net.IP) bool {
	for _, cidr := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"} {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return false
			// panic("failed to parse hardcoded rfc1918 cidr: " + err.Error())
		}
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func Rfc4193Private(ip net.IP) bool {
	_, subnet, err := net.ParseCIDR("fd00::/8")
	if err != nil {
		return false
		// panic("failed to parse hardcoded rfc4193 cidr: " + err.Error())
	}
	return subnet.Contains(ip)
}

func IsLoopback(ip net.IP) bool {
	for _, cidr := range []string{"127.0.0.0/8", "::1/128"} {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return false
			// panic("failed to parse hardcoded loopback cidr: " + err.Error())
		}
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func MightBePublic(ip net.IP) bool {
	if ip != nil {
		return !Rfc1918Private(ip) && !Rfc4193Private(ip) && !IsLoopback(ip)
	} else {
		return false
	}
}

func IsReversedIP(addr string) bool {
	ip := net.ParseIP(addr)
	if ip != nil {
		return IsLoopback(ip)
	} else {
		return false
	}
}
