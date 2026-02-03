package routes

import (
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/http/controllers"
)

func Web() {

	// facades.Route().Middleware(httpmiddleware.Throttle("testResponse")).Get("/", func(ctx http.Context) http.Response {
	// 	return ctx.Response().Json(http.StatusOK, http.Json{
	// 		"version": "1.0.0",
	// 	})
	// })

	// Swagger
	swaggerController := controllers.NewSwaggerController()
	facades.Route().Get("/swagger/*any", swaggerController.Index)

	// 健康检查
	facades.Route().Get("/health", func(ctx http.Context) http.Response {
		return ctx.Response().Json(200, http.Json{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

}
