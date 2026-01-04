#!/bin/bash
# 部署配置文件
# 在这里统一配置端口，所有脚本都会读取此配置

# 蓝绿部署端口配置
export BLUE_PORT="${BLUE_PORT:-3000}"      # 蓝环境端口（宿主机端口）
export GREEN_PORT="${GREEN_PORT:-3001}"    # 绿环境端口（宿主机端口）
export CONTAINER_PORT="${CONTAINER_PORT:-3000}"  # 容器内部端口（应用监听端口）

# 其他配置
export DEPLOY_DIR="${DEPLOY_DIR:-/www/goravel-admin}"
export NGINX_CONF="${NGINX_CONF:-/etc/nginx/sites-available/goravel-admin}"

