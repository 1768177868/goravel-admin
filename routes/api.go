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
	queueTestController := api.NewQueueTestController()

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

	// 队列驱动测试路由（开发调试使用，仅允许开发环境）
	facades.Route().Prefix("api/queue-test").Middleware(middleware.DevelopmentOnly()).Group(func(router route.Router) {
		// 一次性覆盖 dispatch + delay + long-running + fail。
		router.Post("all-in-one", queueTestController.AllInOne)
		// 一次性覆盖 delay + fail + reclaim（偏异常/边界场景）。
		router.Post("all-special", queueTestController.AllSpecial)
		// reclaim 测试：在 redis_stream 下可验证 claim
		router.Post("reclaim", queueTestController.Reclaim)
		// 默认队列立即投递。
		router.Post("dispatch", queueTestController.Dispatch)
		// 默认队列延迟投递（rabbitmq 需 delayed-message 插件才能严格延迟）。
		router.Post("delay", queueTestController.Delay)
		// 投递到 long-running 逻辑队列。
		router.Post("long-running", queueTestController.LongRunning)
		// 投递一个必失败任务，用于失败重试链路验证。
		router.Post("fail", queueTestController.Fail)
		// 分段重试测试：5s -> 10s -> 20s。
		router.Post("backoff", queueTestController.Backoff)
		// 查看各测试 Job 的执行结果缓存。
		router.Get("result", queueTestController.Result)
		// 清空测试结果缓存，便于下一轮测试对比。
		router.Post("reset", queueTestController.Reset)
	})
}
