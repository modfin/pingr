package sec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
)

type User struct {
	Password   string `json:"password" db:"-"`
	Ciphertext string `json:"-" db:"Ciphertext"`
}

func (u *User) Seal() error {
	plaintext, err  := json.Marshal(u)
	if err != nil{
		return err
	}

	u.Ciphertext, err = seal(plaintext)
	if err != nil{
		return err
	}
	u.Password = ""
	return nil
}

func (u *User) Open() error {
	data, err := open(u.Ciphertext)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &u)
	if err != nil {
		return err
	}
	u.Ciphertext = ""
	return nil
}

type SSHKey struct {
	PEM        string `json:"key" db:"-"`
	Ciphertext string `json:"-" db:"Ciphertext"`
}

func (u *SSHKey) Seal() error {
	plaintext, err  := json.Marshal(u)
	if err != nil{
		return err
	}

	u.Ciphertext, err = seal(plaintext)
	if err != nil{
		return err
	}
	u.PEM = ""
	return nil
}

func (u *SSHKey) Open() error {
	data, err := open(u.Ciphertext)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &u)
	if err != nil {
		return err
	}
	u.Ciphertext = ""
	return nil
}


func seal(plaintext []byte) (string, error){
	// TODO get real key
	key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")


	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "",err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(nonce) + "." + base64.StdEncoding.EncodeToString(ciphertext), nil

}
func open(data string) ([]byte, error){
	// TODO get real key
	key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")

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

