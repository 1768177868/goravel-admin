#!/bin/bash
# 从 Git 回滚到指定版本并部署
# 用于回滚到特定的 Git 提交

set -e

# 配置
GIT_BRANCH="${GIT_BRANCH:-main}"
DEPLOY_DIR="${DEPLOY_DIR:-/www/goravel-admin}"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${YELLOW}=== Git 版本回滚 ===${NC}"

# 检查参数
if [ -z "$1" ]; then
    echo -e "${BLUE}用法: $0 <commit-hash|tag|branch>${NC}"
    echo "示例:"
    echo "  $0 abc1234          # 回滚到指定提交"
    echo "  $0 v1.0.0           # 回滚到指定标签"
    echo "  $0 HEAD~1           # 回滚到上一个提交"
    echo "  $0 origin/main      # 回滚到远程分支"
    echo ""
    echo -e "${YELLOW}最近的提交历史:${NC}"
    cd "$DEPLOY_DIR" 2>/dev/null || cd "$(dirname "$0")/../../"
    git log --oneline -10 2>/dev/null || echo "无法获取 Git 历史"
    exit 1
fi

TARGET_VERSION="$1"

# 检查部署目录
if [ ! -d "$DEPLOY_DIR" ]; then
    # 尝试使用当前目录
    if [ -d ".git" ]; then
        DEPLOY_DIR="$(pwd)"
    else
        echo -e "${RED}错误: 部署目录不存在: $DEPLOY_DIR${NC}"
        exit 1
    fi
fi

cd "$DEPLOY_DIR"

# 检查是否是 Git 仓库
if [ ! -d ".git" ]; then
    echo -e "${RED}错误: 当前目录不是 Git 仓库${NC}"
    exit 1
fi

# 显示当前状态
CURRENT_COMMIT=$(git rev-parse --short HEAD)
echo -e "当前提交: ${YELLOW}$CURRENT_COMMIT${NC}"
echo -e "目标版本: ${YELLOW}$TARGET_VERSION${NC}"
echo ""

# 确认回滚
read -p "确认要回滚到 $TARGET_VERSION 吗? (y/N): " confirm
if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo -e "${YELLOW}取消回滚${NC}"
    exit 0
fi

# 检查目标版本是否存在
if ! git rev-parse --verify "$TARGET_VERSION" > /dev/null 2>&1; then
    echo -e "${RED}错误: 未找到版本 $TARGET_VERSION${NC}"
    echo -e "${YELLOW}提示: 尝试先执行 git fetch origin${NC}"
    exit 1
fi

# 回滚到目标版本
echo -e "${GREEN}回滚到 $TARGET_VERSION...${NC}"
git fetch origin 2>/dev/null || true
git checkout "$TARGET_VERSION"

if [ $? -ne 0 ]; then
    echo -e "${RED}回滚失败！${NC}"
    exit 1
fi

NEW_COMMIT=$(git rev-parse --short HEAD)
echo -e "${GREEN}✓ 已回滚到 $NEW_COMMIT${NC}"
echo ""

# 执行部署
if [ -f "scripts/deploy/docker-blue-green.sh" ]; then
    echo -e "${GREEN}执行部署...${NC}"
    chmod +x scripts/deploy/docker-blue-green.sh
    ./scripts/deploy/docker-blue-green.sh
else
    echo -e "${RED}错误: 未找到部署脚本${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== Git 回滚完成 ===${NC}"
echo -e "当前版本: ${YELLOW}$NEW_COMMIT${NC}"

