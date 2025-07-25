//  Copyright 2019 The Go Authors. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

package netutil

import (
	"fmt"
	"math/big"
	"net"
	"sync"
)

const IPPattern = `(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)`

var localIPv4Str = "0.0.0.0"
var localIPv4Once = new(sync.Once)

func LocalIPV4() string {
	localIPv4Once.Do(func() {
		if ias, err := net.InterfaceAddrs(); err == nil {
			for _, address := range ias {
				if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						localIPv4Str = ipNet.IP.String()
						return
					}
				}
			}
		}
	})
	return localIPv4Str
}

func GetIPV4(addr net.Addr) string {
	if addr == nil {
		return ""
	}

	if ipNet, ok := addr.(*net.TCPAddr); ok {
		return ipNet.IP.String()
	}

	if ipNet, ok := addr.(*net.UDPAddr); ok {
		return ipNet.IP.String()
	}

	return ""
}

func LocalIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP, nil
			}
		}
	}

	return nil, fmt.Errorf("no local ip")
}

func IP2BigInt(IP net.IP) *big.Int {
	ret := big.NewInt(0)
	ret.SetBytes(IP)
	return ret
}

func GetHostIp() (string, error) {
	conn, err := net.Dial("udp", "114.114.114.114:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
