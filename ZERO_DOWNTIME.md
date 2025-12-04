# 零停机部署方案

## 问题说明

使用 `overseer` 单实例部署时，由于需要启动新进程并初始化框架（`bootstrap.Boot()`），会有 1-2 秒的中断。这是正常的，因为：

1. 新进程需要 fork
2. 新进程需要初始化框架（数据库连接、配置加载等）
3. 新进程需要启动 HTTP 服务器
4. 在新进程完全准备好之前，会有短暂中断

**注意**：即使使用 `overseer.State.Listener` 传递 socket 文件描述符，由于 Goravel 框架的初始化需要时间，仍然会有短暂中断。

## 真正的零停机方案

### 方案 1：Nginx + 双实例部署（⭐ 推荐，最可靠）

这是生产环境最可靠的零停机方案，被广泛使用。

#### 工作原理

1. 运行两个应用实例（端口 3000 和 3001）
2. Nginx 负载均衡到两个实例
3. 部署新版本时：
   - 启动新版本到备用端口（如 3001）
   - 健康检查通过后，Nginx 自动切换到新实例
   - 停止旧实例
   - 实现真正的零停机

#### 配置步骤

**1. 配置 Nginx 负载均衡**

创建或编辑 `/etc/nginx/sites-available/goravel-admin`：

```nginx
upstream goravel_backend {
    least_conn;  # 使用最少连接负载均衡
    
    # 主实例（当前运行版本）
    server 127.0.0.1:3000 max_fails=3 fail_timeout=30s;
    
    # 备用实例（部署新版本时启用）
    # server 127.0.0.1:3001 max_fails=3 fail_timeout=30s backup;
    
    keepalive 32;
}

server {
    listen 80;
    server_name api.xuancheng888.top;
    
    # 如果需要 HTTPS
    # listen 443 ssl http2;
    # ssl_certificate /path/to/cert.pem;
    # ssl_certificate_key /path/to/key.pem;
    
    client_max_body_size 100M;
    
    location / {
        proxy_pass http://goravel_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # 健康检查：自动切换到备用实例
        proxy_next_upstream error timeout http_500 http_502 http_503 http_504;
        proxy_next_upstream_tries 2;
        proxy_next_upstream_timeout 10s;
    }
}
```

**2. 创建第二个实例的 systemd 服务文件**

`/etc/systemd/system/goravel-admin-3001.service`：

```ini
[Unit]
Description=Goravel Admin Application (Port 3001)
After=network.target mysql.service
Requires=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/www/goravel-admin
Environment="APP_PORT=3001"
ExecStart=/www/goravel-admin/main
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=goravel-admin-3001

KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/www/goravel-admin/storage

LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

**3. 零停机部署流程**

```bash
# 1. 编译新版本
go build -tags overseer --ldflags "-extldflags -static -s -w" -o main .

# 2. 上传到服务器
scp main user@server:/www/goravel-admin/main.new

# 3. 在服务器上执行部署
ssh user@server

# 确定当前运行的端口
CURRENT_PORT=$(systemctl show -p ExecStart goravel-admin | grep -oP 'APP_PORT=\K[0-9]+' || echo "3000")
if [ "$CURRENT_PORT" = "3000" ]; then
    NEW_PORT=3001
    OLD_PORT=3000
else
    NEW_PORT=3000
    OLD_PORT=3001
fi

echo "当前运行端口: $OLD_PORT"
echo "新实例端口: $NEW_PORT"

# 备份并替换文件
cd /www/goravel-admin
mv main main.old.$(date +%Y%m%d_%H%M%S)
mv main.new main
chmod +x main

# 启动新实例
sudo systemctl start goravel-admin-$NEW_PORT

# 等待新实例启动并健康检查
for i in {1..10}; do
    if curl -f -s http://localhost:$NEW_PORT/api/admin/health > /dev/null; then
        echo "✅ 新实例健康检查通过"
        break
    fi
    echo "等待新实例启动... ($i/10)"
    sleep 2
done

# 更新 Nginx 配置，将新端口添加到 upstream
sudo sed -i "s/# server 127.0.0.1:$NEW_PORT/server 127.0.0.1:$NEW_PORT/" /etc/nginx/sites-available/goravel-admin
sudo nginx -t && sudo nginx -s reload

# 等待旧实例请求完成
sleep 5

# 停止旧实例
sudo systemctl stop goravel-admin-$OLD_PORT
sudo systemctl disable goravel-admin-$OLD_PORT

echo "✅ 零停机部署完成"
```

### 方案 2：优化启动时间（减少中断时间）

通过优化框架初始化来减少中断时间：

1. **延迟加载**：只加载必要的服务
2. **连接池预热**：提前建立数据库连接
3. **缓存配置**：减少配置读取时间
4. **预编译模板**：减少模板编译时间

**优化示例**：

```go
// 在 bootstrap.Boot() 之前，可以预先加载一些配置
// 但注意：这需要框架支持

// 减少不必要的中间件
// 使用连接池
// 缓存配置到内存
```



## 快速部署脚本

创建一个自动化脚本来自动完成双实例部署：

```bash
#!/bin/bash
# deploy-zero-downtime.sh

set -e

# 配置
APP_DIR="/www/goravel-admin"
SERVICE_PREFIX="goravel-admin"
PORT1=3000
PORT2=3001

# 确定当前运行的端口
if systemctl is-active --quiet ${SERVICE_PREFIX}-${PORT1}; then
    CURRENT_PORT=$PORT1
    NEW_PORT=$PORT2
else
    CURRENT_PORT=$PORT2
    NEW_PORT=$PORT1
fi

echo "当前运行端口: $CURRENT_PORT"
echo "新实例端口: $NEW_PORT"

# 备份并替换文件
cd $APP_DIR
mv main main.old.$(date +%Y%m%d_%H%M%S)
mv main.new main
chmod +x main

# 启动新实例
echo "启动新实例..."
sudo systemctl start ${SERVICE_PREFIX}-${NEW_PORT}

# 健康检查
echo "等待新实例启动..."
for i in {1..30}; do
    if curl -f -s http://localhost:${NEW_PORT}/api/admin/health > /dev/null 2>&1; then
        echo "✅ 新实例健康检查通过"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ 新实例启动失败"
        exit 1
    fi
    sleep 1
done

# 更新 Nginx 配置
echo "更新 Nginx 配置..."
sudo sed -i "s/# server 127.0.0.1:${NEW_PORT}/server 127.0.0.1:${NEW_PORT}/" /etc/nginx/sites-available/goravel-admin
sudo nginx -t && sudo nginx -s reload

# 等待旧实例请求完成
echo "等待旧实例请求完成..."
sleep 5

# 停止旧实例
echo "停止旧实例..."
sudo systemctl stop ${SERVICE_PREFIX}-${CURRENT_PORT}

echo "✅ 零停机部署完成"
```


