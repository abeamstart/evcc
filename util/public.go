package util

import (
	"fmt"
	"net"
	"os"
)

func PublicAddr(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)

	// tcpAddr, err := net.ResolveTCPAddr("tcp", addr)

	// is host an ip address?
	ip := net.ParseIP(host)
	if err == nil && ip != nil &&
		(ip.String() == "127.0.0.1" || ip.String() == "0.0.0.0") {
		host, err = os.Hostname()
	}

	return fmt.Sprintf("http://%s:%s\n", host, port), err
}
