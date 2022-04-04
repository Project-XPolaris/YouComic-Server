package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
	"github.com/projectxpolaris/youcomic/config"
	"github.com/rs/cors"
)

func GetEngine() *haruka.Engine {
	engine := haruka.NewEngine()
	engine.UseCors(cors.AllowAll())
	engine.UseMiddleware(middleware.NewLoggerMiddleware())
	engine.UseMiddleware(AuthMiddleware{})
	engine.UseMiddleware(StaticMiddleware{})
	SetRouter(engine)
	engine.Router.Static("/assets", config.Instance.Store.Root)
	return engine
}
