package common

import (
	"fmt"
	"net"
	"strings"
)

type CollectEntry struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

func GetOutboundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}
