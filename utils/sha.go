package utils

import (
	"crypto/sha1"
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"io"
)

func EncryptSha1(data string) (string, error) {
	t := sha1.New()
	_, err := io.WriteString(t, data)
	if err != nil {
		return "", err
	}
	enString := fmt.Sprintf("%x", t.Sum(nil))
	return enString, nil
}

func EncryptSha1WithSalt(data string) (string, error) {
	enData, err := EncryptSha1(data + config.Instance.Security.Salt)
	if err != nil {
		return "", err
	}
	return enData, nil
}
