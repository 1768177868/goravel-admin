#!/bin/bash

# 构建脚本 - 支持指定不同的 .env 文件
# 使用方法: ./build.sh [env_file] [output_name] [target_os] [target_arch]
# 示例: 
#   ./build.sh .env.production           # 使用生产配置，生成 Linux 二进制文件 (main)
#   ./build.sh .env.local main linux amd64  # 指定所有参数
#   ./build.sh .env.production main windows amd64  # 生成 Windows 可执行文件
#   ./build.sh                           # 默认使用 .env，生成 Linux 二进制文件

ENV_FILE="${1:-.env}"
OUTPUT_NAME="${2:-main}"
TARGET_OS="${3:-linux}"
TARGET_ARCH="${4:-amd64}"

echo "=========================================="
echo "构建配置"
echo "=========================================="
echo "环境文件: $ENV_FILE"
echo "输出文件: $OUTPUT_NAME"
echo "目标系统: $TARGET_OS"
echo "目标架构: $TARGET_ARCH"
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

# 保存原始的 Go 环境变量
ORIGINAL_GOOS="${GOOS:-}"
ORIGINAL_GOARCH="${GOARCH:-}"
ORIGINAL_CGO_ENABLED="${CGO_ENABLED:-}"

# 设置跨平台编译环境变量
export GOOS="$TARGET_OS"
export GOARCH="$TARGET_ARCH"
export CGO_ENABLED=0

# 执行构建（Linux 二进制文件）
go build --ldflags "-extldflags -static" -o "$OUTPUT_NAME" .

BUILD_RESULT=$?

# 恢复原始的 Go 环境变量
if [ -n "$ORIGINAL_GOOS" ]; then
    export GOOS="$ORIGINAL_GOOS"
else
    unset GOOS
fi

if [ -n "$ORIGINAL_GOARCH" ]; then
    export GOARCH="$ORIGINAL_GOARCH"
else
    unset GOARCH
fi

if [ -n "$ORIGINAL_CGO_ENABLED" ]; then
    export CGO_ENABLED="$ORIGINAL_CGO_ENABLED"
else
    unset CGO_ENABLED
fi

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

