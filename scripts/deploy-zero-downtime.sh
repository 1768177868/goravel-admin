#!/bin/bash
# 零停机部署脚本（Nginx + 双实例）
# 使用方法: ./scripts/deploy-zero-downtime.sh [服务器用户] [服务器地址] [应用路径]

set -e

# 配置
SERVER_USER=${1:-"www-data"}
SERVER_HOST=${2:-"your-server.com"}
APP_PATH=${3:-"/www/goravel-admin"}
SERVICE_PREFIX="goravel-admin"
PORT1=3000
PORT2=3001

echo "=========================================="
echo "开始零停机部署"
echo "=========================================="

# 1. 编译应用（在本地）
echo ""
echo "1. 正在编译应用..."
go build --ldflags "-extldflags -static -s -w" -o main .
if [ $? -ne 0 ]; then
    echo "❌ 编译失败！"
    exit 1
fi
echo "✅ 编译成功"

# 2. 上传到服务器临时目录
echo ""
echo "2. 正在上传到服务器..."
scp main ${SERVER_USER}@${SERVER_HOST}:/tmp/goravel-admin-main.new
if [ $? -ne 0 ]; then
    echo "❌ 上传失败！"
    exit 1
fi
echo "✅ 上传成功"

# 3. 在服务器上执行零停机部署
echo ""
echo "3. 正在执行零停机部署..."
ssh ${SERVER_USER}@${SERVER_HOST} << EOF
set -e

cd ${APP_PATH}

# 确定当前运行的端口
if systemctl is-active --quiet ${SERVICE_PREFIX}-${PORT1} 2>/dev/null; then
    CURRENT_PORT=${PORT1}
    NEW_PORT=${PORT2}
elif systemctl is-active --quiet ${SERVICE_PREFIX}-${PORT2} 2>/dev/null; then
    CURRENT_PORT=${PORT2}
    NEW_PORT=${PORT1}
else
    echo "⚠️  没有检测到运行中的实例，将使用端口 ${PORT1}"
    CURRENT_PORT=""
    NEW_PORT=${PORT1}
fi

echo "当前运行端口: \${CURRENT_PORT:-无}"
echo "新实例端口: \${NEW_PORT}"

# 备份并替换文件
echo ""
echo "4. 备份并替换文件..."
if [ -f main ]; then
    BACKUP_NAME="main.old.\$(date +%Y%m%d_%H%M%S)"
    mv main "\${BACKUP_NAME}"
    echo "✅ 已备份到: \${BACKUP_NAME}"
fi

mv /tmp/goravel-admin-main.new main
chmod +x main
chown www-data:www-data main
echo "✅ 文件替换完成"

# 启动新实例
echo ""
echo "5. 启动新实例（端口 \${NEW_PORT}）..."
sudo systemctl start ${SERVICE_PREFIX}-\${NEW_PORT} || {
    echo "❌ 启动新实例失败"
    exit 1
}

# 健康检查
echo ""
echo "6. 等待新实例启动并健康检查..."
HEALTH_CHECK_PASSED=false
for i in {1..30}; do
    if curl -f -s http://localhost:\${NEW_PORT}/api/admin/health > /dev/null 2>&1; then
        echo "✅ 新实例健康检查通过（尝试 \${i}/30）"
        HEALTH_CHECK_PASSED=true
        break
    fi
    echo "等待新实例启动... (\${i}/30)"
    sleep 1
done

if [ "\${HEALTH_CHECK_PASSED}" != "true" ]; then
    echo "❌ 新实例健康检查失败，部署中止"
    sudo systemctl stop ${SERVICE_PREFIX}-\${NEW_PORT} || true
    exit 1
fi

# 更新 Nginx 配置
if [ -n "\${CURRENT_PORT}" ]; then
    echo ""
    echo "7. 更新 Nginx 配置，添加新实例..."
    NGINX_CONF="/etc/nginx/sites-available/goravel-admin"
    
    # 检查新端口是否已经在配置中
    if grep -q "server 127.0.0.1:\${NEW_PORT}" "\${NGINX_CONF}"; then
        echo "⚠️  新端口已在 Nginx 配置中"
    else
        # 添加新端口到 upstream（取消注释或添加）
        sudo sed -i "s/# server 127.0.0.1:\${NEW_PORT}/server 127.0.0.1:\${NEW_PORT}/" "\${NGINX_CONF}" || \
        sudo sed -i "/upstream goravel_backend {/a\\    server 127.0.0.1:\${NEW_PORT} max_fails=3 fail_timeout=30s;" "\${NGINX_CONF}"
        
        # 测试 Nginx 配置
        if sudo nginx -t; then
            sudo nginx -s reload
            echo "✅ Nginx 配置已更新并重载"
        else
            echo "❌ Nginx 配置测试失败"
            exit 1
        fi
    fi
    
    # 等待旧实例请求完成
    echo ""
    echo "8. 等待旧实例请求完成..."
    sleep 5
    
    # 停止旧实例
    echo ""
    echo "9. 停止旧实例（端口 \${CURRENT_PORT}）..."
    sudo systemctl stop ${SERVICE_PREFIX}-\${CURRENT_PORT} || true
    
    # 从 Nginx 配置中移除旧端口
    echo ""
    echo "10. 从 Nginx 配置中移除旧端口..."
    sudo sed -i "s/server 127.0.0.1:\${CURRENT_PORT}/# server 127.0.0.1:\${CURRENT_PORT}/" "\${NGINX_CONF}" || true
    sudo nginx -t && sudo nginx -s reload
    echo "✅ 旧实例已停止，Nginx 配置已更新"
else
    echo ""
    echo "7. 首次部署，配置 Nginx..."
    # 首次部署时，确保 Nginx 配置正确
    sudo nginx -t && sudo nginx -s reload || echo "⚠️  Nginx 配置需要手动检查"
fi

echo ""
echo "=========================================="
echo "✅ 零停机部署完成！"
echo "=========================================="
echo "新实例运行在端口: \${NEW_PORT}"
if [ -n "\${CURRENT_PORT}" ]; then
    echo "旧实例已停止（端口: \${CURRENT_PORT}）"
fi
echo ""
echo "查看服务状态:"
echo "  sudo systemctl status ${SERVICE_PREFIX}-\${NEW_PORT}"
echo ""
echo "查看日志:"
echo "  sudo journalctl -u ${SERVICE_PREFIX}-\${NEW_PORT} -f"
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "=========================================="
    echo "✅ 部署成功完成！"
    echo "=========================================="
else
    echo ""
    echo "=========================================="
    echo "❌ 部署失败！"
    echo "=========================================="
    exit 1
fi

