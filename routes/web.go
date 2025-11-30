package routes

import (
	"github.com/goravel/framework/facades"

	"goravel/app/http/controllers"
)

func Web() {
	// facades.Route().Get("/", func(ctx http.Context) http.Response {
	// 	return ctx.Response().View().Make("welcome.tmpl", map[string]any{
	// 		"version": support.Version,
	// 	})
	// })

	// Swagger
	swaggerController := controllers.NewSwaggerController()
	facades.Route().Get("/swagger/*any", swaggerController.Index)

	// Single Page Application
	// 1. Add your single page application to `resources/views/*`
	// 2. Add route to `/route/web.go`, needs to contain your home page and static routes
	// 3. Configure nginx based on the /nginx.conf file
	// facades.Route().StaticFile("index.html", "./resources/views/index.html")
	// facades.Route().Static("css", "./resources/views/css")

	// View Nesting
	// Check the views in `resources/views/admin/*`
	// facades.Route().Get("view", func(ctx http.Context) http.Response {
	// 	return ctx.Response().View().Make("admin/index.tmpl", map[string]any{
	// 		"name": "Goravel",
	// 	})
	// })

	// Session
	// facades.Route().Prefix("session").Group(func(router route.Router) {
	// 	router.Get("put", func(ctx http.Context) http.Response {
	// 		ctx.Request().Session().Put("name", "Goravel")

	// 		return ctx.Response().Success().Json(http.Json{
	// 			"name": cast.ToString(ctx.Request().Session().Get("name")),
	// 		})
	// 	})
	// 	router.Get("get", func(ctx http.Context) http.Response {
	// 		return ctx.Response().Success().Json(http.Json{
	// 			"name": ctx.Request().Session().Get("name"),
	// 		})
	// 	})
	// })

	// Cookie
	// facades.Route().Prefix("cookie").Group(func(router route.Router) {
	// 	router.Get("put", func(ctx http.Context) http.Response {
	// 		ctx.Response().Cookie(http.Cookie{
	// 			Name:  "name",
	// 			Value: "Goravel",
	// 		})

	// 		return ctx.Response().Success().String("Set cookie: name=Goravel")
	// 	})
	// 	router.Get("get", func(ctx http.Context) http.Response {
	// 		return ctx.Response().Success().Json(http.Json{
	// 			"name": ctx.Request().Cookie("name"),
	// 		})
	// 	})
	// })

	// facades.Route().Fallback(func(ctx http.Context) http.Response {
	// 	return ctx.Response().String(http.StatusNotFound, "fallback")
	// })

	// 1. 聊天室页面路由（访问聊天界面）
	// facades.Route().StaticFile("chat.html", "./resources/views/chat.html")

	// 2. WebSocket连接路由（后端处理消息）
	// chatController := controllers.NewChatController()
	// facades.Route().Get("/ws/chat", chatController.Server)

	// 3. HTTP接口发送聊天室消息（支持GET/POST）
	// facades.Route().Post("/api/chat/send", chatController.SendMsgByHttp)
	// facades.Route().Get("/api/chat/send", chatController.SendMsgByHttp)
}
