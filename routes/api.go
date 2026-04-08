package routes

import (
	"github.com/goravel/framework/contracts/route"
	httpmiddleware "github.com/goravel/framework/http/middleware"

	"goravel/app/facades"
	"goravel/app/http/controllers/api"
	"goravel/app/http/middleware"
)

func Api() {
	authController := api.NewAuthController()
	orderSearchController := api.NewOrderController()

	// C端用户路由组：统一前缀
	facades.Route().Prefix("api/user").Group(func(router route.Router) {

		// 登录注册相关（不需要认证，但需要限流）
		router.Middleware(middleware.Lang(), httpmiddleware.Throttle("login")).Group(func(router route.Router) {
			router.Post("register", authController.Register)
			router.Post("login", authController.Login)
		})

		// 需要认证的路由
		router.Middleware(middleware.UserJwt()).Group(func(router route.Router) {
			// 用户信息
			router.Get("info", authController.Info)
			// 登出
			router.Post("logout", authController.Logout)
			// ES 关键词搜「我的订单」演示（需开启 ELASTICSEARCH_ENABLED，索引需已有数据）
			router.Get("orders/search", orderSearchController.SearchMyOrders)
		})

	})
}
