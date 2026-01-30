# 租户功能预留说明

当前已预留租户相关代码，**未启用多租户时不会对现有查询产生任何影响**。后续需要多租户时按以下步骤接入即可。

## 已预留文件

| 文件 | 说明 |
|------|------|
| `app/http/helpers/tenant.go` | 租户辅助函数：GetTenantIDFromContext、ScopeTenant、ScopeTenantOrGlobal 等 |
| `app/services/tenant_query_service.go` | 租户查询服务：统一查询入口，自动按租户过滤 |

## 当前行为

- `GetTenantIDFromContext(ctx)`：未设置租户时返回 `(0, false)`。
- `ScopeTenant(ctx, query)`：未设置租户时直接返回原 `query`，不添加任何条件。
- `NewTenantQueryService(ctx).QueryModel(&User{})`：等价于 `facades.Orm().Query().Model(&User{})`。

因此现有代码无需修改，也不会被租户逻辑影响。

## 日后启用多租户时的步骤

1. **数据库**
   - 新增 `tenants` 表（租户主表）。
   - 在需要隔离的业务表上增加 `tenant_id` 字段并建索引。

2. **中间件**
   - 在 JWT 之后增加租户中间件：根据当前管理员解析出 `tenant_id`，并执行：
     - `ctx.WithValue("tenant_id", tenantID)`
   - 超级管理员可约定 `tenant_id == 0` 或不在 context 中设置，表示不过滤租户。

3. **查询方式二选一**
   - **方式 A**：在需要按租户过滤的地方，用 `helpers.ScopeTenant(ctx, query)` 包装原有 query。
   - **方式 B（推荐）**：统一使用 `NewTenantQueryService(ctx)` 的 `Query()` / `QueryModel()` / `QueryTable()`，由服务内部自动加租户条件。

4. **写入**
   - 创建业务数据时，从 `helpers.GetTenantIDFromContext(ctx)` 或 `TenantQueryService.GetTenantID()` 取租户 ID，写入模型的 `tenant_id` 字段。

## 使用示例（启用租户后）

```go
// 控制器中
tenantQuery := services.NewTenantQueryService(ctx)
var users []models.User
tenantQuery.QueryModel(&models.User{}).Where("status", 1).Find(&users)

// 或继续用 facades.Orm，手动加 ScopeTenant
query := facades.Orm().Query().Model(&models.User{})
query = helpers.ScopeTenant(ctx, query)
query.Find(&users)
```

当前仅做代码预留，无需改表或改业务逻辑；等需要租户时再按上述步骤接入即可。
