package utils

import (
	"context"
	"reflect"

	appfacades "goravel/app/facades"

	"github.com/goravel/framework/contracts/database/orm"
)

// UniqueReusePolicy 软删除后唯一键是否可复用。
type UniqueReusePolicy bool

const (
	// UniqueReuseAllow 软删后可复用：仅与未软删记录冲突（Model 查询）。
	// 无软删表删行后自然可复用，行为与 Model 查询一致。
	UniqueReuseAllow UniqueReusePolicy = true
	// UniqueReuseDeny 软删后仍占位：与含软删记录冲突（裸表查询）。
	UniqueReuseDeny UniqueReusePolicy = false
)

// 各模块业务约定：
//   - users: Allow（注销后可再注册）
//   - admins, payment_methods, currencies: Deny（审计/配置代码永久占位）
//   - menus, roles, permissions: Allow + 硬删（删除后 slug/name 可复用）
//   - orders, payments: Deny（单号永不重复，创建逻辑保证）

func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == ""
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Pointer && rv.IsNil() {
			return true
		}
		return false
	}
}

func buildUniqueQuery(ctx context.Context, tableName string, model any, policy UniqueReusePolicy) orm.Query {
	if policy == UniqueReuseAllow && model != nil {
		return appfacades.OrmQuery(ctx).Model(model)
	}
	return appfacades.OrmQuery(ctx).Table(tableName)
}

// ExistsColumnValue 按策略检查某列值是否已存在。
func ExistsColumnValue(ctx context.Context, tableName string, model any, policy UniqueReusePolicy, column string, value any, excludeID uint) (bool, error) {
	if isEmptyValue(value) {
		return false, nil
	}

	query := buildUniqueQuery(ctx, tableName, model, policy).Where(column, value)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	return query.Exists()
}

// ExistsColumnValueAny 按策略检查多列 OR 是否已存在。
func ExistsColumnValueAny(ctx context.Context, tableName string, model any, policy UniqueReusePolicy, excludeID uint, orEquals map[string]any) (bool, error) {
	if len(orEquals) == 0 {
		return false, nil
	}

	query := buildUniqueQuery(ctx, tableName, model, policy)
	first := true
	for column, value := range orEquals {
		if isEmptyValue(value) {
			continue
		}
		if first {
			query = query.Where(column, value)
			first = false
			continue
		}
		query = query.OrWhere(column, value)
	}
	if first {
		return false, nil
	}
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	return query.Exists()
}
