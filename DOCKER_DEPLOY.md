# Docker 蓝绿部署快速指南

## 📋 概述

本项目支持 Docker Compose 蓝绿部署，实现零停机更新。适合本地没有 Docker 环境，但服务器有 Docker 的场景。

## 🚀 快速开始

### 1. 本地开发（无需 Docker）

```bash
# 正常开发，提交代码
git add .
git commit -m "更新功能"
git push origin main
```

### 2. 服务器部署

#### 首次部署

```bash
# SSH 登录服务器
ssh user@your-server.com

# 创建部署目录
mkdir -p /www/goravel-admin
cd /www/goravel-admin

# 克隆仓库
git clone https://github.com/your-username/goravel-admin.git .

# 配置环境变量
cp .env.example .env  # 如果存在
vim .env  # 编辑配置

# 执行部署
chmod +x scripts/deploy/git-deploy.sh
./scripts/deploy/git-deploy.sh
```

#### 后续部署

```bash
# SSH 登录服务器
ssh user@your-server.com

# 进入部署目录
cd /www/goravel-admin

# 方式一：使用 Git 部署脚本（推荐）
./scripts/deploy/git-deploy.sh

# 方式二：手动拉取代码后部署
git pull origin main
./scripts/deploy/docker-blue-green.sh
```

## 📁 文件结构

```
项目根目录/
├── docker-compose.blue.yml      # 蓝环境配置（端口 3000）
├── docker-compose.green.yml     # 绿环境配置（端口 3001）
├── Dockerfile                    # Docker 镜像构建文件
└── scripts/
    └── deploy/
        ├── docker-blue-green.sh # 蓝绿部署主脚本
        ├── git-deploy.sh        # Git 拉取并部署脚本
        └── README.md            # 详细说明文档
```

## 🔧 配置说明

### 环境变量

在服务器上创建 `.env` 文件，配置数据库、Redis 等：

```env
APP_ENV=production
APP_HOST=0.0.0.0
APP_PORT=3000
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=your_database
DB_USERNAME=your_username
DB_PASSWORD=your_password
# ... 其他配置
```

### Git 仓库配置

如果使用 `git-deploy.sh`，可以设置环境变量：

```bash
export GIT_REPO_URL="https://github.com/your-username/goravel-admin.git"
export GIT_BRANCH="main"
export DEPLOY_DIR="/www/goravel-admin"
```

或直接编辑 `scripts/deploy/git-deploy.sh` 文件中的配置。

## 📊 部署流程

1. **检测当前版本** - 自动检测运行的是 `blue` 还是 `green`
2. **构建新版本** - 在备用环境构建新 Docker 镜像
3. **启动新版本** - 启动新版本容器（使用不同端口）
4. **健康检查** - 等待新版本通过健康检查
5. **切换流量** - 更新 Nginx 配置（如果存在）
6. **停止旧版本** - 停止旧版本容器

## 🔍 查看状态

```bash
# 查看运行中的容器
docker ps | grep goravel-admin

# 查看容器日志
docker logs -f goravel-admin-blue
docker logs -f goravel-admin-green

# 查看健康状态
docker inspect --format='{{.State.Health.Status}}' goravel-admin-blue
```

## 🔄 回滚

如果需要回滚到上一个版本：

```bash
cd /www/goravel-admin

# 方式一：手动切换容器
docker-compose -f docker-compose.blue.yml up -d  # 启动 blue
docker-compose -f docker-compose.green.yml down  # 停止 green

# 方式二：回滚代码后重新部署
git checkout <previous-commit>
./scripts/deploy/docker-blue-green.sh
```

## ⚠️ 故障处理

如果部署失败：
- 脚本会自动停止新版本容器
- 旧版本继续运行
- 检查日志：`docker logs goravel-admin-<color>`

## 📝 注意事项

1. **首次部署**：确保服务器已安装 Docker 和 Docker Compose
2. **环境变量**：确保 `.env` 文件配置正确
3. **健康检查**：应用需要提供 `/health` 端点（已配置）
4. **端口冲突**：确保 3000 和 3001 端口未被占用
5. **存储持久化**：`storage` 目录会持久化到宿主机

## 🔗 相关文档

- 详细部署文档：`docs/BUILD.md`
- 脚本说明：`scripts/deploy/README.md`

## 💡 提示

- 本地开发无需 Docker，只需 Go 环境
- 部署在服务器上自动完成，无需手动操作
- 支持零停机部署，用户体验无影响
- 支持快速回滚，降低部署风险

