#!/bin/bash
# overseer 热更新部署脚本
# 使用方法: ./scripts/deploy_overseer.sh [服务器用户] [服务器地址] [应用路径]

set -e

# 配置
SERVER_USER=${1:-"www-data"}
SERVER_HOST=${2:-"your-server.com"}
APP_PATH=${3:-"/www/goravel-admin"}

echo "开始部署..."

# 1. 编译应用
echo "正在编译应用..."
go build -tags overseer -o main .
if [ $? -ne 0 ]; then
    echo "编译失败！"
    exit 1
fi

# 2. 上传到服务器临时目录
echo "正在上传到服务器..."
scp main ${SERVER_USER}@${SERVER_HOST}:/tmp/goravel-admin-main

# 3. 在服务器上执行热更新
echo "正在执行热更新..."
ssh ${SERVER_USER}@${SERVER_HOST} << EOF
set -e
cd ${APP_PATH}

# 备份当前版本
if [ -f main ]; then
    mv main main.backup.\$(date +%Y%m%d_%H%M%S)
fi

# 替换为新版本
mv /tmp/goravel-admin-main main
chmod +x main
chown www-data:www-data main

# 触发热更新
sudo systemctl reload goravel-admin

# 等待服务启动
sleep 2

# 检查服务状态
if sudo systemctl is-active --quiet goravel-admin; then
    echo "✅ 热更新成功！"
    sudo systemctl status goravel-admin --no-pager -l
else
    echo "❌ 热更新失败，正在回滚..."
    if [ -f main.backup.* ]; then
        mv main.backup.* main
        sudo systemctl restart goravel-admin
    fi
    exit 1
fi
EOF

echo "部署完成！"

