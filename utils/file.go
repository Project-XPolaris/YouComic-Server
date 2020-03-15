package utils

import (
	"crypto/md5"
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"path/filepath"
	"strconv"
	"time"
)

func EncodeFileName(fileName string) string {
	nowString := time.Now().String()
	ext := filepath.Ext(fileName)
	return fmt.Sprintf("%x%s", md5.Sum([]byte(fileName+nowString)), ext)
}

func GetBookStorePath(bookId uint) string {
	return filepath.Join(config.Config.Store.Books, strconv.Itoa(int(bookId)))
}
