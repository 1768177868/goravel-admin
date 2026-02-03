package models

import (
	"github.com/goravel/framework/contracts/database/orm"
)

// TenantScope 租户全局作用域工厂函数（预留）
// 注意：由于 GlobalScopes 的函数签名是 func(orm.Query) orm.Query，无法直接访问 HTTP Context
// 因此推荐在 TenantQueryService 中手动应用租户过滤，而不是在模型的 GlobalScopes 中定义
//
// 如果需要在模型上定义 GlobalScopes，可以使用以下方式：
// func (r *User) GlobalScopes() map[string]func(orm.Query) orm.Query {
//     return map[string]func(orm.Query) orm.Query{
//         "tenant": func(query orm.Query) orm.Query {
//             // 注意：这里无法访问 HTTP Context，需要通过其他方式获取租户ID
//             // 推荐使用 TenantQueryService 来处理租户过滤
//             return query
//         },
//     }
// }
//
// 推荐方案：在 TenantQueryService 中手动应用租户过滤，并使用 WithoutGlobalScopes("tenant") 排除
func TenantScope() func(orm.Query) orm.Query {
	// 由于无法访问 HTTP Context，这里返回一个空作用域
	// 实际的租户过滤应该在 TenantQueryService 中处理
	return func(query orm.Query) orm.Query {
		return query
	}
}
