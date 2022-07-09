package plugin

import (
	"github.com/allentom/harukap/plugins/storage"
	"github.com/projectxpolaris/youcomic/config"
)

var StorageEngine = &storage.Engine{}

func GetDefaultStorage() storage.FileSystem {
	defaultStorageName := config.DefaultConfigProvider.Manager.GetString("storage.default")
	return StorageEngine.GetStorage(defaultStorageName)
}

func GetDefaultBucket() string {
	return config.DefaultConfigProvider.Manager.GetString("storage.defaultBucket")
}
