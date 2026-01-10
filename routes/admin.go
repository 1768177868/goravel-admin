package routes

import (
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	httpmiddleware "github.com/goravel/framework/http/middleware"

	"goravel/app/http/controllers/admin"
	"goravel/app/http/middleware"
)

func Admin() {
	adminAuthController := admin.NewAuthController()
	adminController := admin.NewAdminController()
	roleController := admin.NewRoleController()
	permissionController := admin.NewPermissionController()
	menuController := admin.NewMenuController()
	departmentController := admin.NewDepartmentController()
	dictionaryController := admin.NewDictionaryController()
	configController := admin.NewConfigController()
	blacklistController := admin.NewBlacklistController()
	onlineAdminController := admin.NewOnlineAdminController()
	operationLogController := admin.NewOperationLogController()
	loginLogController := admin.NewLoginLogController()
	systemLogController := admin.NewSystemLogController()
	dashboardController := admin.NewDashboardController()
	// debugController := admin.NewDebugController()
	monitorController := admin.NewMonitorController()
	notificationController := admin.NewNotificationController()
	notificationWsController := admin.NewNotificationWsController()
	optionController := admin.NewOptionController()
	exportController := admin.NewExportController()
	attachmentController := admin.NewAttachmentController()
	orderController := admin.NewOrderController()
	userController := admin.NewUserController()
	userBalanceLogController := admin.NewUserBalanceLogController()
	paymentMethodController := admin.NewPaymentMethodController()
	paymentController := admin.NewPaymentController()

	// Admin 路由组：统一前缀和域名限制
	facades.Route().Prefix("api/admin").Middleware(middleware.Domain(facades.Config().Get("domains.admin"))).Group(func(router route.Router) {

		// 登录相关（不需要认证，但需要多语言）
		router.Middleware(middleware.Lang()).Group(func(router route.Router) {
			router.Middleware(httpmiddleware.Throttle("login")).Post("login", adminAuthController.Login)
			router.Get("login/captcha", adminAuthController.Captcha)
		})

		// 基础功能（需要认证和多语言，但不需要权限验证和操作日志）
		router.Middleware(middleware.Lang(), middleware.Jwt()).Group(func(router route.Router) {
			// 认证相关
			router.Get("info", adminAuthController.Info)

			router.Post("logout", adminAuthController.Logout)
			router.Get("heartbeat", adminAuthController.Heartbeat)

			// 谷歌验证码相关
			router.Get("google-authenticator/status", adminAuthController.GetGoogleAuthenticatorStatus)
			router.Get("google-authenticator/qrcode", adminAuthController.GetGoogleAuthenticatorQRCode)
			router.Post("google-authenticator/bind", adminAuthController.BindGoogleAuthenticator)
			router.Post("google-authenticator/unbind", adminAuthController.UnbindGoogleAuthenticator)

			// 通知中心
			router.Get("notifications", notificationController.Index)
			router.Get("notifications/unread-count", notificationController.UnreadCount)
			router.Get("notifications/recent", notificationController.Recent)
			router.Post("notifications/{id}/read", notificationController.MarkRead)
			router.Post("notifications/read-all", notificationController.MarkAllRead)

			// 统一的下拉选项接口（不需要权限验证）
			router.Get("options", optionController.Index)
		})

		// 需要认证、多语言、权限验证和操作日志的路由
		router.Middleware(middleware.Lang(), middleware.Jwt(), middleware.Permission(), middleware.OperationLog()).Group(func(router route.Router) {

			router.Put("profile", adminAuthController.UpdateProfile)

			// 密码管理
			passwordController := admin.NewPasswordController()
			router.Put("password", passwordController.UpdatePassword)
			router.Put("admins/{id}/password", passwordController.ResetPassword)

			// 管理员管理
			router.Resource("admins", adminController)
			router.Post("admins/export", adminController.Export)
			router.Delete("admins/{id}/tokens", adminAuthController.KickOutUser)                     // 踢出指定用户的所有token
			router.Post("admins/{id}/unbind-google-auth", adminController.UnbindGoogleAuthenticator) // 解绑管理员的谷歌验证码

			// 角色管理 - 使用 Resource 路由
			router.Resource("roles", roleController)

			// 权限管理 - 使用 Resource 路由
			router.Resource("permissions", permissionController)

			// 菜单管理 - 使用 Resource 路由
			router.Resource("menus", menuController)

			// 部门管理 - 使用 Resource 路由
			router.Resource("departments", departmentController)

			// 字典管理
			router.Resource("dictionaries", dictionaryController)
			router.Get("dictionaries/type/{type}", dictionaryController.GetByType)

			// 配置管理
			router.Get("configs/group/{group}", configController.GetByGroup)
			router.Post("configs/save", configController.Save)
			router.Post("configs/test-email", configController.TestEmail)

			// 黑名单管理
			router.Resource("blacklists", blacklistController)

			// 在线管理员管理
			router.Get("online-admins", onlineAdminController.Index)
			router.Delete("online-admins/{id}", onlineAdminController.KickOut)
			router.Post("online-admins/batch-kick-out", onlineAdminController.BatchKickOut)

			// 操作日志
			router.Get("operation-logs", operationLogController.Index)
			router.Get("operation-logs/title-options", operationLogController.GetTitleOptions)
			router.Get("operation-logs/{id}", operationLogController.Show)
			router.Delete("operation-logs/{id}", operationLogController.Destroy)
			router.Post("operation-logs/batch-delete", operationLogController.BatchDestroy)
			// router.Post("operation-logs/clean", operationLogController.Clean)

			// 导出管理
			router.Get("exports", exportController.Index)
			router.Get("exports/{id}/download", exportController.Download)
			// SSE 路由：实时推送导出任务进度（之前没有进度查询接口，直接使用 SSE）
			// router.Get("exports/{id}/progress", exportController.StreamExportProgress)
			router.Delete("exports/{id}", exportController.Destroy)
			router.Post("exports/batch-delete", exportController.BatchDestroy)

			// 登录日志
			router.Get("login-logs", loginLogController.Index)
			router.Get("login-logs/{id}", loginLogController.Show)
			router.Delete("login-logs/{id}", loginLogController.Destroy)
			router.Post("login-logs/batch-delete", loginLogController.BatchDestroy)
			// router.Post("login-logs/clean", loginLogController.Clean)

			// 系统日志
			router.Get("system-logs", systemLogController.Index)
			router.Get("system-logs/{id}", systemLogController.Show)
			router.Delete("system-logs/{id}", systemLogController.Destroy)
			router.Post("system-logs/batch-delete", systemLogController.BatchDestroy)
			// router.Post("system-logs/clean", systemLogController.Clean)

			// Dashboard 统计
			// 原路由：按需查询特定数据（适合一次性查询或按需刷新）
			router.Get("dashboard/count", dashboardController.GetCount)
			router.Get("dashboard/user-access-source", dashboardController.GetUserAccessSource)
			router.Get("dashboard/weekly-user-activity", dashboardController.GetWeeklyUserActivity)
			router.Get("dashboard/monthly-sales", dashboardController.GetMonthlySales)
			router.Get("dashboard/recent-activities", dashboardController.GetRecentActivities)
			// SSE 路由：实时推送所有 Dashboard 数据（适合实时 Dashboard 页面，自动更新）
			router.Get("dashboard/stream", dashboardController.StreamDashboardData)

			// 服务监控
			// 原路由：手动刷新、一次性查询（适合按需查看或定时刷新）
			router.Get("monitor/system-info", monitorController.GetSystemInfo)
			// SSE 路由：实时推送系统监控数据（适合实时监控页面，自动更新）
			router.Get("monitor/system-info/stream", monitorController.StreamSystemInfo)

			// 系统公告/通知
			router.Post("notifications", notificationController.Store)

			// 调试: trace id 日志验证
			// router.Get("debug/trace-test", debugController.TraceTest)

			// 附件管理
			router.Get("attachments", attachmentController.Index)
			router.Post("attachments/upload", attachmentController.Upload)
			// 统一的分片上传接口（POST，action参数：init/upload/merge）
			router.Post("attachments/chunk", attachmentController.ChunkUpload)
			// 获取上传进度（GET，action=progress）- 适合断点续传检查、一次性查询
			router.Get("attachments/chunk", attachmentController.ChunkUpload)
			router.Get("attachments/{id}/preview", attachmentController.Preview)
			router.Get("attachments/{id}/download", attachmentController.Download)
			router.Put("attachments/{id}/display-name", attachmentController.UpdateDisplayName)
			router.Delete("attachments/{id}", attachmentController.Destroy)
			router.Post("attachments/batch-delete", attachmentController.BatchDestroy)

			// 订单管理
			router.Resource("orders", orderController)
			router.Post("orders/export", orderController.Export)
			router.Post("orders/import", orderController.Import)
			router.Get("orders/export/status/:id", orderController.GetExportStatus)

			// 用户管理
			router.Resource("users", userController)
			router.Post("users/{id}/update-balance", userController.UpdateBalance)
			router.Put("users/{id}/password", userController.ResetPassword)
			router.Post("users/export", userController.Export)

			// 用户余额变动记录
			router.Get("user-balance-logs", userBalanceLogController.Index)
			router.Post("user-balance-logs", userBalanceLogController.Store)
			router.Get("user-balance-logs/statistics", userBalanceLogController.Statistics)

			// 支付方式管理
			router.Resource("payment-methods", paymentMethodController)

			// 支付记录管理
			router.Get("payments", paymentController.Index)
			router.Get("payments/{id}", paymentController.Show)
			router.Post("payments/export", paymentController.Export)
			router.Get("payments/export/status/{id}", paymentController.GetExportStatus)

		})

	})

	// 通知 WebSocket（不在域名限制范围内）
	facades.Route().Get("/ws/admin/notifications", notificationWsController.Server)
}
