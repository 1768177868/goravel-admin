#### 使用 systemd 部署（推荐）


```bash
# 常规编译（普通版本）
go build .
```

```bash
# 普通编译（当前平台）
go run . artisan build
```

```bash
# 基础静态编译（当前平台，普通版本）
go build --ldflags "-extldflags -static" -o main .
# 去除符号表和调试信息
go build --ldflags "-extldflags -static -s -w" -o main .

# 生产环境推荐使用静态编译
go build --ldflags "-extldflags -static -s -w" -o main .
```

```bash
# Linux 服务器交叉编译（在 Windows/Mac 上编译 Linux 版本）

# ========== 普通版本 ==========
# Windows PowerShell:
$env:GOOS="linux"; $env:GOARCH="amd64"; go build --ldflags "-extldflags -static" -o main .

# Windows CMD (需要分开执行):
set GOOS=linux
set GOARCH=amd64
go build --ldflags "-extldflags -static" -o main .

# Linux/Mac:
GOOS=linux GOARCH=amd64 go build --ldflags "-extldflags -static" -o main .

# ========== 生产环境静态编译（推荐）==========
# Windows PowerShell:
$env:GOOS="linux"; $env:GOARCH="amd64"; go build --ldflags "-extldflags -static -s -w" -o main .

# Windows CMD (需要分开执行):
set GOOS=linux
set GOARCH=amd64
go build --ldflags "-extldflags -static -s -w" -o main .

# Linux/Mac:
GOOS=linux GOARCH=amd64 go build --ldflags "-extldflags -static -s -w" -o main .

# 还原环境变量（Windows CMD）
SET GOOS=windows
SET GOARCH=amd64
```

**重要提示：**
- 如果在 Windows 上编译，但要在 Linux 服务器上运行，必须使用交叉编译
- 使用 `GOOS=linux GOARCH=amd64` 指定目标平台
- 如果服务器是 ARM 架构，使用 `GOARCH=arm64`
- **推荐使用静态编译**：添加 `--ldflags "-extldflags -static -s -w"` 参数，生成独立可执行文件


使用 systemd 可以将应用作为系统服务运行，支持开机自启、自动重启、日志管理等功能。

##### 步骤 1：准备部署文件 

在服务器上创建应用目录并上传必要文件：

```bash
# 在服务器上创建应用目录
sudo mkdir -p /www/goravel-admin
sudo chown -R www-data:www-data /www/goravel-admin

# 上传二进制文件
scp main user@server:/www/goravel-admin/

# 上传配置文件和其他必要文件
scp .env user@server:/www/goravel-admin/.env
scp -r storage/ user@server:/www/goravel-admin/
scp -r resources/ user@server:/www/goravel-admin/
scp -r public/ user@server:/www/goravel-admin/
```

##### 步骤 2：配置 systemd 服务

复制 systemd 服务文件到系统目录：

```bash
# 上传服务文件（双实例方案）
scp scripts/systemd/goravel-admin-3000.service user@server:/tmp/
scp scripts/systemd/goravel-admin-3001.service user@server:/tmp/

# 在服务器上安装服务文件
sudo cp /tmp/goravel-admin-3000.service /etc/systemd/system/
sudo cp /tmp/goravel-admin-3001.service /etc/systemd/system/
```

编辑服务文件，根据实际情况修改路径和用户：

```bash
sudo nano /etc/systemd/system/goravel-admin-3000.service
sudo nano /etc/systemd/system/goravel-admin-3001.service
```

主要配置项：
- `User` 和 `Group`：运行服务的用户和组（如 `www-data`）
- `WorkingDirectory`：应用工作目录（如 `/www/goravel-admin`）
- `ExecStart`：二进制文件路径（如 `/www/goravel-admin/main`）
- `ReadWritePaths`：需要写入权限的目录（如 `/www/goravel-admin/storage`）

##### 步骤 3：设置文件权限

```bash
# 设置二进制文件权限
sudo chmod +x /www/goravel-admin/main

# 设置存储目录权限
sudo chmod -R 775 /www/goravel-admin/storage
sudo chown -R www-data:www-data /www/goravel-admin/storage

# 设置配置文件权限（保护敏感信息）
sudo chmod 600 /www/goravel-admin/.env
sudo chown www-data:www-data /www/goravel-admin/.env

```

```bash
# 你也许想将生产环境的 env 文件添加到版本控制中，但又不想将敏感信息暴露出来，这时你可以使用 env:encrypt 命令来加密 env 文件
go run . artisan env:encrypt
# 解密 env
go run . artisan env:decrypt
```

```bash
./main artisan key:generate
```

```bash
# 迁移
./main artisan migrate
# 填充
./main artisan db:seed
# 测试运行 ctrl+c 结束
./main 
```


##### 步骤 4：启动和管理服务

```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start goravel-admin

# 设置开机自启
sudo systemctl enable goravel-admin

# 查看服务状态
sudo systemctl status goravel-admin

# 停止服务
sudo systemctl stop goravel-admin

# 重启服务
sudo systemctl restart goravel-admin

**⚠️ 重要提示：服务重启会有短暂中断**

使用 `systemctl restart` 重启服务时，会有短暂的服务中断（通常 1-3 秒）：
- 应用会收到 `SIGTERM` 信号并执行优雅关闭
- 正在处理的请求会被完成，但新请求会被拒绝
- 旧进程完全退出后，新进程才会启动
- 在此期间，服务不可用

**最小化中断的方法：**

1. **优化 systemd 配置**
   
   在服务文件中添加以下配置，确保优雅关闭：
   ```ini
   KillMode=mixed
   KillSignal=SIGTERM
   TimeoutStopSec=30
   ```
   这样 systemd 会：
   - 先发送 `SIGTERM` 信号给主进程，允许优雅关闭
   - 等待最多 30 秒让应用完成正在处理的请求
   - 如果超时仍未退出，再发送 `SIGKILL` 强制终止

2. **使用零停机部署方案**（⭐ 推荐，生产环境最可靠）
   
   通过 Nginx 反向代理 + 双实例部署实现真正的零停机：
   
   **工作原理：**
   - 运行两个应用实例（端口 3000 和 3001）
   - Nginx 负载均衡到两个实例
   - 部署新版本时：
     1. 启动新版本到备用端口（如 3001）
     2. 健康检查通过后，Nginx 自动切换到新实例
     3. 停止旧实例
     4. 实现真正的零停机
   
   详细配置请参考 `ZERO_DOWNTIME.md` 文档。

4. **使用容器化方案**（现代化推荐）
   
   **Docker + Docker Compose：**
   ```yaml
   # docker-compose.yml
   version: '3'
   services:
     app:
       build: .
       ports:
         - "3000:3000"
       deploy:
         replicas: 2
         update_config:
           parallelism: 1
           delay: 10s
   ```
   
   滚动更新：
   ```bash
   docker-compose up -d --scale app=2 --no-deps --build app
   ```
   
   **Kubernetes：**
   ```yaml
   # deployment.yaml
   apiVersion: apps/v1
   kind: Deployment
   spec:
     replicas: 2
     strategy:
       type: RollingUpdate
       rollingUpdate:
         maxSurge: 1
         maxUnavailable: 0  # 零停机
   ```
   
   优点：
   - ✅ 真正的零停机滚动更新
   - ✅ 自动健康检查
   - ✅ 自动回滚
   - ✅ 资源隔离

5. **使用进程管理器**
   
   **Supervisor：**
   ```ini
   [program:goravel-admin]
   command=/www/goravel-admin/main
   numprocs=2  # 多进程
   autostart=true
   autorestart=true
   ```
   
   通过多进程 + Nginx 负载均衡实现零停机。

6. **在低峰期重启**
   
   选择业务低峰期进行服务重启，减少对用户的影响。

# 查看服务日志
sudo journalctl -u goravel-admin -f

# 查看最近 100 行日志
sudo journalctl -u goravel-admin -n 100

# 查看今天的日志
sudo journalctl -u goravel-admin --since today
```

##### 步骤 5：验证部署

```bash
# 检查服务是否运行
sudo systemctl is-active goravel-admin

# 检查端口是否监听（假设端口是 3000）
sudo netstat -tlnp | grep 3000
# 或使用 ss 命令
sudo ss -tlnp | grep 3000

# 测试 API 接口（本地）
curl http://localhost:3000/api/admin/health
```

**重要：如果无法从外部 IP 访问**

默认配置中 `APP_HOST=127.0.0.1` 只允许本地访问。要允许外部访问，需要修改 `.env` 文件：

```env
# 修改为 0.0.0.0 允许所有网络接口访问
APP_HOST=0.0.0.0
APP_PORT=3000
```

然后重启服务：
```bash
sudo systemctl restart goravel-admin
```

**检查防火墙设置：**

```bash
# CentOS/RHEL 系统
sudo firewall-cmd --list-ports
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --reload

# Ubuntu/Debian 系统
sudo ufw status
sudo ufw allow 3000/tcp
sudo ufw reload

# 或者临时关闭防火墙测试（不推荐生产环境）
sudo systemctl stop firewalld  # CentOS/RHEL
sudo ufw disable              # Ubuntu/Debian
```

**验证外部访问：**

```bash
# 在服务器上测试
curl http://服务器IP:3000/api/admin/health

# 或者从其他机器测试
curl http://服务器IP:3000/api/admin/health
```

##### systemd 服务文件示例

项目已包含 systemd 服务文件模板：
- `scripts/systemd/goravel-admin-3000.service` - 端口 3000 实例
- `scripts/systemd/goravel-admin-3001.service` - 端口 3001 实例

主要特性：
- **自动重启**：服务异常退出时自动重启
- **日志管理**：日志输出到 systemd journal
- **安全设置**：限制文件系统访问权限
- **资源限制**：设置文件描述符限制
- **依赖管理**：等待网络和数据库服务启动
- **优雅关闭**：支持优雅关闭，确保正在处理的请求完成

##### Nginx + 双实例零停机部署方案（⭐ 推荐）

这是生产环境最可靠的零停机部署方案，通过 Nginx 负载均衡和双实例部署实现真正的零停机。

**工作原理：**
- 运行两个应用实例（端口 3000 和 3001）
- Nginx 负载均衡到两个实例
- 部署新版本时：
  1. 启动新版本到备用端口（如 3001）
  2. 健康检查通过后，Nginx 自动切换到新实例
  3. 停止旧实例
  4. 实现真正的零停机

**步骤 1：配置双实例 systemd 服务**

项目已提供两个服务文件：
- `scripts/systemd/goravel-admin-3000.service` - 端口 3000
- `scripts/systemd/goravel-admin-3001.service` - 端口 3001

在服务器上安装：

```bash
# 上传服务文件
scp scripts/systemd/goravel-admin-3000.service user@server:/tmp/
scp scripts/systemd/goravel-admin-3001.service user@server:/tmp/

# 在服务器上安装
sudo cp /tmp/goravel-admin-3000.service /etc/systemd/system/
sudo cp /tmp/goravel-admin-3001.service /etc/systemd/system/
sudo systemctl daemon-reload
```

**步骤 2：配置 Nginx 负载均衡**

项目已提供 Nginx 配置文件：`scripts/nginx/goravel-admin.conf`

在服务器上安装：

```bash
# 上传 Nginx 配置
scp scripts/nginx/goravel-admin.conf user@server:/tmp/

# 在服务器上安装
sudo cp /tmp/goravel-admin.conf /etc/nginx/sites-available/goravel-admin
sudo ln -s /etc/nginx/sites-available/goravel-admin /etc/nginx/sites-enabled/
sudo nginx -t
sudo nginx -s reload
```

**步骤 3：启动服务**

```bash
# 启动第一个实例（端口 3000）
sudo systemctl start goravel-admin-3000
sudo systemctl enable goravel-admin-3000

# 检查状态
sudo systemctl status goravel-admin-3000
```

**步骤 4：零停机部署**

使用项目提供的部署脚本：

```bash
# 使用默认配置
./scripts/deploy-zero-downtime.sh

# 或指定参数
./scripts/deploy-zero-downtime.sh www-data your-server.com /www/goravel-admin
```

脚本会自动完成：
1. 编译应用
2. 上传到服务器
3. 备份旧版本
4. 启动新实例到备用端口
5. 健康检查
6. 更新 Nginx 配置
7. 停止旧实例

**优势：**

- ✅ **真正的零停机**：新实例启动时，旧实例仍在处理请求
- ✅ **高可用性**：一个实例故障时，另一个实例继续服务
- ✅ **简单可靠**：无需修改代码，使用标准 systemd 和 Nginx
- ✅ **易于回滚**：如果新版本有问题，可以快速切换回旧实例

**详细文档：**

更多配置细节和故障排查，请参考 `ZERO_DOWNTIME.md` 文档。

##### 其他部署方案

**容器化方案（现代化推荐）**

使用 Docker 或 Kubernetes 实现滚动更新。

**Docker Compose 示例：**
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 10s
        failure_action: rollback
```

**Kubernetes 滚动更新：**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goravel-admin
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0  # 确保零停机
  template:
    spec:
      containers:
      - name: app
        image: goravel-admin:latest
```

**方案对比表：**

| 方案 | 零中断 | 复杂度 | 适用场景 | 推荐度 |
|------|--------|--------|----------|--------|
| **Nginx + 双实例** | ✅ 是 | 中 | 生产环境，最可靠 | ⭐⭐⭐⭐⭐ |
| **Docker/K8s** | ✅ 是 | 高 | 容器化环境 | ⭐⭐⭐⭐⭐ |
| **systemd socket** | ✅ 是 | 高 | 系统级管理 | ⭐⭐⭐ |
| **Supervisor** | ⚠️ 需配合 | 低 | 简单场景 | ⭐⭐⭐ |

**推荐选择：**
- **生产环境（单机/多机）**：使用 Nginx + 双实例（⭐ 推荐）
- **容器化环境**：使用 Docker Compose 或 Kubernetes
- **开发环境**：使用普通 systemd 服务即可

##### 零停机部署方案（生产环境推荐）

对于生产环境，如果需要零停机更新，可以使用以下方案：

**方案 1：Nginx 反向代理 + 双实例部署**

1. 准备两个应用目录：
```bash
/www/goravel-admin-v1  # 当前运行版本
/www/goravel-admin-v2  # 新版本
```

2. 配置 Nginx upstream（支持健康检查）：
```nginx
upstream goravel_backend {
    least_conn;  # 使用最少连接负载均衡
    server 127.0.0.1:3000 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:3001 max_fails=3 fail_timeout=30s backup;
    
    keepalive 32;
}

server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://goravel_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 健康检查
        proxy_next_upstream error timeout http_500 http_502 http_503;
    }
}
```

3. 部署流程：
```bash
# 1. 部署新版本到 v2 目录，使用端口 3001
cd /www/goravel-admin-v2
# 上传新版本文件...
# 修改 .env 中的 APP_PORT=3001

# 2. 启动新版本服务
sudo systemctl start goravel-admin-v2

# 3. 健康检查
curl http://localhost:3001/api/admin/health

# 4. 更新 Nginx 配置，将新实例加入负载均衡
sudo nginx -t  # 测试配置
sudo nginx -s reload  # 重新加载配置

# 5. 等待旧实例请求完成（观察日志）
sudo journalctl -u goravel-admin-v1 -f

# 6. 停止旧版本服务
sudo systemctl stop goravel-admin-v1

# 7. 下次部署时，v1 和 v2 角色互换
```

**方案 2：使用 systemd socket 激活（高级）**

通过 systemd socket 激活可以实现更平滑的切换，但配置较复杂。

**方案 3：容器化部署（Docker/Kubernetes）**

使用容器编排工具可以实现真正的零停机滚动更新：
```bash
# Docker Compose 示例
docker-compose up -d --scale goravel-admin=2 --no-deps --build goravel-admin
```

##### 常见问题排查

1. **服务无法启动**
   ```bash
   # 查看详细错误信息
   sudo journalctl -u goravel-admin -n 50
   
   # 检查二进制文件权限
   ls -l /www/goravel-admin/main
   
   # 检查配置文件
   sudo -u www-data /www/goravel-admin/main --help
   ```

2. **权限问题**
   ```bash
   # 确保存储目录有写入权限
   sudo chown -R www-data:www-data /www/goravel-admin/storage
   sudo chmod -R 775 /www/goravel-admin/storage
   ```

3. **端口被占用**
   ```bash
   # 检查端口占用
   sudo lsof -i :3000
   # 或修改 .env 文件中的端口配置
   ```

4. **数据库连接失败**
   ```bash
   # 检查数据库配置
   cat /www/goravel-admin/.env | grep DB_
   
   # 测试数据库连接
   mysql -h localhost -u username -p database_name
   ```

