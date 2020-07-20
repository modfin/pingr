package poll

import (
	"testing"
)

func TestTLS(t *testing.T) {
	_, err := TLS("golang.org", "https", 1)
	if err != nil{
		t.Log(err)
		t.Fail()
	}
}

// https://badssl.com/
var badTLS = []struct {
	host string
	port string
	exp  string
}{
	{
		host: "expired.badssl.com",
		port: "https",
		exp:"x509: certificate has expired or is not yet valid",
	},
	{
		host: "wrong.host.badssl.com",
		port: "https",
		exp:"x509: certificate is valid for *.badssl.com, badssl.com, not wrong.host.badssl.com",
	},
	{
		host: "self-signed.badssl.com",
		port: "https",
		exp:"x509: certificate signed by unknown authority",
	},
	{
		host: "untrusted-root.badssl.com",
		port: "https",
		exp:"x509: certificate signed by unknown authority",
	},
	{
		host: "revoked.badssl.com",
		port: "https",
		exp:"certificate has been revoked by issuer",
	},
	//{ //TODO not sure what this should test...
	//	host: "pinning-test.badssl.com",
	//	port: "https",
	//	exp:"",
	//},
	// Bad encryption
	{
		host: "rc4-md5.badssl.com",
		port: "https",
		exp:"remote error: tls: handshake failure",
	},
	{
		host: "rc4.badssl.com",
		port: "https",
		exp:"remote error: tls: handshake failure",
	},
	{
		host: "null.badssl.com",
		port: "https",
		exp:"remote error: tls: handshake failure",
	},
	{
		host: "3des.badssl.com",
		port: "https",
		exp:"an insecure cipher suite is used, TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
	},

	// Bad hand shake
	{
		host: "dh480.badssl.com",
		port: "https",
		exp:"remote error: tls: handshake failure",
	},
	{
		host: "dh512.badssl.com",
		port: "https",
		exp:"remote error: tls: handshake failure",
	},
	{
		host: "dh1024.badssl.com",
		port: "https",
		exp:"remote error: tls: handshake failure",
	},
	{
		host: "dh-small-subgroup.badssl.com",
		port: "https",
		exp:"remote error: tls: handshake failure",
	},
	{
		host: "dh-composite.badssl.com",
		port: "https",
		exp:"remote error: tls: handshake failure",
	},
	//{ // TODO implement check for this
	//	host: "no-sct.badssl.com",
	//	port: "https",
	//	exp:"remote error: tls: handshake failure",
	//},


}

func TestBadTLS(t *testing.T){
	for _, test := range badTLS{
		_, err := TLS(test.host, test.port, 1)
		if err == nil{
			t.Log("Expected error for", test)
			t.Fail()
			continue
		}
		if err.Error() != test.exp{
			t.Logf("Expected error \"%v\" got \"%v\"", test.exp, err.Error())
			t.Logf("  for, %s", test.host )
			t.Fail()
		}
	}
}