# 系统架构文档

本文档描述 Goravel Admin 后台管理系统的整体架构设计。

## 目录

- [系统概览](#系统概览)
- [技术架构](#技术架构)
- [后端架构](#后端架构)
- [前端架构](#前端架构)
- [数据库设计](#数据库设计)
- [安全架构](#安全架构)
- [部署架构](#部署架构)

---

## 系统概览

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端 (Browser)                         │
├─────────────────────────────────────────────────────────────────┤
│                    Vue 3 SPA (Element Plus)                      │
├─────────────────────────────────────────────────────────────────┤
│                         Nginx / CDN                              │
├─────────────────────────────────────────────────────────────────┤
│                      Goravel API Server                          │
├─────────────────────────────────────────────────────────────────┤
│           MySQL / PostgreSQL          │         Redis            │
└─────────────────────────────────────────────────────────────────┘
```

---

## 技术架构

### 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| **后端框架** | Goravel | v1.14+ |
| **编程语言** | Go | 1.21+ |
| **前端框架** | Vue | 3.4+ |
| **UI 组件** | Element Plus | 2.4+ |
| **表格组件** | VXE-Table | 4.7+ |
| **状态管理** | Pinia | 2.1+ |
| **数据库** | MySQL/PostgreSQL | 8.0+ / 15+ |
| **缓存** | Redis | 7.0+ |
| **认证** | JWT | - |

### 架构模式

```
┌─────────────────────────────────────────────────────────────────┐
│                          前端 (Vue 3)                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────────┐ │
│  │  Views   │  │Components│  │Composables│ │    Store (Pinia) │ │
│  └────┬─────┘  └────┬─────┘  └────┬──────┘ └────────┬─────────┘ │
│       └─────────────┴─────────────┴─────────────────┘           │
│                              │ API                               │
├──────────────────────────────┼──────────────────────────────────┤
│                          后端 (Goravel)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │  Controllers │  │  Middleware  │  │      Routes          │   │
│  └──────┬───────┘  └──────┬───────┘  └──────────────────────┘   │
│         │                 │                                      │
│  ┌──────▼───────┐  ┌──────▼───────┐                             │
│  │   Services   │  │   Helpers    │                             │
│  └──────┬───────┘  └──────────────┘                             │
│         │                                                        │
│  ┌──────▼───────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │    Models    │  │    Events    │  │       Jobs           │   │
│  └──────┬───────┘  └──────────────┘  └──────────────────────┘   │
├─────────┼───────────────────────────────────────────────────────┤
│         ▼                                                        │
│  ┌──────────────┐  ┌──────────────┐                             │
│  │   Database   │  │    Redis     │                             │
│  └──────────────┘  └──────────────┘                             │
└─────────────────────────────────────────────────────────────────┘
```

---

## 后端架构

### 目录结构

```
app/
├── console/                 # 命令行
│   └── commands/           # 自定义命令
├── http/
│   ├── controllers/        # 控制器
│   │   └── admin/         # 后台管理控制器
│   ├── middleware/         # 中间件
│   ├── requests/           # 请求验证
│   ├── helpers/            # 辅助函数
│   ├── response/           # 统一响应
│   └── trans/              # 翻译工具
├── models/                  # 数据模型
├── services/                # 业务逻辑
├── utils/                   # 工具函数
│   ├── logger/            # 日志工具
│   ├── traceid/           # 链路追踪
│   └── errorlog/          # 错误日志
├── events/                  # 事件
├── listeners/               # 事件监听器
├── jobs/                    # 队列任务
├── providers/               # 服务提供者
├── rules/                   # 自定义验证规则
└── websocket/               # WebSocket
    └── notifications/      # 通知推送
```

### 分层架构

#### 1. 控制器层 (Controllers)

负责接收请求、验证参数、调用服务、返回响应。

```go
// app/http/controllers/admin/admin_controller.go
func (a *AdminController) Index(ctx http.Context) http.Response {
    // 1. 获取查询参数
    // 2. 调用 Service 获取数据
    // 3. 返回统一响应
}
```

#### 2. 服务层 (Services)

封装业务逻辑，可被多个控制器复用。

```go
// app/services/admin_service.go
type AdminService struct{}

func (s *AdminService) GetAdminWithRoles(id uint) (*models.Admin, error) {
    // 业务逻辑处理
}
```

#### 3. 模型层 (Models)

定义数据结构和数据库映射。

```go
// app/models/admin.go
type Admin struct {
    orm.Model
    Username     string       `gorm:"size:50;uniqueIndex" json:"username"`
    Password     string       `gorm:"size:255" json:"-"`
    Roles        []*Role      `gorm:"many2many:admin_role" json:"roles,omitempty"`
}
```

#### 4. 中间件层 (Middleware)

处理认证、权限、日志等横切关注点。

```go
// 中间件执行顺序
Request → TraceID → JWT Auth → Permission → Controller → OperationLog → Response
```

**核心中间件：**

| 中间件 | 功能 |
|--------|------|
| `jwt.go` | JWT 认证 |
| `permission.go` | 权限验证 |
| `operation_log.go` | 操作日志记录 |
| `blacklist.go` | IP 黑名单检查 |
| `rate_limiter.go` | 请求限流 |
| `trace_id.go` | 链路追踪 |

### 统一响应

```go
// 成功响应
response.Success(ctx, data)

// 错误响应
response.Error(ctx, http.StatusBadRequest, "错误信息")

// 泛型查找
admin, resp := response.FindByID[models.Admin](ctx, id, nil)
```

---

## 前端架构

### 目录结构

```
html/src/
├── api/                     # API 请求
├── components/              # 通用组件
│   ├── SearchForm.vue      # 搜索表单
│   ├── Pagination.vue      # 分页组件
│   ├── ErrorBoundary.vue   # 错误边界
│   └── ColumnSettingDialog.vue  # 列设置
├── composables/             # 可复用逻辑 (TypeScript)
│   ├── useCrud.ts          # CRUD 操作
│   ├── useDebounce.ts      # 防抖
│   ├── useTableSort.ts     # 表格排序
│   ├── usePermission.ts    # 权限检查
│   ├── useListPage.js      # 列表页面
│   └── useColumnSetting.js # 列设置
├── i18n/                    # 国际化
│   └── locales/
│       ├── zh-CN.json
│       └── en-US.json
├── layouts/                 # 布局组件
├── router/                  # 路由配置
├── store/                   # 状态管理 (Pinia)
│   ├── user.js             # 用户状态
│   ├── app.js              # 应用状态
│   └── tabs.js             # 标签页状态
├── types/                   # TypeScript 类型
│   ├── index.d.ts          # 实体类型
│   └── composables.d.ts    # Composable 类型
├── utils/                   # 工具函数
│   ├── request.js          # Axios 封装
│   ├── storage.js          # 存储工具
│   └── validation.js       # 验证器
└── views/                   # 页面组件
    ├── admin/
    ├── role/
    ├── menu/
    └── ...
```

### Composables 设计

核心 Composables 采用 TypeScript 编写，提供完整类型支持：

```typescript
// useCrud.ts - CRUD 操作封装
const { 
  dialogVisible,      // 对话框状态
  editId,             // 编辑ID
  handleAdd,          // 添加
  handleEdit,         // 编辑
  handleDelete,       // 删除
  handleBatchDelete   // 批量删除
} = useCrud({ deleteApi, batchDeleteApi })
```

```typescript
// usePermission.ts - 权限检查
const { 
  hasPermission,      // 检查权限
  shouldShowButton,   // 是否显示按钮
  getButtonState      // 获取按钮状态
} = usePermission()
```

### 状态管理

```
┌─────────────────────────────────────────┐
│              Pinia Store                 │
├─────────────┬─────────────┬─────────────┤
│  userStore  │  appStore   │  tabsStore  │
├─────────────┼─────────────┼─────────────┤
│ - userInfo  │ - sidebar   │ - tabs      │
│ - token     │ - fullscreen│ - activeTab │
│ - menus     │ - language  │             │
│ - permissions│            │             │
└─────────────┴─────────────┴─────────────┘
```

---

## 数据库设计

### ER 图

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   admins    │─────│ admin_role  │─────│    roles    │
├─────────────┤     ├─────────────┤     ├─────────────┤
│ id          │     │ admin_id    │     │ id          │
│ username    │     │ role_id     │     │ name        │
│ password    │     └─────────────┘     │ slug        │
│ department_id│                        │ status      │
└──────┬──────┘                        └──────┬──────┘
       │                                       │
       │    ┌─────────────┐                   │
       │    │ departments │                   │
       └────┤             │     ┌─────────────┴─────────────┐
            │ id          │     │       role_permission      │
            │ name        │     ├───────────────────────────┤
            │ parent_id   │     │ role_id                   │
            └─────────────┘     │ permission_id             │
                               └─────────────┬─────────────┘
                                             │
                               ┌─────────────▼─────────────┐
                               │       permissions         │
                               ├───────────────────────────┤
                               │ id                        │
                               │ name                      │
                               │ slug                      │
                               │ method                    │
                               │ path                      │
                               │ menu_id                   │
                               └───────────────────────────┘
```

### 核心表结构

| 表名 | 说明 |
|------|------|
| `admins` | 管理员 |
| `roles` | 角色 |
| `permissions` | 权限 |
| `menus` | 菜单 |
| `departments` | 部门 |
| `dictionaries` | 字典 |
| `blacklists` | 黑名单 |
| `operation_logs` | 操作日志 |
| `login_logs` | 登录日志 |
| `system_logs` | 系统日志 |
| `personal_access_tokens` | Token 管理 |
| `notifications` | 通知 |
| `attachments` | 附件 |
| `exports` | 导出任务 |

---

## 安全架构

### 认证流程

```
┌─────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Login  │───▶│ Validate    │───▶│ Generate    │───▶│   Return    │
│ Request │    │ Credentials │    │   JWT       │    │   Token     │
└─────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                     │
                     ▼
              ┌─────────────┐
              │ Rate Limit  │
              │   Check     │
              └─────────────┘
```

### 权限验证流程

```
Request → JWT Middleware → Permission Middleware → Controller
              │                    │
              ▼                    ▼
        验证 Token           检查权限
              │                    │
              ▼                    ▼
        解析用户信息         匹配路由权限
              │                    │
              ▼                    ▼
        写入 Context         允许/拒绝
```

### 安全特性

| 特性 | 实现方式 |
|------|----------|
| **认证** | JWT Token + 滑动过期 |
| **授权** | RBAC 权限模型 |
| **限流** | Token Bucket 算法 |
| **黑名单** | IP/Token 双重黑名单 |
| **日志** | 操作/登录日志全记录 |
| **追踪** | TraceID 链路追踪 |
| **敏感字段** | 密码等字段 `json:"-"` |

---

## 部署架构

### 单机部署

```
┌────────────────────────────────────────┐
│              Server                     │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐ │
│  │  Nginx  │──│ Goravel │──│  MySQL  │ │
│  └─────────┘  └─────────┘  └─────────┘ │
│                    │                    │
│               ┌────▼────┐              │
│               │  Redis  │              │
│               └─────────┘              │
└────────────────────────────────────────┘
```

### 高可用部署

```
                    ┌─────────────┐
                    │   Client    │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │     CDN     │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Load Balancer│
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           │               │               │
    ┌──────▼──────┐ ┌──────▼──────┐ ┌──────▼──────┐
    │   Goravel   │ │   Goravel   │ │   Goravel   │
    │   Server 1  │ │   Server 2  │ │   Server 3  │
    └──────┬──────┘ └──────┬──────┘ └──────┬──────┘
           │               │               │
           └───────────────┼───────────────┘
                           │
              ┌────────────┼────────────┐
              │                         │
       ┌──────▼──────┐          ┌──────▼──────┐
       │ MySQL Master│──────────│ Redis Cluster│
       └──────┬──────┘          └──────────────┘
              │
       ┌──────▼──────┐
       │ MySQL Slave │
       └─────────────┘
```

### Docker 部署

```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"
    depends_on:
      - mysql
      - redis

  mysql:
    image: mysql:8.0
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
```

---

## 性能优化

### 后端优化

| 优化项 | 实现 |
|--------|------|
| 数据库连接池 | 配置合理的连接数 |
| 查询优化 | 预加载关联、索引优化 |
| 缓存策略 | Redis 缓存热点数据 |
| 日志分级 | Debug/Info 分离 |

### 前端优化

| 优化项 | 实现 |
|--------|------|
| 虚拟滚动 | VXE-Table 大数据渲染 |
| 代码分割 | 路由懒加载 |
| 状态管理 | Pinia 响应式优化 |
| 防抖节流 | useDebounce composable |

---

## 扩展指南

### 添加新模块

1. **后端**
   - 创建 Model: `app/models/xxx.go`
   - 创建 Controller: `app/http/controllers/admin/xxx_controller.go`
   - 创建 Request: `app/http/requests/admin/xxx_request.go`
   - 注册路由: `routes/admin.go`

2. **前端**
   - 创建 API: `html/src/api/xxx.js`
   - 创建页面: `html/src/views/xxx/XxxList.vue`
   - 创建表单: `html/src/views/xxx/XxxForm.vue`
   - 注册路由: `html/src/router/index.js`

### 添加新权限

1. 数据库添加权限记录 (`permissions` 表)
2. 关联到菜单 (`menu_id`)
3. 分配给角色 (`role_permission` 表)

---

## 参考资料

- [Goravel 官方文档](https://www.goravel.dev)
- [Vue 3 文档](https://vuejs.org)
- [Element Plus 文档](https://element-plus.org)
- [VXE-Table 文档](https://vxetable.cn)

