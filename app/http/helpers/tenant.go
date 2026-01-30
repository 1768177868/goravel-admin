package helpers

import (
	"errors"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
)

// 租户相关辅助函数（预留）
// 启用多租户时：在中间件中设置 ctx.WithValue("tenant_id", tenantID)，并为业务表添加 tenant_id 字段
// 当前未启用时：GetTenantIDFromContext 返回 (0, false)，ScopeTenant 直接返回原 query，不影响现有逻辑

const contextKeyTenantID = "tenant_id"

// GetTenantIDFromContext 从 context 中获取租户ID
// 返回: (tenantID, exists)。未启用多租户或未设置时 exists 为 false
func GetTenantIDFromContext(ctx http.Context) (uint, bool) {
	if ctx == nil {
		return 0, false
	}
	v := ctx.Value(contextKeyTenantID)
	if v == nil {
		return 0, false
	}
	switch id := v.(type) {
	case uint:
		return id, id > 0
	case uint8:
		return uint(id), id > 0
	case uint16:
		return uint(id), id > 0
	case uint32:
		return uint(id), id > 0
	case uint64:
		return uint(id), id > 0
	case int:
		if id > 0 {
			return uint(id), true
		}
	case int64:
		if id > 0 {
			return uint(id), true
		}
	}
	return 0, false
}

// ScopeTenant 为查询添加租户范围过滤（预留）
// 未启用多租户时直接返回原 query；启用后当 context 中有 tenant_id 时自动添加 WHERE tenant_id = ?
func ScopeTenant(ctx http.Context, query orm.Query) orm.Query {
	if ctx == nil {
		return query
	}
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok || tenantID == 0 {
		return query
	}
	return query.Where("tenant_id", tenantID)
}

// ScopeTenantOrGlobal 为查询添加租户范围，并允许查询全局数据 tenant_id = 0（预留）
// 适用于配置表等需要“租户配置 + 全局默认”的场景
func ScopeTenantOrGlobal(ctx http.Context, query orm.Query) orm.Query {
	if ctx == nil {
		return query
	}
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok || tenantID == 0 {
		return query.Where("tenant_id", 0)
	}
	return query.Where("tenant_id = ? OR tenant_id = ?", tenantID, 0)
}

// ScopeTenantStrict 严格租户过滤：无租户ID时返回空结果（预留）
func ScopeTenantStrict(ctx http.Context, query orm.Query) orm.Query {
	if ctx == nil {
		return query.Where("1 = 0")
	}
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok || tenantID == 0 {
		return query.Where("1 = 0")
	}
	return query.Where("tenant_id", tenantID)
}

// IsSuperAdmin 判断当前请求是否为“超级管理员”（无租户ID，预留）
func IsSuperAdmin(ctx http.Context) bool {
	if ctx == nil {
		return false
	}
	tenantID, ok := GetTenantIDFromContext(ctx)
	return !ok || tenantID == 0
}

// RequireTenant 要求必须有租户ID，否则返回错误（预留）
func RequireTenant(ctx http.Context) (uint, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok || tenantID == 0 {
		return 0, errors.New("tenant_id is required")
	}
	return tenantID, nil
}
