# 多租户迁移到 GlobalScopes 指南

> **Goravel v1.18**：推荐 `appfacades.OrmQuery(ctx)` 替代 `facades.Orm().Query().WithContext(ctx)`，与 Service `NewXxxService(ctx)` 模式一致。下文部分示例仍保留历史写法，新代码请统一使用 `OrmQuery(ctx)`。

## 概述

Goravel v1.17 引入了 `GlobalScopes` 和 `WithoutGlobalScopes` 功能，可以将租户过滤作为全局作用域实现，使代码更优雅、更自动化。

## 优势对比

### 旧方式（手动 ScopeTenant）
```go
// 需要在每个查询处手动添加（v1.18 推荐 OrmQuery）
query := appfacades.OrmQuery(ctx).Model(&models.User{})
query = helpers.ScopeTenant(ctx, query)
query.Find(&users)

// 或使用 TenantQueryService
tenantQuery := services.NewTenantQueryService(ctx)
tenantQuery.QueryModel(&models.User{}).Find(&users)
```

### 新方式（GlobalScopes）
```go
// 自动应用租户过滤，无需手动调用（v1.18）
var users []models.User
appfacades.OrmQuery(ctx).Model(&models.User{}).Find(&users)

// 需要排除租户过滤时（如超级管理员查看所有数据）
appfacades.OrmQuery(ctx).Model(&models.User{}).WithoutGlobalScopes("tenant").Find(&users)
```

## 迁移步骤

### 1. 为需要租户隔离的模型添加 GlobalScopes

在模型的 `GlobalScopes()` 方法中返回租户作用域：

```go
package models

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
)

type User struct {
	orm.Model
	TenantID uint `gorm:"index;comment:租户ID" json:"tenant_id"`
	// ... 其他字段
}

func (r *User) GlobalScopes() map[string]func(orm.Query) orm.Query {
	return map[string]func(orm.Query) orm.Query{
		"tenant": func(query orm.Query) orm.Query {
			// 从 context 中获取租户ID
			// 注意：需要在查询时传递 http.Context
			tenantID, ok := helpers.GetTenantIDFromContext(ctx)
			if !ok || tenantID == 0 {
				return query
			}
			return query.Where("tenant_id", tenantID)
		},
	}
}
```

### 2. 在查询时传递 HTTP Context

由于全局作用域需要访问 HTTP Context 来获取租户ID，需要在查询时传递：

```go
// 方式1：通过 OrmQuery 传递 context（v1.18 推荐）
appfacades.OrmQuery(ctx).Model(&models.User{}).Find(&users)

// 方式2：历史写法（仍可用）
facades.Orm().Query().WithContext(ctx).Model(&models.User{}).Find(&users)
```

### 3. 使用 WithoutGlobalScopes 排除租户过滤

当需要查询所有租户的数据时（如超级管理员）：

```go
// 排除所有全局作用域
facades.Orm().Query().Model(&models.User{}).WithoutGlobalScopes().Find(&users)

// 只排除租户作用域
facades.Orm().Query().Model(&models.User{}).WithoutGlobalScopes("tenant").Find(&users)
```

### 4. 更新现有代码

#### 替换 TenantQueryService

**旧代码：**
```go
tenantQuery := services.NewTenantQueryService(ctx)
var users []models.User
tenantQuery.QueryModel(&models.User{}).Find(&users)
```

**新代码：**
```go
var users []models.User
facades.Orm().Query().WithContext(ctx).Model(&models.User{}).Find(&users)
```

#### 替换手动 ScopeTenant

**旧代码：**
```go
query := facades.Orm().Query().Model(&models.User{})
query = helpers.ScopeTenant(ctx, query)
query.Find(&users)
```

**新代码：**
```go
var users []models.User
facades.Orm().Query().WithContext(ctx).Model(&models.User{}).Find(&users)
```

## 注意事项

1. **Context 传递**：全局作用域需要访问 HTTP Context，确保在查询时传递了 context
2. **向后兼容**：迁移时可以保留 `TenantQueryService` 和 `ScopeTenant` 作为过渡方案
3. **性能**：全局作用域会在所有查询中自动应用，确保只在需要租户隔离的模型上使用
4. **超级管理员**：使用 `WithoutGlobalScopes("tenant")` 可以让超级管理员查看所有租户的数据

## 示例：完整的模型实现

```go
package models

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/helpers"
)

type User struct {
	orm.Model
	TenantID uint   `gorm:"index;comment:租户ID" json:"tenant_id"`
	Username string `gorm:"not null;size:50;uniqueIndex" json:"username"`
	// ... 其他字段
}

func (r *User) GlobalScopes() map[string]func(orm.Query) orm.Query {
	return map[string]func(orm.Query) orm.Query{
		"tenant": func(query orm.Query) orm.Query {
			// 从 query 的 context 中获取 http.Context
			// 注意：这需要在查询时通过 WithContext 传递
			ctx := query.GetContext()
			if httpCtx, ok := ctx.Value("http_context").(http.Context); ok {
				tenantID, ok := helpers.GetTenantIDFromContext(httpCtx)
				if ok && tenantID > 0 {
					return query.Where("tenant_id", tenantID)
				}
			}
			return query
		},
	}
}
```

## 推荐方案：结合使用

考虑到 Goravel 的 GlobalScopes 函数签名是 `func(orm.Query) orm.Query`，无法直接访问 HTTP Context，推荐以下混合方案：

### 方案：改进 TenantQueryService，内部使用 GlobalScopes

**更新 TenantQueryService：**

```go
package services

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
)

type TenantQueryService struct {
	ctx http.Context
}

func NewTenantQueryService(ctx http.Context) *TenantQueryService {
	return &TenantQueryService{ctx: ctx}
}

// QueryModel 创建模型查询，自动应用租户过滤
func (s *TenantQueryService) QueryModel(model any) orm.Query {
	query := facades.Orm().Query().Model(model)
	
	// 如果模型实现了 GlobalScopes，先应用其他作用域
	// 然后手动添加租户过滤（因为 GlobalScopes 无法访问 HTTP Context）
	tenantID, ok := helpers.GetTenantIDFromContext(s.ctx)
	if ok && tenantID > 0 {
		query = query.Where("tenant_id", tenantID)
	}
	
	return query
}

// QueryModelWithoutTenant 创建模型查询，排除租户过滤（用于超级管理员）
func (s *TenantQueryService) QueryModelWithoutTenant(model any) orm.Query {
	// 使用 WithoutGlobalScopes 排除租户作用域（如果模型定义了）
	query := facades.Orm().Query().Model(model).WithoutGlobalScopes("tenant")
	return query
}
```

### 使用示例

```go
// 普通查询（自动应用租户过滤）
tenantQuery := services.NewTenantQueryService(ctx)
var users []models.User
tenantQuery.QueryModel(&models.User{}).Find(&users)

// 超级管理员查询（排除租户过滤）
if helpers.IsSuperAdmin(ctx) {
	var allUsers []models.User
	tenantQuery.QueryModelWithoutTenant(&models.User{}).Find(&allUsers)
}
```

### 优势

1. **保持现有 API**：`TenantQueryService` 的接口不变，现有代码无需修改
2. **利用新特性**：可以使用 `WithoutGlobalScopes("tenant")` 排除租户过滤
3. **灵活性**：可以根据需要选择是否应用租户过滤
4. **向后兼容**：未启用多租户时，`GetTenantIDFromContext` 返回 `(0, false)`，不会添加任何条件

## 总结

虽然 Goravel 的 GlobalScopes 无法直接访问 HTTP Context，但通过改进 `TenantQueryService`，我们仍然可以：

1. **保持代码简洁**：继续使用 `TenantQueryService` 封装查询逻辑
2. **利用新特性**：使用 `WithoutGlobalScopes("tenant")` 让超级管理员查看所有数据
3. **向后兼容**：现有代码无需修改，平滑迁移

这样既利用了 v1.17 的新特性，又保持了代码的灵活性和可维护性。
