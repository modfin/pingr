package sec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"pingr/internal/config"
	"strings"
)

type Protected struct {
	Plain  string `json:"plain" db:"-"`
	Cipher string `json:"-" db:"cipher"`
}

func (u *Protected) Seal() error {
	var err error
	u.Cipher, err = seal([]byte(u.Plain))
	if err != nil {
		return err
	}
	u.Plain = ""
	return nil
}

func (u *Protected) Open() error {
	data, err := open(u.Cipher)
	if err != nil {
		return err
	}
	u.Plain = string(data)
	u.Cipher = ""
	return nil
}

func seal(plaintext []byte) (string, error) {
	key, _ := hex.DecodeString(config.Get().AESKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(nonce) + "." + base64.StdEncoding.EncodeToString(ciphertext), nil

}
func open(data string) ([]byte, error) {
	key, _ := hex.DecodeString(config.Get().AESKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(data, ".")
	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonce, ciphertext, nil)
}
