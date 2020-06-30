package poll

import (
	"fmt"
	"net"
	"time"
)

func TCP(hostname string, port string) (rt time.Duration, err error) {
	start := time.Now()

	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", hostname, port))
	if err != nil {
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	rt = time.Since(start)
	return
}