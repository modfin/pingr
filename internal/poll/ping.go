package poll

import (
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
)

func Ping(hostname string, timeOut time.Duration) (responseTime time.Duration, err error) {
	p := fastping.NewPinger()
	p.MaxRTT = 10*time.Second

	ra, err := net.ResolveIPAddr("ip4", hostname)
	if err != nil {
		return
	}

	p.AddIPAddr(ra)

	p.MaxRTT = timeOut

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		responseTime = rtt
	}

	err = p.Run()

	return
}