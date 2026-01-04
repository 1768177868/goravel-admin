#!/bin/bash
# 从 Git 拉取代码并部署
# 在服务器上执行此脚本

set -e

# 配置（根据实际情况修改）
GIT_REPO_URL="${GIT_REPO_URL:-https://github.com/your-username/goravel-admin.git}"  # 修改为你的 Git 仓库
GIT_BRANCH="${GIT_BRANCH:-main}"  # 或 master
DEPLOY_DIR="${DEPLOY_DIR:-/www/goravel-admin}"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}=== 从 Git 拉取并部署 ===${NC}"
echo "仓库: $GIT_REPO_URL"
echo "分支: $GIT_BRANCH"
echo "目录: $DEPLOY_DIR"
echo ""

# 1. 检查部署目录
if [ ! -d "$DEPLOY_DIR" ]; then
    echo "创建部署目录: $DEPLOY_DIR"
    mkdir -p "$DEPLOY_DIR"
fi

cd "$DEPLOY_DIR"

# 2. 如果是第一次，克隆仓库
if [ ! -d ".git" ]; then
    echo -e "${GREEN}首次部署，克隆仓库...${NC}"
    if [ -z "$GIT_REPO_URL" ] || [ "$GIT_REPO_URL" = "https://github.com/your-username/goravel-admin.git" ]; then
        echo -e "${RED}错误: 请设置 GIT_REPO_URL 环境变量或修改脚本中的仓库地址${NC}"
        exit 1
    fi
    git clone -b "$GIT_BRANCH" "$GIT_REPO_URL" .
else
    echo -e "${GREEN}拉取最新代码...${NC}"
    git fetch origin
    git reset --hard origin/$GIT_BRANCH
    git pull origin $GIT_BRANCH || {
        echo -e "${YELLOW}警告: git pull 失败，尝试强制重置${NC}"
        git fetch origin
        git reset --hard origin/$GIT_BRANCH
    }
fi

# 3. 检查必要文件
if [ ! -f "Dockerfile" ]; then
    echo -e "${RED}错误: 未找到 Dockerfile${NC}"
    exit 1
fi

if [ ! -f "docker-compose.blue.yml" ] || [ ! -f "docker-compose.green.yml" ]; then
    echo -e "${RED}错误: 未找到 Docker Compose 配置文件${NC}"
    exit 1
fi

# 4. 确保 .env 文件存在
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}警告: .env 文件不存在${NC}"
    if [ -f ".env.example" ]; then
        echo "从 .env.example 复制..."
        cp .env.example .env
        echo -e "${YELLOW}请编辑 .env 文件并设置正确的配置${NC}"
    else
        echo -e "${RED}错误: 未找到 .env 或 .env.example 文件${NC}"
        exit 1
    fi
fi

# 5. 确保部署脚本可执行
if [ -f "scripts/deploy/docker-blue-green.sh" ]; then
    chmod +x scripts/deploy/docker-blue-green.sh
    echo -e "${GREEN}执行部署脚本...${NC}"
    ./scripts/deploy/docker-blue-green.sh
else
    echo -e "${RED}错误: 未找到 deploy.sh${NC}"
    exit 1
fi

echo -e "${GREEN}=== 部署完成 ===${NC}"

