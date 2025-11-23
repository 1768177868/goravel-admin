package routes

import (
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"

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
	operationLogController := admin.NewOperationLogController()
	loginLogController := admin.NewLoginLogController()
	systemLogController := admin.NewSystemLogController()
	dashboardController := admin.NewDashboardController()

	// 登录相关（不需要认证，但需要多语言）
	facades.Route().Prefix("api/admin").Middleware(middleware.Lang()).Group(func(router route.Router) {
		router.Post("login", adminAuthController.Login)
	})

	// 刷新token接口（允许token过期但仍在刷新窗口内的请求）

	// 基础功能（需要认证和多语言，但不需要权限验证和操作日志）
	facades.Route().Prefix("api/admin").Middleware(middleware.Lang(), middleware.Jwt()).Group(func(router route.Router) {
		// 认证相关
		router.Get("info", adminAuthController.Info)
		router.Put("profile", adminAuthController.UpdateProfile)
		router.Post("logout", adminAuthController.Logout)
		// Token管理
		router.Get("tokens", adminAuthController.Tokens)
		router.Delete("tokens/{id}", adminAuthController.RevokeToken)
		router.Delete("tokens", adminAuthController.RevokeAllTokens)
	})

	// 需要认证、多语言、权限验证和操作日志的路由
	facades.Route().Prefix("api/admin").Middleware(middleware.Lang(), middleware.Jwt(), middleware.Permission(), middleware.OperationLog()).Group(func(router route.Router) {
		// 密码管理
		passwordController := admin.NewPasswordController()
		router.Put("password", passwordController.UpdatePassword)
		router.Put("admins/{id}/password", passwordController.ResetPassword)

		// 管理员管理（有额外路由，不能完全用 Resource）
		router.Get("admins", adminController.Index)
		router.Get("admins/export", adminController.Export)
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

		// 字典管理（有额外路由，不能完全用 Resource）
		router.Resource("dictionaries", dictionaryController)
		router.Get("dictionaries/type/{type}", dictionaryController.GetByType)

		// 操作日志
		router.Get("operation-logs", operationLogController.Index)
		router.Get("operation-logs/{id}", operationLogController.Show)
		router.Delete("operation-logs/{id}", operationLogController.Destroy)
		router.Post("operation-logs/batch-delete", operationLogController.BatchDestroy)
		router.Post("operation-logs/clean", operationLogController.Clean)

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
	})
}
