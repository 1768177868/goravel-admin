# goravel dm driver

DM (达梦) driver for Goravel DB/ORM.

## 说明

本驱动用于将达梦数据库接入 Goravel 的 `database` 模块，接入风格与 `goravel/mysql`、`goravel/postgres` 保持一致，支持通过 `config/database.go` 的 `via` 方式注入。

**本仓库（monorepo）**中源码路径为 `driver/dm`，包导入前缀为 `goravel/driver/dm/...`（与当前工程 `go.mod` 的 `module` 名一致）。

### 独立仓库与 `go get`（可选）

若使用拆出去的镜像仓库 **[wangxuancheng-dev/goravel-dm](https://github.com/wangxuancheng-dev/goravel-dm)**，**不要**在 `go get` 里写 `https://`；应使用 **模块路径**：

```bash
go get github.com/wangxuancheng-dev/goravel-dm@latest
```

也可指定分支或伪版本，例如 `go get github.com/wangxuancheng-dev/goravel-dm@main`。

拉取后还需满足：

1. 该仓库根目录有 **`go.mod`**，且第一行 `module` 与上面路径一致（例如 `module github.com/wangxuancheng-dev/goravel-dm`），否则 `go get` 无法作为依赖使用。
2. 业务代码里的 **`import` 路径**要改为该模块路径下的包（例如 `facades` 子目录），与当前 monorepo 里的 `goravel/driver/dm/facades` 不同，需要一并替换。

**达梦官方驱动 `dm`** 仍须按下文「依赖说明」单独 `require` + `replace` 到本地驱动目录；`go get github.com/.../goravel-dm` **不能**替代官方 `dm` 包。

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
  // 非空时自动 DSN 追加 /模式名；留空则不拼路径（用登录用户默认模式，避免把 DB_DATABASE 里的实例名误写入 URL）
  "schema":    config.Env("DB_SCHEMA", ""),
  // 可选。留空时由 host/port/username/password 与可选 schema 自动拼成 dm://... DSN
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
# 可不写 DB_DSN：默认生成 dm://user:pass@host:port（无尾部路径，使用登录用户默认模式）
# 若必须在 URL 里指定模式：DB_SCHEMA=SYSDBA（须为库中真实存在的模式名，不要填实例名 DAMENG）
# DB_DSN=dm://SYSDBA:yourpass@127.0.0.1:5236/SYSDBA
```

### 错误 -2103「无效的模式名」

常见原因是 DSN 末尾的 **模式名** 写成了 **实例/服务名**（如 `DAMENG`），而达梦只接受 **已创建的模式**。自动 DSN **不再**把 `DB_DATABASE` 拼进 URL，因此仅配 `DB_DATABASE=DAMENG` 一般不会触发该问题；若手写 `DB_DSN` 或设置了 `DB_SCHEMA`，请改为真实模式（如 `SYSDBA`）。

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
set DM_TEST_DSN=dm://SYSDBA:SYSDBA@127.0.0.1:5236
go test -tags dm ./driver/dm/... -run TestDMCrudAndTransaction -v
```

当前该集成测试覆盖：

- CRUD（增查改删）
- 事务回滚与事务提交
- 唯一约束冲突

## 依赖说明

底层方言依赖达梦官方 Go 驱动模块 **`dm`**（`import _ "dm"`）。未正确声明依赖时，带 `-tags dm` 编译会出现 `package dm is not in std` 或 `missing go.sum entry` 等错误。

达梦驱动**通常不在**公网 `proxy.golang.org`，需要本地解压官方提供的 Go 驱动源码，并用 **`replace`** 指到该目录。

### 在项目中声明 `dm` 依赖（推荐流程）

1. 从达梦安装介质解压得到驱动目录（内含 `go.mod`、模块名一般为 `dm`），例如 `.../drivers/go/dm`。
2. 在项目根目录执行（二选一即可）：

   **方式 A — 先拉 require，再 replace（与当前仓库一致）**

   ```bash
   go get dm
   ```

   随后在 `go.mod` 中增加 **`replace dm => <本地驱动目录>`**（或执行 `go mod edit -replace=dm=<路径>`），再 `go mod tidy`。仅 `go get` 而不 `replace` 时，多数环境仍无法得到真实源码。

   若公网无法解析 `dm`，改用方式 B。

   **方式 B — 直接写 `go.mod`**

   在 `go.mod` 中增加占位版本（满足 `replace` 要求即可），例如：

   ```go
   require dm v0.0.0-00010101000000-000000000000

   replace dm => /绝对或相对路径/达梦驱动/dm
   ```

   然后执行：

   ```bash
   go mod tidy
   ```

3. 启用 DM 时再编译/测试：

   ```bash
   go build -tags dm ./...
   ```

`replace` 路径示例：

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
