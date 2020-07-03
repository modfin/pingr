package poll

import (
	"fmt"
	"net"
	"time"
)

func TCP(hostname string, port string, timeOut time.Duration) (time.Duration, error) {
	start := time.Now()

	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", hostname, port))
	if err != nil {
		return time.Since(start), err
	}

	dialer := net.Dialer{Timeout: timeOut}
	conn, err := dialer.Dial("tcp", tcpAddr.String())
	if err != nil {
		return time.Since(start), err
	}
	defer conn.Close()

	return time.Since(start), err
}