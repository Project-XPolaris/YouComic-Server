package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
	"github.com/projectxpolaris/youcomic/config"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/module"
	"github.com/rs/cors"
)

func GetEngine() *haruka.Engine {
	engine := haruka.NewEngine()
	engine.UseCors(cors.AllowAll())
	engine.UseMiddleware(middleware.NewLoggerMiddleware())
	module.Auth.AuthMiddleware.OnError = func(c *haruka.Context, err error) {
		ApiError.RaiseApiError(c, err, nil)
		c.Abort()
		return
	}
	engine.UseMiddleware(module.Auth.AuthMiddleware)
	engine.UseMiddleware(StaticMiddleware{})
	SetRouter(engine)
	engine.Router.Static("/assets", config.Instance.Store.Root)

	return engine
}
