# Docker 蓝绿部署脚本

## 文件说明

- `docker-blue-green.sh` - 蓝绿部署主脚本，执行零停机部署
- `git-deploy.sh` - 从 Git 拉取代码并自动部署的脚本

## 使用方法

### 方式一：手动部署（已拉取代码）

```bash
cd /www/goravel-admin
chmod +x scripts/deploy/docker-blue-green.sh
./scripts/deploy/docker-blue-green.sh
```

### 方式二：从 Git 拉取并部署

```bash
# 设置环境变量（可选）
export GIT_REPO_URL="https://github.com/your-username/goravel-admin.git"
export GIT_BRANCH="main"
export DEPLOY_DIR="/www/goravel-admin"

# 执行部署
chmod +x scripts/deploy/git-deploy.sh
./scripts/deploy/git-deploy.sh
```

或者直接修改脚本中的配置：

```bash
# 编辑脚本
vim scripts/deploy/git-deploy.sh

# 修改这些变量：
# GIT_REPO_URL="你的仓库地址"
# GIT_BRANCH="main"
# DEPLOY_DIR="/www/goravel-admin"
```

## 部署流程

1. **检测当前版本** - 自动检测运行的是 blue 还是 green
2. **构建新版本** - 在备用环境构建新镜像
3. **启动新版本** - 启动新版本容器（不同端口）
4. **健康检查** - 等待新版本通过健康检查
5. **切换流量** - 更新 Nginx 配置（如果存在）
6. **停止旧版本** - 停止旧版本容器

## 前置要求

- Docker 和 Docker Compose 已安装
- 服务器上已配置 `.env` 文件
- 如果需要 Nginx 切换，需要配置 Nginx

## 健康检查

应用需要提供 `/health` 端点，已在 `routes/web.go` 中配置。

## 故障处理

如果部署失败，脚本会自动回滚：
- 停止新版本容器
- 保持旧版本运行

## 查看日志

```bash
# 查看运行中的容器日志
docker logs -f goravel-admin-blue
docker logs -f goravel-admin-green

# 查看容器状态
docker ps | grep goravel-admin
```

