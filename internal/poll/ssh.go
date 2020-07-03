package poll

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"time"
)

/*
TODO: Add support for both PublicKey validation with or without password
TODO: Add support for username and password without and Keys
TODO: How to read private keys?
*/
func SSH(hostname string, port string, timeOut time.Duration, username string, password string, useKeyPair bool) (time.Duration, error) {
	var authMethod ssh.AuthMethod

	if useKeyPair {
		// Somehow read a key pair
		var key []byte
		key, err := ioutil.ReadFile("id_ed25519")
		if err != nil {
			return 0, err
		}
		var signer ssh.Signer
		if password != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(password))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return 0, err
		}
		authMethod = ssh.PublicKeys(signer)
	} else {
		authMethod = ssh.Password(password)
	}

	config := &ssh.ClientConfig{
		User:              username,
		Auth:              []ssh.AuthMethod{
			authMethod,
		},
		Timeout: timeOut,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Add public key check
	}

	start := time.Now() // Maybe there's a better way of doing this
	client, err := ssh.Dial("tcp",
							fmt.Sprintf("%s:%s", hostname, port),
							config)
	if err != nil {
		return time.Since(start), err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return time.Since(start), err
	}
	defer session.Close()

	return time.Since(start), err
}