#!/bin/sh
# Docker 容器启动脚本
# 在应用启动前执行数据库迁移

set -e

echo "=== 容器启动脚本 ==="

# 检查是否需要执行迁移
# 可以通过环境变量控制
SKIP_MIGRATE="${SKIP_MIGRATE:-false}"

if [ "$SKIP_MIGRATE" != "true" ]; then
    echo "执行数据库迁移..."
    if /www/main artisan migrate; then
        echo "✓ 数据库迁移完成"
    else
        echo "✗ 数据库迁移失败，容器退出"
        exit 1
    fi
else
    echo "跳过数据库迁移 (SKIP_MIGRATE=true)"
fi

# 启动应用
echo "启动应用..."
exec /www/main "$@"

