# GORM Sharding 配置说明

## 概述

用户余额变动记录表（`user_balance_logs`）使用 GORM Sharding 插件进行分表，基于 `user_id` 字段进行哈希分表。

## 配置步骤

### 1. 安装依赖

已安装 `gorm.io/sharding` 插件。

### 2. 配置 GORM Sharding

**✅ 配置已完成**

GORM Sharding 插件已在 `app/providers/database_service_provider.go` 中自动配置。应用启动时会自动初始化。

配置详情：
- 通过反射从 Goravel ORM 获取原生 GORM DB 实例
- 自动注册 `user_balance_logs` 表的分表配置
- 如果配置失败，会记录警告日志但不阻止应用启动（数据会写入单表）

配置代码位置：`app/providers/database_service_provider.go` 的 `initGormSharding()` 方法。

### 3. 分表配置参数

- **ShardingKey**: `user_id` - 分表键，所有查询必须包含此字段
- **NumberOfShards**: `64` - 分表数量（建议为 2 的幂次）
- **PrimaryKeyGenerator**: `sharding.PKSnowflake` - 主键生成器（确保全局唯一）

### 4. 使用注意事项

1. **所有查询必须包含 `user_id`**：GORM Sharding 需要 ShardingKey 来路由到正确的分表
2. **不支持跨分表查询**：如果需要查询多个用户的数据，需要分别查询后合并
3. **分表自动创建**：首次插入数据时，插件会自动创建对应的分表

### 5. 分表命名规则

分表命名格式：`user_balance_logs_{shard_index}`

例如：
- `user_balance_logs_0`
- `user_balance_logs_1`
- ...
- `user_balance_logs_63`

### 6. 验证配置

创建一条余额变动记录，检查是否正确路由到分表：

```go
log := &models.UserBalanceLog{
    UserID: 123,
    Type: "income",
    Amount: 100.00,
    Balance: 100.00,
    // ...
}
facades.Orm().Query().Create(log)
```

检查数据库中是否创建了对应的分表。

## 故障排查

1. **如果分表未创建**：
   - 检查应用启动日志，查看是否有 "GORM Sharding 配置成功" 的日志
   - 如果有 "GORM Sharding 配置失败" 的警告，说明无法获取原生 GORM DB 实例
   - 此时数据会写入单表 `user_balance_logs`，不会分表
   - 可以通过查看数据库表结构确认是否创建了分表（如 `user_balance_logs_0`, `user_balance_logs_1` 等）

2. **如果查询报错 "missing sharding key"**：
   - 确保所有查询条件中都包含 `user_id` 字段
   - GORM Sharding 需要 ShardingKey 来路由到正确的分表

3. **如果无法获取原生 GORM DB**：
   - 检查 `app/utils/gorm_sharding.go` 中的 `GetGormDB()` 方法
   - 可能需要根据 Goravel 框架的实际实现调整反射逻辑
   - 如果反射失败，可以考虑直接创建 GORM 连接（使用相同的数据库配置）

## 参考文档

- [GORM Sharding 官方文档](https://gorm.io/zh_CN/docs/sharding.html)
- [GORM Sharding GitHub](https://github.com/go-gorm/sharding)

