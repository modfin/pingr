package poll

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"golang.org/x/crypto/ocsp"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

func TLS(hostname string, port string, timeOut time.Duration) (time.Duration, error) {

	now := time.Now()
	then := now.AddDate(0, 1, 0)

	dialer := net.Dialer{Timeout: timeOut * time.Second}
	netconn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%s", hostname, port))
	if err != nil {
		return time.Since(now), err
	}
	defer netconn.Close()

	conf := &tls.Config{ServerName: hostname}
	cli := tls.Client(netconn, conf)
	defer cli.Close()

	err = cli.Handshake()
	if err != nil {
		return time.Since(now), err
	}

	err = cli.VerifyHostname(hostname)
	if err != nil {
		return time.Since(now), err
	}

	//// Upgrading to go1.14 is probably preferable in order to do better inspection of CipherSuites
	//// https://golang.org/doc/go1.14#crypto/tls

	state := cli.ConnectionState()
	certs := state.PeerCertificates
	suite := GetCipherSuite(state.CipherSuite)

	if suite == nil {
		err = fmt.Errorf("could not find valid cipher suite for %d", state.CipherSuite)
		return time.Since(now), err
	}
	if suite.Insecure {
		err = fmt.Errorf("an insecure cipher suite is used, %s", suite.Name)
		return time.Since(now), err

	}

	named := map[string]*x509.Certificate{}
	for _, cert := range certs {
		named[cert.Subject.String()] = cert
	}
	for _, cert := range certs {
		if now.Before(cert.NotBefore) {
			err = fmt.Errorf("cetificat is not yet valid: %s", cert.Subject.String())
			return time.Since(now), err
		}
		if then.After(cert.NotAfter) {
			err = fmt.Errorf("cetificat will expire in %s: %s", cert.NotAfter.Sub(now).String(), cert.Subject.String())
			return time.Since(now), err
		}

		if !cert.IsCA {
			issuer, ok := named[cert.Issuer.String()]
			if !ok {
				err = fmt.Errorf("could not find issuer for %s", cert.Subject.String())
				return time.Since(now), err
			}
			var res *ocsp.Response
			res, err = GetOCSP(cert, issuer)
			if err != nil {
				return time.Since(now), err
			}
			if res.Status == ocsp.Revoked {
				err = fmt.Errorf("certificate has been revoked by issuer")
				return time.Since(now), err
			}
		}

		//fmt.Println("cert - exp:", cert.NotAfter)
		//fmt.Println("      name:", cert.Subject)
		//fmt.Println("    issuer:", cert.Issuer)
		//fmt.Println("       sub:", base64.StdEncoding.EncodeToString(cert.SubjectKeyId))
		//fmt.Println("      auth:", base64.StdEncoding.EncodeToString(cert.AuthorityKeyId))
		//fmt.Println("        ca:", cert.IsCA)
	}

	return time.Since(now), err
}

func GetOCSP(clientCert, issuerCert *x509.Certificate) (res *ocsp.Response, err error) {
	servers := issuerCert.OCSPServer
	if len(servers) < 1 {
		return nil, fmt.Errorf("could not find any ocsp servers")
	}
	ocspServer := servers[0]

	opts := &ocsp.RequestOptions{Hash: crypto.SHA1}
	buffer, err := ocsp.CreateRequest(clientCert, issuerCert, opts)
	if err != nil {
		return nil, err
	}
	httpRequest, err := http.NewRequest(http.MethodPost, ocspServer, bytes.NewBuffer(buffer))
	if err != nil {
		return
	}
	ocspUrl, err := url.Parse(ocspServer)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Add("Content-Type", "application/ocsp-request")
	httpRequest.Header.Add("Accept", "application/ocsp-response")
	httpRequest.Header.Add("host", ocspUrl.Host)
	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()
	output, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	return ocsp.ParseResponse(output, issuerCert)
}

type CipherSuite struct {
	ID   uint16
	Name string

	// Insecure is true if the cipher suite has known security issues
	// due to its primitives, design, or implementation.
	Insecure bool
}

func CipherSuites() []*CipherSuite {
	return []*CipherSuite{
		{tls.TLS_RSA_WITH_AES_128_CBC_SHA, "TLS_RSA_WITH_AES_128_CBC_SHA", false},
		{tls.TLS_RSA_WITH_AES_256_CBC_SHA, "TLS_RSA_WITH_AES_256_CBC_SHA", false},
		{tls.TLS_RSA_WITH_AES_128_GCM_SHA256, "TLS_RSA_WITH_AES_128_GCM_SHA256", false},
		{tls.TLS_RSA_WITH_AES_256_GCM_SHA384, "TLS_RSA_WITH_AES_256_GCM_SHA384", false},

		{tls.TLS_AES_128_GCM_SHA256, "TLS_AES_128_GCM_SHA256", false},
		{tls.TLS_AES_256_GCM_SHA384, "TLS_AES_256_GCM_SHA384", false},
		{tls.TLS_CHACHA20_POLY1305_SHA256, "TLS_CHACHA20_POLY1305_SHA256", false},

		{tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA, "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA", false},
		{tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA, "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA", false},
		{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA", false},
		{tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA, "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA", false},
		{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256", false},
		{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", false},
		{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", false},
		{tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384", false},
		//{tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256, "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256", false},
		//{tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256, "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256", false},
	}
}

// InsecureCipherSuites returns a list of cipher suites currently implemented by
// this package and which have security issues.
//
// Most applications should not use the cipher suites in this list, and should
// only use those returned by CipherSuites.
func InsecureCipherSuites() []*CipherSuite {
	// RC4 suites are broken because RC4 is.
	// CBC-SHA256 suites have no Lucky13 countermeasures.
	return []*CipherSuite{
		{tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA, "TLS_RSA_WITH_3DES_EDE_CBC_SHA", true},
		{tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA, "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA", true},
		{tls.TLS_RSA_WITH_RC4_128_SHA, "TLS_RSA_WITH_RC4_128_SHA", true},
		{tls.TLS_RSA_WITH_AES_128_CBC_SHA256, "TLS_RSA_WITH_AES_128_CBC_SHA256", true},
		{tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA, "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA", true},
		{tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA, "TLS_ECDHE_RSA_WITH_RC4_128_SHA", true},
		{tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256, "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256", true},
		{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256, "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256", true},
	}
}

// CipherSuiteName returns the standard name for the passed cipher suite ID
// (e.g. "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"), or a fallback representation
// of the ID value if the cipher suite is not implemented by this package.
func GetCipherSuite(id uint16) *CipherSuite {
	for _, c := range CipherSuites() {
		if c.ID == id {
			return c
		}
	}
	for _, c := range InsecureCipherSuites() {
		if c.ID == id {
			return c
		}
	}
	return nil
}
