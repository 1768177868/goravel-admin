#!/bin/bash

# 构建脚本 - 支持指定不同的 .env 文件
# 使用方法: ./build.sh [env_file]
# 示例: 
#   ./build.sh .env.local      # 使用本地配置
#   ./build.sh .env.production # 使用生产配置
#   ./build.sh                 # 默认使用 .env

ENV_FILE="${1:-.env}"
OUTPUT_NAME="${2:-main}"

echo "=========================================="
echo "构建配置"
echo "=========================================="
echo "环境文件: $ENV_FILE"
echo "输出文件: $OUTPUT_NAME"
echo "=========================================="

# 检查指定的环境文件是否存在
if [ ! -f "$ENV_FILE" ]; then
    echo "错误: 环境文件 $ENV_FILE 不存在！"
    echo "请先创建 $ENV_FILE 文件"
    exit 1
fi

# 如果存在 .env 文件，先备份为 .env.bak
if [ -f ".env" ]; then
    cp .env .env.bak
    echo "已备份 .env 为 .env.bak"
fi

# 复制指定的环境文件为 .env（构建时使用）
cp "$ENV_FILE" .env

echo "已复制 $ENV_FILE 为 .env"
echo "开始构建..."

# 执行构建
go build --ldflags "-extldflags -static" -o "$OUTPUT_NAME" .

BUILD_RESULT=$?

# 构建完成后，恢复原来的 .env 文件
if [ -f ".env.bak" ]; then
    mv .env.bak .env
    echo "已恢复 .env.bak 为 .env"
fi

if [ $BUILD_RESULT -eq 0 ]; then
    echo "=========================================="
    echo "构建成功！"
    echo "输出文件: $OUTPUT_NAME"
    echo "=========================================="
    exit 0
else
    echo "=========================================="
    echo "构建失败！"
    echo "=========================================="
    exit 1
fi

