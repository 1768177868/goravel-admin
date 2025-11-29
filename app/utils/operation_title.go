package utils

import (
	"github.com/goravel/framework/contracts/http"
)

// GetOperationTitleFromContext 从 context 中获取操作标题
// 只使用权限标识（permission_slug），直接返回权限标识，前端会使用多语言翻译
// 如果没有权限标识，则返回 operation.unknown
func GetOperationTitleFromContext(ctx http.Context) string {
	if ctx == nil {
		return "operation.unknown"
	}

	// 优先从 context 中获取权限标识（由权限中间件设置）
	permissionSlugValue := ctx.Value("permission_slug")
	if permissionSlugValue != nil {
		if permissionSlug, ok := permissionSlugValue.(string); ok && permissionSlug != "" {
			// 直接返回权限标识，前端多语言文件中已有对应翻译
			return permissionSlug
		}
	}

	// 没有权限标识时，统一返回未知操作
	return "operation.unknown"
}
