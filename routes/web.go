package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpmiddleware "github.com/goravel/framework/http/middleware"
)

func Web() {

	facades.Route().Middleware(httpmiddleware.Throttle("testResponse")).Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().Json(http.StatusOK, http.Json{
			"version": "1.0.0",
		})
	})

}
