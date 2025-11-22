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

	// 登录相关（不需要认证，但需要多语言）
	facades.Route().Prefix("admin").Middleware(middleware.Lang()).Group(func(router route.Router) {
		router.Post("login", adminAuthController.Login)
	})

	// 基础功能（需要认证和多语言，但不需要权限验证和操作日志）
	facades.Route().Prefix("admin").Middleware(middleware.Lang(), middleware.Jwt()).Group(func(router route.Router) {
		// 认证相关
		router.Get("info", adminAuthController.Info)
		router.Post("logout", adminAuthController.Logout)
	})

	// 需要认证、多语言、权限验证和操作日志的路由
	facades.Route().Prefix("admin").Middleware(middleware.Lang(), middleware.Jwt(), middleware.Permission(), middleware.OperationLog()).Group(func(router route.Router) {
		// 密码管理
		passwordController := admin.NewPasswordController()
		router.Put("password", passwordController.UpdatePassword)
		router.Put("admins/{id}/password", passwordController.ResetPassword)

		// 管理员管理
		router.Get("admins", adminController.Index)
		router.Get("admins/{id}", adminController.Show)
		router.Post("admins", adminController.Store)
		router.Put("admins/{id}", adminController.Update)
		router.Delete("admins/{id}", adminController.Destroy)

		// 角色管理
		router.Get("roles", roleController.Index)
		router.Get("roles/{id}", roleController.Show)
		router.Post("roles", roleController.Store)
		router.Put("roles/{id}", roleController.Update)
		router.Delete("roles/{id}", roleController.Destroy)

		// 权限管理
		router.Get("permissions", permissionController.Index)
		router.Get("permissions/{id}", permissionController.Show)
		router.Post("permissions", permissionController.Store)
		router.Put("permissions/{id}", permissionController.Update)
		router.Delete("permissions/{id}", permissionController.Destroy)

		// 菜单管理
		router.Get("menus", menuController.Index)
		router.Get("menus/{id}", menuController.Show)
		router.Post("menus", menuController.Store)
		router.Put("menus/{id}", menuController.Update)
		router.Delete("menus/{id}", menuController.Destroy)

		// 部门管理
		router.Get("departments", departmentController.Index)
		router.Get("departments/{id}", departmentController.Show)
		router.Post("departments", departmentController.Store)
		router.Put("departments/{id}", departmentController.Update)
		router.Delete("departments/{id}", departmentController.Destroy)

		// 字典管理
		router.Get("dictionaries", dictionaryController.Index)
		router.Get("dictionaries/{id}", dictionaryController.Show)
		router.Post("dictionaries", dictionaryController.Store)
		router.Put("dictionaries/{id}", dictionaryController.Update)
		router.Delete("dictionaries/{id}", dictionaryController.Destroy)
		router.Get("dictionaries/type/{type}", dictionaryController.GetByType)

		// 操作日志
		router.Get("operation-logs", operationLogController.Index)
		router.Get("operation-logs/{id}", operationLogController.Show)
		router.Delete("operation-logs/{id}", operationLogController.Destroy)
		router.Post("operation-logs/clean", operationLogController.Clean)

		// 登录日志
		router.Get("login-logs", loginLogController.Index)
		router.Get("login-logs/{id}", loginLogController.Show)
		router.Delete("login-logs/{id}", loginLogController.Destroy)
		router.Post("login-logs/clean", loginLogController.Clean)

		// 系统日志
		router.Get("system-logs", systemLogController.Index)
		router.Get("system-logs/{id}", systemLogController.Show)
		router.Delete("system-logs/{id}", systemLogController.Destroy)
		router.Post("system-logs/clean", systemLogController.Clean)
	})
}
