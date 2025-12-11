<p align="center"><img src="https://www.goravel.dev/logo.png?v=1.14.x" width="300"></p>

English | [中文](./README_zh.md)

## About Goravel

Goravel is a web application framework with complete functions and good scalability. As a starting scaffolding to help Gopher quickly build their own applications.

The framework style is consistent with [Laravel](https://github.com/laravel/laravel), let Phper don't need to learn a new framework, but also happy to play around Golang! Tribute Laravel!

Welcome to star, PR and issues！

## Admin System

This project includes a complete admin management system built with Goravel framework.

> Demo https://admin.xuancheng888.top 

username: demo  
password: demo123

### Screenshots

<p align="center">
  <img src="./images/login.png" alt="Login Page" width="800">
  <p align="center">Login Page</p>
</p>

<p align="center">
  <img src="./images/admin.png" alt="Admin Dashboard" width="800">
  <p align="center">Admin Dashboard</p>
</p>

<p align="center">
  <img src="./images/monitor.png" alt="System Monitoring" width="800">
  <p align="center">System Monitoring</p>
</p>

### Features

#### Core Modules
- **Authentication & Authorization**
  - JWT-based authentication
  - Role-based access control (RBAC)
  - Permission management
  - Multi-token management
  - Online user monitoring and kick-out

- **Admin Management**
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
   go mod tidy
   
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
   # VITE_API_BASE_URL=http://127.0.0.1:3000
   # VITE_API_PREFIX=/api/admin
   
   # Start development server
   npm run dev
   ```

3. **Default Login:**
   - Username: `admin`
   - Password: `admin123`
   - (Please change the default password after first login)

### Build & Deployment

For detailed build and deployment instructions, including cross-platform compilation, Docker deployment, and systemd service setup, please refer to [BUILD.md](./BUILD.md).

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
├── config/                       # Configuration files
├── BUILD.md                      # Build and deployment documentation
└── images/                       # Screenshots
```

### Security Features

- JWT token-based authentication
- Permission middleware for route protection
- Automatic operation logging
- Sensitive data filtering in logs
- Rate limiting on login endpoints
- Blacklist management for IP/Admin blocking
- Token revocation support

## Getting Started

### Start Service

`go run . --no-ansi` or `air`

[About air]: https://www.goravel.dev/getting-started/installation.html#live-reload


### Cloudflare Workers Deployment

Deploy the frontend application to Cloudflare Workers:

```bash
# Build the frontend application
cd html
npm install && npm run build

# Deploy to Cloudflare Workers
npx wrangler deploy --assets ./dist --compatibility-date 2025-11-29 --name admin
```

**Configuration:**

- **Root directory:** `html`
- **Environment variables (Variables & Secrets):**
  - `VITE_API_BASE_URL`: `https://api.xuancheng888.top`
  - `VITE_API_PREFIX`: `/api/admin`
- **Custom domain:** `admin.xuancheng888.top`

**Note:** The `worker.js` file automatically handles SPA routing by returning `index.html` when a file doesn't exist.

### Performance Profiling

pprof is available at: http://localhost:3000/debug/pprof/

## Documentation

Online documentation [https://www.goravel.dev](https://www.goravel.dev)

> To optimize the documentation, please submit a PR to the documentation
> repository [https://github.com/goravel/docs](https://github.com/goravel/docs)

## Group

Welcome more discussion in Discord.

[https://discord.gg/cFc5csczzS](https://discord.gg/cFc5csczzS)

## License

The Goravel framework is open-sourced software licensed under the [MIT license](https://opensource.org/licenses/MIT).
