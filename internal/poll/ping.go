package poll

import (
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
)

func Ping(hostname string, timeOut time.Duration) (responseTime time.Duration, err error) {
	p := fastping.NewPinger()

	ra, err := net.ResolveIPAddr("ip4", hostname)
	if err != nil {
		return
	}

	p.AddIPAddr(ra)

	p.MaxRTT = timeOut

	start:=time.Now()
	err = p.Run()
	responseTime = time.Since(start)

	return
}