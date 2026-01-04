#!/bin/bash
# Docker Compose 蓝绿部署回滚脚本
# 快速回滚到上一个版本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../../"  # 回到项目根目录

# 加载部署配置
source "$SCRIPT_DIR/deploy-config.sh"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检测当前运行的版本
detect_current_version() {
    if docker ps --format "{{.Names}}" | grep -q "goravel-admin-blue"; then
        if [ "$(docker inspect -f '{{.State.Running}}' goravel-admin-blue 2>/dev/null)" = "true" ]; then
            echo "blue"
            return
        fi
    fi
    if docker ps --format "{{.Names}}" | grep -q "goravel-admin-green"; then
        if [ "$(docker inspect -f '{{.State.Running}}' goravel-admin-green 2>/dev/null)" = "true" ]; then
            echo "green"
            return
        fi
    fi
    echo "none"
}

CURRENT_COLOR=$(detect_current_version)

if [ "$CURRENT_COLOR" = "none" ]; then
    echo -e "${RED}错误: 未检测到运行中的容器${NC}"
    exit 1
fi

PREVIOUS_COLOR=$([ "$CURRENT_COLOR" = "blue" ] && echo "green" || echo "blue")
CURRENT_PORT=$([ "$CURRENT_COLOR" = "blue" ] && echo "$BLUE_PORT" || echo "$GREEN_PORT")
PREVIOUS_PORT=$([ "$PREVIOUS_COLOR" = "blue" ] && echo "$BLUE_PORT" || echo "$GREEN_PORT")

echo -e "${YELLOW}=== Docker Compose 蓝绿部署回滚 ===${NC}"
echo -e "当前运行: ${YELLOW}$CURRENT_COLOR${NC} (端口: $CURRENT_PORT)"
echo -e "回滚到: ${YELLOW}$PREVIOUS_COLOR${NC} (端口: $PREVIOUS_PORT)"
echo ""

# 检查 Docker 和 Docker Compose
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker 未安装${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}错误: Docker Compose 未安装${NC}"
    exit 1
fi

# 使用 docker compose 或 docker-compose
COMPOSE_CMD="docker-compose"
if ! command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker compose"
fi

# 导出端口配置到环境变量（供 docker-compose 使用）
export BLUE_PORT GREEN_PORT CONTAINER_PORT

# 检查上一个版本的容器是否存在（可能已停止但未删除）
PREVIOUS_CONTAINER_EXISTS=false
if docker ps -a --format "{{.Names}}" | grep -q "goravel-admin-${PREVIOUS_COLOR}"; then
    PREVIOUS_CONTAINER_EXISTS=true
fi

# 如果上一个版本的容器不存在，需要重新构建
if [ "$PREVIOUS_CONTAINER_EXISTS" = "false" ]; then
    echo -e "${YELLOW}警告: 上一个版本的容器不存在，需要重新构建${NC}"
    echo -e "${BLUE}选项：${NC}"
    echo "1. 从 Git 回滚到上一个提交并重新构建"
    echo "2. 使用当前代码重新构建上一个版本"
    echo "3. 取消回滚"
    read -p "请选择 (1/2/3): " choice
    
    case $choice in
        1)
            echo -e "${GREEN}从 Git 回滚...${NC}"
            read -p "请输入要回滚的 Git 提交哈希 (或按 Enter 使用 HEAD~1): " commit_hash
            commit_hash=${commit_hash:-HEAD~1}
            git checkout $commit_hash
            echo -e "${GREEN}构建回滚版本...${NC}"
            $COMPOSE_CMD -f docker-compose.${PREVIOUS_COLOR}.yml build --no-cache
            ;;
        2)
            echo -e "${GREEN}使用当前代码重新构建...${NC}"
            $COMPOSE_CMD -f docker-compose.${PREVIOUS_COLOR}.yml build --no-cache
            ;;
        3)
            echo -e "${YELLOW}取消回滚${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}无效选择${NC}"
            exit 1
            ;;
    esac
fi

# 1. 启动上一个版本
echo -e "${GREEN}[1/4] 启动上一个版本容器 (端口 $PREVIOUS_PORT)...${NC}"
$COMPOSE_CMD -f docker-compose.${PREVIOUS_COLOR}.yml up -d

if [ $? -ne 0 ]; then
    echo -e "${RED}启动失败！${NC}"
    exit 1
fi

# 2. 等待容器启动
echo -e "${GREEN}[2/4] 等待容器启动...${NC}"
sleep 5

# 3. 健康检查
echo -e "${GREEN}[3/4] 执行健康检查...${NC}"
HEALTH_CHECK_FAILED=true
for i in {1..30}; do
    # 检查容器是否运行
    if ! docker ps --format "{{.Names}}" | grep -q "goravel-admin-${PREVIOUS_COLOR}"; then
        echo "容器未运行，等待中... ($i/30)"
        sleep 2
        continue
    fi
    
    # 检查健康状态
    HEALTH=$(docker inspect --format='{{.State.Health.Status}}' goravel-admin-${PREVIOUS_COLOR} 2>/dev/null || echo "none")
    
    # 尝试 HTTP 健康检查
    if curl -f -s http://localhost:$PREVIOUS_PORT/health > /dev/null 2>&1 || \
       curl -f -s http://localhost:$PREVIOUS_PORT/ > /dev/null 2>&1; then
        echo -e "${GREEN}✓ 回滚版本健康检查通过 (尝试 $i/30)${NC}"
        HEALTH_CHECK_FAILED=false
        break
    fi
    
    if [ "$HEALTH" = "healthy" ]; then
        echo -e "${GREEN}✓ 回滚版本健康检查通过 (Docker 健康检查)${NC}"
        HEALTH_CHECK_FAILED=false
        break
    fi
    
    echo "等待健康检查... ($i/30)"
    sleep 2
done

if [ "$HEALTH_CHECK_FAILED" = "true" ]; then
    echo -e "${RED}✗ 健康检查失败，回滚中止${NC}"
    echo -e "${YELLOW}保持当前版本运行${NC}"
    $COMPOSE_CMD -f docker-compose.${PREVIOUS_COLOR}.yml down
    exit 1
fi

# 4. 切换 Nginx 流量
echo -e "${GREEN}[4/4] 切换 Nginx 流量...${NC}"
NGINX_CONF="/etc/nginx/sites-available/goravel-admin"
if [ -f "$NGINX_CONF" ]; then
    # 备份配置
    BACKUP_FILE="${NGINX_CONF}.backup.rollback.$(date +%Y%m%d_%H%M%S)"
    cp "$NGINX_CONF" "$BACKUP_FILE"
    echo "已备份 Nginx 配置到: $BACKUP_FILE"
    
    # 更新 upstream 配置
    sed -i "s/server 127.0.0.1:$BLUE_PORT/server 127.0.0.1:$PREVIOUS_PORT/" "$NGINX_CONF"
    sed -i "s/server 127.0.0.1:$GREEN_PORT/server 127.0.0.1:$PREVIOUS_PORT/" "$NGINX_CONF"
    
    # 测试并重载 Nginx
    if nginx -t 2>/dev/null; then
        nginx -s reload
        echo -e "${GREEN}✓ Nginx 流量已切换到端口 $PREVIOUS_PORT${NC}"
    else
        echo -e "${YELLOW}警告: Nginx 配置测试失败，恢复备份${NC}"
        mv "$BACKUP_FILE" "$NGINX_CONF"
    fi
else
    echo -e "${YELLOW}提示: Nginx 配置文件不存在，跳过流量切换${NC}"
    echo -e "${YELLOW}请手动配置负载均衡指向端口 $PREVIOUS_PORT${NC}"
fi

# 5. 等待流量切换完成
echo -e "${GREEN}等待流量切换完成...${NC}"
sleep 5

# 6. 停止当前版本
echo -e "${GREEN}停止当前版本 ($CURRENT_COLOR)...${NC}"
$COMPOSE_CMD -f docker-compose.${CURRENT_COLOR}.yml down

echo ""
echo -e "${GREEN}=== 回滚完成 ===${NC}"
echo -e "当前运行版本: ${YELLOW}$PREVIOUS_COLOR${NC} (端口: $PREVIOUS_PORT)"
echo -e "已停止版本: ${YELLOW}$CURRENT_COLOR${NC}"
echo ""
echo "查看日志: docker logs -f goravel-admin-${PREVIOUS_COLOR}"
echo "查看状态: docker ps | grep goravel-admin"

