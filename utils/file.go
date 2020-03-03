package utils

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
	"time"
)

func EncodeFileName(fileName string) string {
	nowString := time.Now().String()
	ext := filepath.Ext(fileName)
	return fmt.Sprintf("%x%s", md5.Sum([]byte(fileName+nowString)), ext)
}