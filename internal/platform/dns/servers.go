package dns

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

var (
	once sync.Once

	_DNSServerEndpoints = [8]string{ // https://public-dns.info/
		"dk", // Denmark
		"fi", // Finland
		"de", // Germany
		"is", // Iceland
		"no", // Norway
		"se", // Sweden
		"gb", // UK
		"us", // US
	}
	dnsServerIpAddr []string
	defaultDNSServers = []string{"8.8.8.8", "1.1.1.1"}
)

type publicDNSJSON struct {
	Ip 			string 		`json:"ip"`
	Reliability float32 	`json:"reliability"`
	CheckedAt 	time.Time 	`json:"checked_at"`
}

func Get() []string {
	once.Do(func(){
		oneMonthAgo := time.Now().AddDate(0, -1, 0)
		for _, country := range _DNSServerEndpoints {
			resp, err := http.Get(fmt.Sprintf("https://public-dns.info/nameserver/%s.json", country))
			if err != nil {
				logrus.Error(fmt.Sprintf("Error fetching DNS server from %s: %s", country, err.Error()))
			}
			var dns []publicDNSJSON
			err = json.NewDecoder(resp.Body).Decode(&dns)
			if err != nil {
				logrus.Error(fmt.Sprintf("Unable to parse DNSJSON from %s: %s", country, err.Error()))
			}
			for _, dnsServer := range dns {
				if dnsServer.Reliability == 1 && oneMonthAgo.After(dnsServer.CheckedAt){
					dnsServerIpAddr = append(dnsServerIpAddr, dnsServer.Ip)
					break
				}
			}
		}
	})
	if len(dnsServerIpAddr) > 0 {
		return dnsServerIpAddr
	}
	return defaultDNSServers
}