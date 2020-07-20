package poll

import (
	"context"
	"errors"
	"fmt"
	"net"
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

type Record string

const (
	A     Record = "A"
	CNAME Record = "CNAME"
	TXT   Record = "TXT"
	MX    Record = "MX"
	NS    Record = "NS"
)

type Strategy string

const (

	// check = a, b, c
	// res = a,b,c,d
	// -> true
	// check = a, b, c, e
	// res = a, b, c, d
	// -> false
	CheckIsSubset Strategy = "check_is_subset"

	// check = a,b,c,d
	// res = a, b, c
	// -> true
	// check = a, b, c, d
	// res = a, b, c, e
	// -> false
	DNSIsSubset Strategy = "dns_is_subset"

	// check = a, b, c
	// res = a,d,e,f
	// -> true
	// check = a, b, c
	// res = d,e,f
	// -> false
	Intersects Strategy = "intersects"

	// check = a, b, c
	// res = a, b, c
	// -> true
	// check = a, b, c
	// res = a, b, c, d
	// -> false
	// check = a, b, c. d
	// res = a, b, c
	// -> false
	Exact Strategy = "exact"
)

// mfn.se
//
func DNS(resolvers []string, domain string, timeout time.Duration, record Record, strategy Strategy, check []string) (time.Duration, error) {
	start := time.Now()

	Loop:
	for _, addr := range resolvers {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: timeout,
				}
				return d.DialContext(ctx, "udp", fmt.Sprintf("%s:%s", addr, _port))
			},
		}

		var res []string
		var err error
		switch record {
		case A:
			res, err = r.LookupHost(context.Background(), domain)
			if err != nil {
				return time.Since(start), err
			}
		case CNAME:
			name, err := r.LookupCNAME(context.Background(), domain)
			if err != nil {
				return time.Since(start), err
			}
			res = []string{name}
		case TXT:
			res, err = r.LookupTXT(context.Background(), domain)
			if err != nil {
				return time.Since(start), err
			}
		case MX:
			mx, err := r.LookupMX(context.Background(), domain)
			if err != nil {
				return time.Since(start), err
			}
			for _, x := range mx {
				res = append(res, x.Host)
			}
		case NS:
			ns, err := r.LookupNS(context.Background(), domain)
			if err != nil {
				return time.Since(start), err
			}
			for _, n := range ns {
				res = append(res, n.Host)
			}
		default:
			return 0, fmt.Errorf("selected records, %v, is not implmented", record)
		}

		var dnsSet = map[string]struct{}{}
		for _, r := range res{
			dnsSet[r] = struct{}{}
		}
		var checkSet = map[string]struct{}{}
		for _, r := range check{
			checkSet[r] = struct{}{}
		}


		switch strategy {
		case Exact:
			if len(dnsSet) != len(checkSet) {
				return time.Since(start), errors.New("dns result size does not match expected number of records")
			}
			fallthrough
		case CheckIsSubset:
			for c := range checkSet {
				_, ok := dnsSet[c]
				if !ok {
					return time.Since(start), errors.New("all checks were not contained in dns result")
				}
			}
			continue Loop

		case DNSIsSubset:
			for c := range dnsSet {
				_, ok := checkSet[c]
				if !ok {
					return time.Since(start), errors.New("all dns results where not contained in checks")
				}
			}
			continue Loop

		case Intersects:
			for c := range checkSet {
				_, ok := dnsSet[c]
				if ok {
					continue Loop
				}
			}
			return time.Since(start), errors.New("dns result did not intersect with check")

		default:
			return 0, fmt.Errorf("selected strategy, %v, is not implmented", strategy)
		}


	}

	return time.Since(start), nil
}
