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
	DefaultEntry = entry.NewEntityClient(config.Instance.YouPlus.EntityConfig.Name, config.Instance.YouPlus.EntityConfig.Version, &entry.EntityExport{}, DefaultRPCClient)
	DefaultEntry.HeartbeatRate = 3000
}
