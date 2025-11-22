# JWT Token 配置说明

## .env 配置

在 `.env` 文件中添加以下配置：

```env
# JWT 配置
JWT_SECRET=your-jwt-secret-key-here
JWT_TTL=60
JWT_REFRESH_TTL=20160
```

### 配置说明

- `JWT_SECRET`: JWT签名密钥，用于加密token。可以通过 `go run . artisan jwt:secret` 生成
- `JWT_TTL`: Token有效期（分钟），默认60分钟（1小时）
  - 设置为 `0` 表示永久有效（不推荐，除非特殊需求）
  - 建议值：60（1小时）、120（2小时）、1440（24小时）
- `JWT_REFRESH_TTL`: Token刷新窗口（分钟），默认20160分钟（14天）
  - 在此时间窗口内，过期的token可以刷新
  - 设置为 `0` 表示无限刷新时间（不推荐）

## 永久Token配置

系统支持为特定用户配置永久有效的token。

### 数据库配置

1. 运行数据库迁移：
```bash
go run . artisan migrate
```

2. 在 `admins` 表中，将用户的 `token_never_expires` 字段设置为 `1`（true）：
```sql
UPDATE admins SET token_never_expires = 1 WHERE id = 1;
```

或者通过后台管理界面设置。

### 功能说明

- **永久Token用户**：
  - Token不会过期
  - 即使token过期，系统也会自动重新生成
  - 不进行滑动过期处理（因为不需要）

- **普通Token用户**：
  - Token按照 `JWT_TTL` 配置过期
  - 支持滑动过期（每次请求自动延长过期时间）
  - Token过期后需要重新登录

### 安全建议

1. **谨慎使用永久Token**：只给需要长期访问的系统用户或API用户配置
2. **定期审查**：定期检查哪些用户配置了永久Token
3. **监控异常**：监控永久Token的使用情况，发现异常及时处理
4. **结合其他安全措施**：即使使用永久Token，也要结合IP白名单、操作日志等安全措施

## 滑动过期机制

系统实现了滑动过期机制：

- 每次请求时，如果token有效，会自动生成新的token（延长过期时间）
- 这样活跃用户不会因为token过期而意外退出
- 长期不活跃的用户，token过期后需要重新登录

## 示例配置

### 开发环境
```env
JWT_TTL=1440          # 24小时
JWT_REFRESH_TTL=10080 # 7天
```

### 生产环境
```env
JWT_TTL=60            # 1小时
JWT_REFRESH_TTL=20160 # 14天
```

### 高安全环境
```env
JWT_TTL=30            # 30分钟
JWT_REFRESH_TTL=1440  # 1天
```

