package public

import (
	"fmt"
	"net"
	"os"
)

var (
	Listener string
	Addr     string
)

func genericInterface(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil && (ip.String() == "127.0.0.1" || ip.String() == "0.0.0.0")
}

func SetListener(addr string) (string, error) {
	Listener = addr

	var err error
	if err == nil && Addr == "" {
		_, err = SetAddr(Listener)
	}

	return Listener, err
}

func SetAddr(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)

	if host == "" || genericInterface(host) {
		host, err = os.Hostname()
	}

	if err == nil {
		Addr = fmt.Sprintf("http://%s:%s", host, port)
	}

	return Addr, err
}
