package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
	"github.com/allentom/youcomic-api/config"
	"github.com/rs/cors"
)

func RunApiService() {
	engine := haruka.NewEngine()
	engine.UseCors(cors.AllowAll())
	engine.UseMiddleware(middleware.NewLoggerMiddleware())
	engine.UseMiddleware(AuthMiddleware{})
	engine.UseMiddleware(StaticMiddleware{})
	SetRouter(engine)
	engine.Router.Static("/assets", config.Instance.Store.Root)
	engine.RunAndListen(config.Instance.Application.Host + ":" + config.Instance.Application.Port)
}
