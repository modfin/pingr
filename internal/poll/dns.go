package poll

import (
	"context"
	"errors"
	"fmt"
	"net"
	"pingr/internal/platform/dns"
	"strings"
	"time"
)

/*
	Check the CNAME matches a certain name
		LookupCNAME
	Check that the Addr matches one of a one/many address
		LookupHost or LookupIpAddr??
	Check that the TXT contains a particular value
		LookupTXT, Many records for the same name? What should happen?

	DNSServer addr port always 53???
		Should the network always be udp?

	In TXT:
		Should all "TXT's" contain the value?
*/
const (
	_port = "53"
)

func DNS(domain string, timeout time.Duration, ipAddrCheck string, cnameCheck string, txtCheck []string) (time.Duration, error) {
	start := time.Now()

	DNSServers := dns.Get()

	for _, DNSAddr := range DNSServers {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: timeout,
				}
				return d.DialContext(ctx, "udp", fmt.Sprintf("%s:%s", DNSAddr, _port))
			},
		}

		// Check that ips match
		if ipAddrCheck != "" {
			ips, err := r.LookupHost(context.Background(), domain)
			if err != nil {
				return 0, err
			}
			ipMatch := false
			for _, ip := range ips {
				if ipAddrCheck == ip {
					ipMatch = true
				}
			}
			if !ipMatch {
				return 0, errors.New(fmt.Sprintf("Expected: %s but got: %s", ipAddrCheck, ips))
			}
		}

		// Check that CNAME match
		if cnameCheck != "" {
			cname, err := r.LookupCNAME(context.Background(), domain)
			if err != nil {
				return 0, err
			}
			if cname != cnameCheck {
				return 0, errors.New(fmt.Sprintf("Expected: %s but got: %s", cnameCheck, cname))
			}
		}

		// Check that the TXT contains values
		if txtCheck != nil {
			txts, err := r.LookupTXT(context.Background(), domain)
			if err != nil {
				return 0, err
			}
			for _, txtReqValue := range txtCheck {
				check := false
				for _, txt := range txts {
					if strings.Contains(txt, txtReqValue) {
						check = true
					}
				}
				if !check {
					return 0, errors.New(fmt.Sprintf("Expected TXT to contain: %s got: %s", txtReqValue, txts))
				}
			}
		}
	}

	return time.Since(start), nil
}