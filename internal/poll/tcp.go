package poll

import (
	"fmt"
	"net"
	"time"
)

func TCP(hostname string, port string) (time.Duration, error) {
	start := time.Now()

	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", hostname, port))
	if err != nil {
		return time.Since(start), err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return time.Since(start), err
	}
	defer conn.Close()

	return time.Since(start), err
}