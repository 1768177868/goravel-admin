package middleware

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/logger"
)

// Permission 权限验证中间件
func Permission() http.Middleware {
	return func(ctx http.Context) {
		// 从context中获取admin信息（由JWT中间件设置）
		adminValue := ctx.Value("admin")
		if adminValue == nil {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    401,
				"message": trans.Get(ctx, "not_logged_in"),
			}).Abort()
			return
		}

		admin, ok := adminValue.(models.Admin)
		if !ok {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    401,
				"message": trans.Get(ctx, "not_logged_in"),
			}).Abort()
			return
		}

		// 加载管理员的角色、权限等关联数据
		adminService := services.NewAdminServiceImpl()
		if err := adminService.LoadRelationsWithPermissions(&admin); err != nil {
			logger.ErrorfHTTP(ctx, "permission middleware load relations failed: %v", err)
			_ = ctx.Response().Json(http.StatusInternalServerError, http.Json{
				"code":    500,
				"message": trans.Get(ctx, "load_permissions_failed"),
			}).Abort()
			return
		}

		// 检查是否是超级管理员（跳过权限检查）
		// 拥有 super-admin 角色的管理员（包括超级管理员和开发者管理员）都跳过权限检查
		for _, role := range admin.Roles {
			if role.Slug == "super-admin" && role.Status == 1 {
				ctx.Request().Next()
				return
			}
		}

		// 获取当前请求的方法和路径
		method := ctx.Request().Method()
		path := ctx.Request().Path()

		// 收集所有角色的权限（已通过预加载获取）
		var allPermissions []models.Permission
		for _, role := range admin.Roles {
			if role.Status == 1 {
				allPermissions = append(allPermissions, role.Permissions...)
			}
		}

		// 检查是否有权限
		hasPermission := false
		for _, perm := range allPermissions {
			if perm.Status == 1 {
				// 检查方法匹配
				if perm.Method == "" || perm.Method == method {
					// 检查路径匹配（支持通配符）
					if perm.Path == "" || perm.Path == path || matchPath(perm.Path, path) {
						hasPermission = true
						break
					}
				}
			}
		}

		if !hasPermission {
			_ = ctx.Response().Json(http.StatusForbidden, http.Json{
				"code":    403,
				"message": trans.Get(ctx, "no_permission"),
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}

// matchPath 简单的路径匹配，支持通配符
func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	// 简单的通配符匹配，如 /admin/admins/* 匹配 /admin/admins/1
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
