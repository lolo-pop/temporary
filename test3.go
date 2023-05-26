package main

import (
	"fmt"
	"net"
)

func getContainerIP() string {
	// 获取容器内部的IP地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

func main() {
	fmt.Println(getContainerIP())
}
