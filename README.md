<p align="center"><img src="https://www.goravel.dev/logo.png?v=1.14.x" width="300"></p>

English | [中文](./README_zh.md)

## About Goravel

Goravel is a web application framework with complete functions and good scalability. As a starting scaffolding to help Gopher quickly build their own applications.

The framework style is consistent with [Laravel](https://github.com/laravel/laravel), let Phper don't need to learn a new framework, but also happy to play around Golang! Tribute Laravel!

Welcome to star, PR and issues！

## Admin System

This project includes a complete admin management system built with Goravel framework.

<p align="center"><img src="./admin.png" alt="Admin System Screenshot" width="800"></p>

### Features

#### Core Modules
- **Authentication & Authorization**
  - JWT-based authentication
  - Role-based access control (RBAC)
  - Permission management
  - Multi-token management
  - Online user monitoring and kick-out

- **User Management**
  - Admin user management
  - Department management
  - Role management
  - Permission assignment
  - Password reset

- **System Configuration**
  - Menu management (dynamic menu)
  - Dictionary management
  - System configuration
  - Blacklist management

- **Logging & Monitoring**
  - Operation logs (with automatic recording)
  - Login logs
  - System logs (with trace ID)
  - Service monitoring

- **Additional Features**
  - Dashboard with statistics
  - Notification center (WebSocket real-time notifications)
  - Data export management
  - Multi-language support (Chinese/English)
  - Responsive UI design

### Tech Stack

**Backend:**
- Goravel Framework (Go)
- JWT Authentication
- RBAC Permission System
- WebSocket Support
- Database Migrations & Seeders

**Frontend:**
- Vue 3
- Element Plus
- vxe-table (Advanced table component)
- Vue Router
- Pinia (State management)
- Axios
- ECharts (Data visualization)
- vue-i18n (Internationalization)

### Quick Start

1. **Backend Setup:**
   ```bash
   # Install dependencies
   go mod download
   
   # Configure database in .env
   # Run migrations and seeders
   go run . artisan migrate
   go run . artisan db:seed
   
   # Start server
   go run . --no-ansi
   # or use air for live reload
   air
   ```

2. **Frontend Setup:**
   ```bash
   cd html
   
   # Install dependencies
   npm install
   
   # Configure API address in .env
   # VITE_API_BASE_URL=http://127.0.0.1:3008
   # VITE_API_PREFIX=/api/admin
   
   # Start development server
   npm run dev
   ```

3. **Default Login:**
   - Username: `admin`
   - Password: `admin123`
   - (Please change the default password after first login)

### API Documentation

The admin API endpoints are prefixed with `/api/admin`. All endpoints require JWT authentication except login and captcha.

For detailed API documentation, see [routes/admin.go](./routes/admin.go)

### Project Structure

```
.
├── app/
│   ├── http/
│   │   ├── controllers/admin/    # Admin controllers
│   │   ├── middleware/           # Custom middleware (JWT, Permission, OperationLog)
│   │   └── helpers/              # Helper functions
│   ├── models/                   # Database models
│   └── services/                 # Business logic services
├── routes/
│   └── admin.go                  # Admin routes
├── database/
│   ├── migrations/               # Database migrations
│   └── seeders/                  # Database seeders
├── html/                         # Frontend Vue application
│   └── src/
│       ├── views/                # Page components
│       ├── components/           # Reusable components
│       ├── api/                 # API client
│       └── store/               # Pinia stores
└── config/                       # Configuration files
```

### Security Features

- JWT token-based authentication
- Permission middleware for route protection
- Automatic operation logging
- Sensitive data filtering in logs
- Rate limiting on login endpoints
- Blacklist management for IP/User blocking
- Token revocation support

## Getting Started

### Start Service

`go run . --no-ansi` or `air`

[About air]: https://www.goravel.dev/getting-started/installation.html#live-reload

### DB

[app/http/controllers/db_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/db_controller.go)

### Websocket

[app/http/controllers/websocket_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/websocket_controller.go)

### Validation

[app/http/controllers/validation_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/validation_controller.go)

### Auth

[app/http/controllers/auth_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/auth_controller.go)

### Unit Test (Testing With Mock)

[app/http/controllers/validation_controller_test.go](https://github.com/goravel/example/blob/master/app/http/controllers/validation_controller_test.go)

### Integration Test (Testing With Configuration)

[tests/controllers/validation_controller_test.go](https://github.com/goravel/example/blob/master/tests/controllers/validation_controller_test.go)

### GRPC

[app/grpc/controllers/user_controller.go](https://github.com/goravel/example/blob/master/app/grpc/controllers/user_controller.go)

### Swagger(For gin HTTP driver)

[app/http/controllers/swagger_controller.go](https://github.com/goravel/example/blob/master/app/http/controllers/swagger_controller.go)

### Integration of single page application into the framework

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L26)

### View nesting

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L33)

### Session

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L42)

### Cookie

[routes/web.go](https://github.com/goravel/example/blob/master/routes/web.go#L58)

### Localization

[routes/api.go](https://github.com/goravel/example/blob/master/routes/api.go#L37)

### GraphQL

```bash
# download and install gqlgen locally, only need to run it once
go get -d github.com/99designs/gqlgen
# regenerate code
go run github.com/99designs/gqlgen generate
```

```
swag init
go run .
http://localhost:3008/swagger/
```

static build
```
go build --ldflags "-extldflags -static" -o main .
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

**重要配置：**

1. **环境变量设置**（在 Cloudflare Pages 项目设置中）：
   - `VITE_API_BASE_URL`: `https://api.xuancheng888.top`
   - `VITE_API_PREFIX`: `/api/admin`

2. **解决刷新 404 问题**：
   
   在 Cloudflare Pages 项目设置中，添加 **重写规则**：
   
   - 进入项目设置 → **Functions** → **Routes**
   - 添加规则：`/*` → `/index.html` (Status: 200)
   
   或者使用 **Functions**（推荐）：
   
   在 `html/functions/_middleware.js` 创建文件：
   ```javascript
   export function onRequest(context) {
     const url = new URL(context.request.url)
     // 如果是静态资源，直接返回
     if (url.pathname.startsWith('/assets/') || 
         url.pathname.startsWith('/favicon.ico') ||
         url.pathname.match(/\.(js|css|png|jpg|jpeg|gif|svg|ico|woff|woff2|ttf|eot)$/)) {
       return context.next()
     }
     // 其他所有请求都返回 index.html
     return context.next({
       request: new Request(new URL('/index.html', context.request.url), context.request)
     })
   }
   ```

3. **验证环境变量**：
   
   构建后，环境变量会被注入到代码中。可以通过浏览器控制台检查：
   ```javascript
   console.log(import.meta.env.VITE_API_BASE_URL)
   console.log(import.meta.env.VITE_API_PREFIX)
   ```

## Documentation

Online documentation [https://www.goravel.dev](https://www.goravel.dev)

> To optimize the documentation, please submit a PR to the documentation
> repository [https://github.com/goravel/docs](https://github.com/goravel/docs)

## Group

Welcome more discussion in Discord.

[https://discord.gg/cFc5csczzS](https://discord.gg/cFc5csczzS)

## License

The Goravel framework is open-sourced software licensed under the [MIT license](https://opensource.org/licenses/MIT).
