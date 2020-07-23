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
	
	switch(credentialType) {
	case "key":
		key := sec.SSHKey{
			Ciphertext: credential,
		}
		err := key.Open()
		if err != nil {
			return 0, errors.New("could not open ssh key: " + err.Error())
		}
		signer, err := ssh.ParsePrivateKey([]byte(key.PEM))
		authMethod = ssh.PublicKeys(signer)
		err = key.Seal()
		if err != nil {
			return 0, errors.New("could not seal ssh key: " + err.Error())
		}
	case "userpass":
		user := sec.User{
			Ciphertext: credential,
		}
		err := user.Open()
		if err != nil {
			return 0, errors.New("could not open password: " + err.Error())
		}
		authMethod = ssh.Password(user.Password)
		err = user.Seal()
		if err != nil {
			return 0, errors.New("could not seal password: " + err.Error())
		}

	default:
		return 0, fmt.Errorf("invalid ssh credential type %s", credentialType)
	}

	config := &ssh.ClientConfig{
		User:              username,
		Auth:              []ssh.AuthMethod{
			authMethod,
		},
		Timeout: timeOut*time.Second,
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