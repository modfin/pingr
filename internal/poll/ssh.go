package poll

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"pingr/internal/sec"
	"time"
)

func SSH(hostname string, port string, timeOut time.Duration, username string, credentialType string, credential string) (time.Duration, error) {
	var authMethod ssh.AuthMethod

	protected := sec.Protected{
		Cipher: credential,
	}
	err := protected.Open()
	if err != nil {
		return 0, errors.New("could not open protected data: " + err.Error())
	}

	switch credentialType {
	case "key":
		signer, err := ssh.ParsePrivateKey([]byte(protected.Plain))
		if err != nil {
			return 0, errors.New("could not parse private key: " + err.Error())
		}
		authMethod = ssh.PublicKeys(signer)
	case "userpass":
		authMethod = ssh.Password(protected.Plain)
	default:
		return 0, fmt.Errorf("invalid ssh credential type %s", credentialType)
	}

	err = protected.Seal()
	if err != nil {
		return 0, errors.New("could not seal ssh key: " + err.Error())
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		Timeout:         timeOut * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Add public key check
	}

	start := time.Now()
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
