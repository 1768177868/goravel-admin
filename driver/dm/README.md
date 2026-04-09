# goravel dm driver

DM (达梦) driver for Goravel DB/ORM.

## 说明

本驱动用于将达梦数据库接入 Goravel 的 `database` 模块，接入风格与 `goravel/mysql`、`goravel/postgres` 保持一致，支持通过 `config/database.go` 的 `via` 方式注入。

## 当前目录结构

- `service_provider.go`: ServiceProvider 注册
- `facades/dm.go`: `Dm(connection)` facade
- `config.go`: 读取并组装 `database.connections.dm` 配置
- `dm_driver.go`: 实现 Goravel `driver.Driver` 接口
- `dialector.go`: 适配底层 GORM Dialector
- `grammar.go`: 独立 Grammar（不依赖 postgres 包）
- `processor.go`: 独立 Processor（不依赖 postgres 包）
- `contracts/config.go`: 驱动配置结构定义

## 使用方式

### 1) 注册 Provider

在 `bootstrap/providers.go` 中注册:

```go
&dm.ServiceProvider{},
```

### 2) 配置 database 连接

在 `config/database.go` 中新增 `dm` 连接，例如:

```go
"dm": map[string]any{
  "host":      config.Env("DB_HOST", "127.0.0.1"),
  "port":      config.Env("DB_PORT", 5236),
  "database":  config.Env("DB_DATABASE", "SYSDBA"),
  "username":  config.Env("DB_USERNAME", "SYSDBA"),
  "password":  config.Env("DB_PASSWORD", "SYSDBA"),
  "schema":    config.Env("DB_SCHEMA", "SYSDBA"),
  "dsn":       config.Env("DB_DSN", ""),
  "gorm_mode": config.Env("DB_GORM_MODE", 0), // dm兼容=0; mysql兼容=1
  "prefix":    "",
  "singular":  false,
  "via": func() (driver.Driver, error) {
    return dmfacades.Dm("dm")
  },
},
```

### 3) 设置默认连接

在 `.env` 中设置:

```env
DB_CONNECTION=dm
```

## 构建开关（重要）

为避免未安装达梦驱动时影响项目默认编译，本驱动使用了 Go build tags：

- 默认构建（不带 tag）：`driver/dm` 会编译通过，但连接 DM 时会返回明确错误提示
- 启用 DM 构建：使用 `-tags dm`

示例：

```bash
go build -tags dm ./...
go test -tags dm ./driver/dm/...
```

集成测试（需要可访问的达梦实例）：

```bash
set DM_TEST_DSN=SYSDBA:SYSDBA@127.0.0.1:5236
go test -tags dm ./driver/dm/... -run TestDMCrudAndTransaction -v
```

当前该集成测试覆盖：

- CRUD（增查改删）
- 事务回滚与事务提交
- 唯一约束冲突

## 依赖说明

底层方言依赖达梦官方 Go 驱动包 `dm`。若本地未安装，会出现 `package dm is not in std` 编译错误。

通常需要按达梦官方文档安装驱动，并在 `go.mod` 中通过 `replace dm = ...` 指向本地驱动路径（如有需要）。

例如：

```go
replace dm => ../dm8/drivers/go/dm
```

### 项目已验证示例（Windows）

如果达梦安装在 `E:/dmdbms`，可按以下方式配置并验证：

```go
replace dm => E:/dmdbms/drivers/go/dm
```

```bash
go build -tags dm ./...
go test -tags dm ./driver/dm/... -run TestDMCrudAndTransaction -v
go run -tags dm .
```

如果 `E:/dmdbms/drivers/go/dm` 不存在，请先解压：

- `E:/dmdbms/drivers/go/dm-go-driver.zip` -> `E:/dmdbms/drivers/go/dm`

## 备注

当前 `driver/dm` 已包含完整方言实现（`dialector.go`、`create.go`、`migrator.go`），不再依赖 `driver/dm-test`。
