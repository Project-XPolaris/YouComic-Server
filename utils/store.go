package utils

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"path"
)

func GetThumbnailStorePath(bookId uint) string {
	return path.Join(config.Config.Store.Root, "generate", fmt.Sprintf("%d", bookId))
}