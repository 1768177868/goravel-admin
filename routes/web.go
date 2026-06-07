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

	// Swagger is disabled by default in production. Enable with SWAGGER_ENABLED=true
	// only for controlled environments.
	if facades.Config().GetBool("swagger.enabled", false) {
		swaggerController := controllers.NewSwaggerController()
		facades.Route().Get("/swagger/*any", swaggerController.Index)
	}

	// 健康检查
	facades.Route().Get("/health", func(ctx http.Context) http.Response {
		return ctx.Response().Json(200, http.Json{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

}
