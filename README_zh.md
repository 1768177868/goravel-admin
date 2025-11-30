<p align="center"><img src="https://www.goravel.dev/logo.png?v=1.14.x" width="300"></p>

[English](./README.md) | 中文

## 关于 Goravel

Goravel 是一个功能完整、可扩展性良好的 Web 应用框架。作为起始脚手架，帮助 Gopher 快速构建自己的应用程序。

框架风格与 [Laravel](https://github.com/laravel/laravel) 保持一致，让 Phper 无需学习新框架，也能愉快地使用 Golang！致敬 Laravel！

欢迎 Star、PR 和 Issues！

## 后台管理系统

本项目包含一个基于 Goravel 框架构建的完整后台管理系统。

<p align="center"><img src="./admin.png" alt="后台管理系统截图" width="800"></p>

### 功能特性

#### 核心模块
- **认证与授权**
  - 基于 JWT 的认证
  - 基于角色的访问控制（RBAC）
  - 权限管理
  - 多令牌管理
  - 在线用户监控与踢出

- **用户管理**
  - 管理员用户管理
  - 部门管理
  - 角色管理
  - 权限分配
  - 密码重置

- **系统配置**
  - 菜单管理（动态菜单）
  - 字典管理
  - 系统配置
  - 黑名单管理

- **日志与监控**
  - 操作日志（自动记录）
  - 登录日志
  - 系统日志（带追踪 ID）
  - 服务监控

- **附加功能**
  - 数据统计仪表盘
  - 通知中心（WebSocket 实时通知）
  - 数据导出管理
  - 多语言支持（中文/英文）
  - 响应式 UI 设计

### 技术栈

**后端：**
- Goravel 框架（Go）
- JWT 认证
- RBAC 权限系统
- WebSocket 支持
- 数据库迁移与填充

**前端：**
- Vue 3
- Element Plus
- vxe-table（高级表格组件）
- Vue Router
- Pinia（状态管理）
- Axios
- ECharts（数据可视化）
- vue-i18n（国际化）

### 快速开始

1. **后端配置：**
   ```bash
   # 安装依赖
   go mod download
   
   # 在 .env 中配置数据库
   # 运行数据库迁移和填充
   go run . artisan migrate
   go run . artisan db:seed
   
   # 启动服务
   go run . --no-ansi
   # 或使用 air 进行热重载
   air
   ```

2. **前端配置：**
   ```bash
   cd html
   
   # 安装依赖
   npm install
   
   # 在 .env 中配置 API 地址
   # VITE_API_BASE_URL=http://127.0.0.1:3000
   # VITE_API_PREFIX=/api/admin
   
   # 启动开发服务器
   npm run dev
   ```

3. **默认登录：**
   - 用户名：`admin`
   - 密码：`admin123`
   - （首次登录后请修改默认密码）

### API 文档

后台管理 API 接口前缀为 `/api/admin`。除登录和验证码接口外，所有接口都需要 JWT 认证。

详细的 API 文档请查看 [routes/admin.go](./routes/admin.go)

### 项目结构

```
.
├── app/
│   ├── http/
│   │   ├── controllers/admin/    # 后台控制器
│   │   ├── middleware/           # 自定义中间件（JWT、权限、操作日志）
│   │   └── helpers/              # 辅助函数
│   ├── models/                   # 数据库模型
│   └── services/                 # 业务逻辑服务
├── routes/
│   └── admin.go                  # 后台路由
├── database/
│   ├── migrations/               # 数据库迁移
│   └── seeders/                  # 数据库填充
├── html/                         # 前端 Vue 应用
│   └── src/
│       ├── views/                # 页面组件
│       ├── components/           # 可复用组件
│       ├── api/                 # API 客户端
│       └── store/               # Pinia 状态管理
└── config/                       # 配置文件
```

### 安全特性

- 基于 JWT 令牌的认证
- 权限中间件保护路由
- 自动操作日志记录
- 日志中敏感数据过滤
- 登录接口限流
- IP/用户黑名单管理
- 令牌撤销支持

## 快速入门

### 启动服务

`go run . --no-ansi` 或 `air`

[关于 air]：https://www.goravel.dev/getting-started/installation.html#live-reload

### 数据库

[app/http/controllers/db_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/db_controller.go)

### WebSocket

[app/http/controllers/websocket_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/websocket_controller.go)

### 数据验证

[app/http/controllers/validation_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/validation_controller.go)

### 认证

[app/http/controllers/auth_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/auth_controller.go)

### 单元测试（使用 Mock 测试）

[app/http/controllers/validation_controller_test.go](https://github.com/goravel/example/blob/master/app/http/controllers/validation_controller_test.go)

### 集成测试（使用配置测试）

[tests/controllers/validation_controller_test.go](https://github.com/goravel/example/blob/master/tests/controllers/validation_controller_test.go)

### GRPC

[app/grpc/controllers/user_controller.go](https://github.com/goravel/example/blob/master/app/grpc/controllers/user_controller.go)

### Swagger（适用于 gin HTTP 驱动）

[app/http/controllers/swagger_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/swagger_controller.go)

### 将单页应用集成到框架中

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L26)

### 视图嵌套

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L33)

### Session

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L42)

### Cookie

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L58)

### 本地化

[routes/api.go](https://github.com/goravel/example/blob/master/routes/api.go#L37)

### GraphQL

```bash
# 在本地下载并安装 gqlgen，只需运行一次
go get -d github.com/99designs/gqlgen
# 重新生成代码
go run github.com/99designs/gqlgen generate
```

```
swag init
go run .
http://localhost:3000/swagger/
```

cloudflare pages
```
# 构建命令
npm install && npm run build
# 部署命令
npx wrangler deploy --assets ./dist --compatibility-date 2025-11-29 --name admin
# 根目录
html
```

## 文档

在线文档 [https://www.goravel.dev](https://www.goravel.dev)

> 要优化文档，请向文档仓库提交 PR
> [https://github.com/goravel/docs](https://github.com/goravel/docs)

## 社区

欢迎在 Discord 中讨论。

[https://discord.gg/cFc5csczzS](https://discord.gg/cFc5csczzS)

## 许可证

Goravel 框架是在 [MIT 许可证](https://opensource.org/licenses/MIT) 下发布的开源软件。


