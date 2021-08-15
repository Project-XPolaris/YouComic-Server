package youplus

import (
	"context"
	"github.com/allentom/youcomic-api/config"
	"github.com/project-xpolaris/youplustoolkit/youplus/rpc"
	"time"
)

var DefaultRPCClient *rpc.YouPlusRPCClient

func LoadYouPlusRPCClient() error {
	DefaultRPCClient = rpc.NewYouPlusRPCClient(config.Config.YouPlus.RPCUrl)
	DefaultRPCClient.KeepAlive = true
	DefaultRPCClient.MaxRetry = 1000
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return DefaultRPCClient.Connect(ctx)
}
