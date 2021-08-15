package youplus

import (
	"github.com/allentom/youcomic-api/config"
	entry "github.com/project-xpolaris/youplustoolkit/youplus/entity"
)

var DefaultEntry *entry.EntityClient

type AppExport struct {
	Addrs []string `json:"addrs"`
}

func InitEntity() {
	DefaultEntry = entry.NewEntityClient(config.Config.YouPlus.EntityConfig.Name, config.Config.YouPlus.EntityConfig.Version, &entry.EntityExport{}, DefaultRPCClient)
	DefaultEntry.HeartbeatRate = 3000
}
