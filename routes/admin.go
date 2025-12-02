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
	onlineUserController := admin.NewOnlineUserController()
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

	// 登录相关（不需要认证，但需要多语言）
	facades.Route().Prefix("api/admin").Middleware(middleware.Lang()).Group(func(router route.Router) {
		router.Middleware(httpmiddleware.Throttle("login")).Post("login", adminAuthController.Login)
		router.Get("login/captcha", adminAuthController.Captcha)
	})

	// 基础功能（需要认证和多语言，但不需要权限验证和操作日志）
	facades.Route().Prefix("api/admin").Middleware(middleware.Lang(), middleware.Jwt()).Group(func(router route.Router) {
		// 认证相关
		router.Get("info", adminAuthController.Info)

		router.Post("logout", adminAuthController.Logout)
		router.Get("heartbeat", adminAuthController.Heartbeat)

		// Token管理
		// router.Get("tokens", adminAuthController.Tokens)
		// router.Delete("tokens/{id}", adminAuthController.RevokeToken)
		// router.Delete("tokens", adminAuthController.RevokeAllTokens)

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
	facades.Route().Prefix("api/admin").Middleware(middleware.Lang(), middleware.Jwt(), middleware.Permission(), middleware.OperationLog()).Group(func(router route.Router) {

		router.Put("profile", adminAuthController.UpdateProfile)

		// 密码管理
		passwordController := admin.NewPasswordController()
		router.Put("password", passwordController.UpdatePassword)
		router.Put("admins/{id}/password", passwordController.ResetPassword)

		// 管理员管理
		router.Get("admins", adminController.Index)
		router.Post("admins/export", adminController.Export)
		router.Get("admins/{id}", adminController.Show)
		router.Post("admins", adminController.Store)
		router.Put("admins/{id}", adminController.Update)
		router.Delete("admins/{id}", adminController.Destroy)
		router.Delete("admins/{id}/tokens", adminAuthController.KickOutUser) // 踢出指定用户的所有token

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

		// 在线用户管理
		router.Get("online-users", onlineUserController.Index)
		router.Delete("online-users/{id}", onlineUserController.KickOut)
		router.Post("online-users/batch-kick-out", onlineUserController.BatchKickOut)

		// 操作日志
		router.Get("operation-logs", operationLogController.Index)
		router.Get("operation-logs/title-options", operationLogController.GetTitleOptions)
		router.Get("operation-logs/{id}", operationLogController.Show)
		router.Delete("operation-logs/{id}", operationLogController.Destroy)
		router.Post("operation-logs/batch-delete", operationLogController.BatchDestroy)
		router.Post("operation-logs/clean", operationLogController.Clean)

		// 导出管理
		router.Get("exports", exportController.Index)
		router.Get("exports/{id}/download", exportController.Download)
		router.Delete("exports/{id}", exportController.Destroy)
		router.Post("exports/batch-delete", exportController.BatchDestroy)

		// 登录日志
		router.Get("login-logs", loginLogController.Index)
		router.Get("login-logs/{id}", loginLogController.Show)
		router.Delete("login-logs/{id}", loginLogController.Destroy)
		router.Post("login-logs/batch-delete", loginLogController.BatchDestroy)
		router.Post("login-logs/clean", loginLogController.Clean)

		// 系统日志
		router.Get("system-logs", systemLogController.Index)
		router.Get("system-logs/{id}", systemLogController.Show)
		router.Delete("system-logs/{id}", systemLogController.Destroy)
		router.Post("system-logs/batch-delete", systemLogController.BatchDestroy)
		router.Post("system-logs/clean", systemLogController.Clean)

		// Dashboard 统计
		router.Get("dashboard/count", dashboardController.GetCount)
		router.Get("dashboard/user-access-source", dashboardController.GetUserAccessSource)
		router.Get("dashboard/weekly-user-activity", dashboardController.GetWeeklyUserActivity)
		router.Get("dashboard/monthly-sales", dashboardController.GetMonthlySales)

		// 服务监控
		router.Get("monitor/system-info", monitorController.GetSystemInfo)

		// 系统公告/通知
		router.Post("notifications", notificationController.Store)

		// 调试: trace id 日志验证
		// router.Get("debug/trace-test", debugController.TraceTest)

		// 附件管理
		router.Get("attachments", attachmentController.Index)
		router.Post("attachments/upload", attachmentController.Upload)
		router.Post("attachments/chunk", attachmentController.ChunkUpload) // 统一的分片上传接口（POST，action参数）
		router.Get("attachments/chunk", attachmentController.ChunkUpload)  // 获取进度（GET，action=progress）
		router.Get("attachments/{id}/preview", attachmentController.Preview)
		router.Get("attachments/{id}/download", attachmentController.Download)
		router.Put("attachments/{id}/display-name", attachmentController.UpdateDisplayName)
		router.Delete("attachments/{id}", attachmentController.Destroy)
		router.Post("attachments/batch-delete", attachmentController.BatchDestroy)
	})

	// 通知 WebSocket
	facades.Route().Get("/ws/admin/notifications", notificationWsController.Server)
}
