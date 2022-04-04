package utils

import (
	"fmt"
	"github.com/projectxpolaris/youcomic/config"
	"path"
)

func GetThumbnailStorePath(bookId uint) string {
	return path.Join(config.Instance.Store.Root, "generate", fmt.Sprintf("%d", bookId))
}
