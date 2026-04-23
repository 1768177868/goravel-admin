package routes

import (
	"github.com/goravel/framework/contracts/route"
	httpmiddleware "github.com/goravel/framework/http/middleware"

	"goravel/app/facades"
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
	positionController := admin.NewPositionController()
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
	observabilityController := admin.NewObservabilityController()
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
	codeGeneratorController := admin.NewCodeGeneratorController()
	articleController := admin.NewArticleController()
	formDemoController := admin.NewFormDemoController()

	// Admin 路由组：统一前缀和域名限制
	facades.Route().Prefix("api/admin").Middleware(middleware.Domain(facades.Config().Get("domains.admin"))).Group(func(router route.Router) {

		// 登录相关
		router.Middleware(middleware.Lang()).Group(func(router route.Router) {
			router.Middleware(httpmiddleware.Throttle("login")).Post("login", adminAuthController.Login)
			router.Get("login/captcha", adminAuthController.Captcha)

			// 公开的图片访问接口
			router.Get("public/images/{id}", attachmentController.Preview)
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
			router.Post("notifications/ws-ticket", notificationWsController.Ticket)
			router.Post("notifications/{id}/read", notificationController.MarkRead)
			router.Post("notifications/read-all", notificationController.MarkAllRead)

			// 统一的下拉选项接口（不需要权限验证）
			router.Get("options", optionController.Index)
			router.Get("form-demo/data", formDemoController.GetData)

			// 菜单树（仅登录即可，不校验菜单权限；用于角色/权限表单、刷新后展示等）
			router.Get("menus/tree", menuController.Tree)

			router.Get("dictionaries/types", dictionaryController.GetAllTypes)

			// 附件预览（公开接口，以便富文本编辑器等可以直接访问图片）
			// 注意：如果需要严格的权限控制，应使用带签名的URL或Query Token
			// router.Get("attachments/{id}/preview", attachmentController.Preview)
			// 新增一个专门的公开图片访问接口（其实复用 preview 即可，这里为了明确意图，可以考虑别名，或者就用这个）
			// 如果您希望有一个完全独立的接口，可以叫 public/images/{id} 之类，但逻辑是一样的
			// 目前 attachmentController.Preview 已经是处理图片流的了
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
			router.Post("admins/{id}/reset-google-auth", adminController.ResetGoogleAuthenticator)   // 重置管理员的谷歌验证码（无需验证码）

			// 角色管理 - 使用 Resource 路由
			router.Resource("roles", roleController)

			// 权限管理 - 使用 Resource 路由
			router.Resource("permissions", permissionController)

			// 菜单管理 - 使用 Resource 路由
			router.Resource("menus", menuController)

			// 部门管理 - 使用 Resource 路由
			router.Resource("departments", departmentController)

			// 岗位管理
			router.Resource("positions", positionController)

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
			router.Get("observability/trace", observabilityController.TraceAggregate)
			router.Get("observability/slow-sql/top", observabilityController.SlowSQLTopN)
			router.Get("observability/audit-timeline", observabilityController.AuditTimeline)
			router.Get("observability/queue-dashboard", observabilityController.QueueDashboard)

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
			// 恢复预览路由，以便兼容需要鉴权的预览请求
			router.Get("attachments/{id}/preview", attachmentController.Preview)
			router.Get("attachments/{id}/download", attachmentController.Download)
			router.Put("attachments/{id}/display-name", attachmentController.UpdateDisplayName)
			router.Delete("attachments/{id}", attachmentController.Destroy)
			router.Post("attachments/batch-delete", attachmentController.BatchDestroy)

			// 订单管理
			router.Resource("orders", orderController)
			router.Post("orders/export", orderController.Export)
			router.Post("orders/import", orderController.Import)
			router.Get("orders/export/status/{id}", orderController.GetExportStatus)

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

			router.Resource("articles", articleController)
			router.Post("articles/export", articleController.Export)

			// 代码生成器（仅在开发环境可用）
			router.Middleware(middleware.DevelopmentOnly()).Group(func(router route.Router) {
				router.Get("code-generator/field-types", codeGeneratorController.GetFieldTypes)
				router.Get("code-generator/tables", codeGeneratorController.GetTables)
				router.Get("code-generator/table-columns", codeGeneratorController.GetTableColumns)
				router.Post("code-generator/preview", codeGeneratorController.Preview)
				router.Post("code-generator/generate", codeGeneratorController.Generate)
				router.Post("code-generator/save", codeGeneratorController.Save)
				router.Post("code-generator/generate-with-ai", codeGeneratorController.GenerateWithAI)
			})

		})

	})

	// 通知 WebSocket（不在域名限制范围内）
	facades.Route().Get("/ws/admin/notifications", notificationWsController.Server)
}
