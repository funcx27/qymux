package utils

import "net"

func GetLocalIP() string {
	// 使用 UDP 拨号，不会真正发起网络连接，不需要对方存在
	// 8.8.8.8:80 是一个通用地址，用于诱导内核选择路由
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// 处理可能的 IPv6 情况或特殊格式
	ip := localAddr.IP.String()
	if ip == "" {
		return "127.0.0.1"
	}
	return ip
}
