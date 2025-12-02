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

# overseer 版本（支持零停机重启）
go build -tags overseer --ldflags "-extldflags -static" -o main .
# 去除符号表和调试信息
go build -tags overseer --ldflags "-extldflags -static -s -w" -o main .
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

# ========== overseer 版本（推荐，支持零停机重启）==========
# Windows PowerShell:
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -tags overseer --ldflags "-extldflags -static" -o main .

# Windows CMD (需要分开执行):
set GOOS=linux
set GOARCH=amd64
go build -tags overseer --ldflags "-extldflags -static" -o main .

# Linux/Mac:
GOOS=linux GOARCH=amd64 go build -tags overseer --ldflags "-extldflags -static" -o main .

# 还原环境变量（Windows CMD）
SET GOOS=windows
SET GOARCH=amd64
```

**重要提示：**
- 如果在 Windows 上编译，但要在 Linux 服务器上运行，必须使用交叉编译
- 使用 `GOOS=linux GOARCH=amd64` 指定目标平台
- 如果服务器是 ARM 架构，使用 `GOARCH=arm64`
- **推荐使用 overseer 版本**：添加 `-tags overseer` 参数，支持零停机重启
- 普通版本和 overseer 版本可以共存，通过构建标签区分


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
# 上传服务文件
scp scripts/systemd/goravel-admin.service user@server:/tmp/

# 在服务器上安装服务文件
sudo cp /tmp/goravel-admin.service /etc/systemd/system/
```

编辑服务文件，根据实际情况修改路径和用户：

```bash
sudo nano /etc/systemd/system/goravel-admin.service
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

1. **优化 systemd 配置**（推荐）
   
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

2. **使用热更新库**（推荐，真正的零中断）

   除了 endless，还有其他热更新库可以选择：

   **方案 A：overseer（推荐，更现代）**
   
   `overseer` 是一个更现代的热更新库，功能更强大：
   ```bash
   go get github.com/jpillora/overseer
   ```
   
   特点：
   - ✅ 支持程序化 API，更灵活
   - ✅ 内置健康检查
   - ✅ 支持自定义重启条件
   - ✅ 更好的错误处理
   
   使用示例：
   ```go
   package main
   
   import (
       "github.com/jpillora/overseer"
       "github.com/goravel/framework/facades"
       "goravel/bootstrap"
   )
   
   func main() {
       overseer.Run(overseer.Config{
           Program: prog,
           Address: ":3000",
       })
   }
   
   func prog(state overseer.State) {
       bootstrap.Boot()
       // 启动服务...
   }
   ```

   **方案 B：grace**
   
   `grace` 是另一个类似 endless 的库：
   ```bash
   go get github.com/facebookgo/grace
   ```
   
   使用方式与 endless 类似，但 API 略有不同。

   **方案 C：使用 systemd socket activation**
   
   通过 systemd 的 socket activation 实现零停机：
   ```ini
   # /etc/systemd/system/goravel-admin.socket
   [Socket]
   ListenStream=3000
   Accept=yes
   
   [Install]
   WantedBy=sockets.target
   ```
   
   优点：
   - ✅ 系统级支持，无需修改代码
   - ✅ 真正的零中断
   - ✅ 自动管理进程
   
   缺点：
   - ⚠️ 配置较复杂
   - ⚠️ 需要应用支持 socket activation

3. **使用零停机部署方案**（生产环境最可靠）
   
   通过 Nginx 反向代理 + 多实例部署实现零停机：
   ```bash
   # 方案：部署两个实例，通过 Nginx 负载均衡
   # 1. 启动新版本实例（使用不同端口，如 3001）
   # 2. 更新 Nginx 配置，将流量切换到新实例
   # 3. 等待旧实例请求完成后停止旧实例
   ```
   
   或者使用蓝绿部署：
   ```bash
   # 1. 部署新版本到备用目录
   # 2. 启动新版本服务（不同端口）
   # 3. 健康检查通过后，切换 Nginx upstream
   # 4. 停止旧版本服务
   ```

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

项目已包含 systemd 服务文件模板：`scripts/systemd/goravel-admin.service`

主要特性：
- **自动重启**：服务异常退出时自动重启
- **日志管理**：日志输出到 systemd journal
- **安全设置**：限制文件系统访问权限
- **资源限制**：设置文件描述符限制
- **依赖管理**：等待网络和数据库服务启动
- **优雅关闭**：支持优雅关闭，确保正在处理的请求完成

##### overseer + systemd 单机热更新方案（推荐）

这是单机环境下实现零停机重启的最佳方案，结合了 overseer 的热更新能力和 systemd 的服务管理。

**步骤 1：安装 overseer**

```bash
go get github.com/jpillora/overseer
```

**步骤 2：使用 overseer 版本的 main.go**

项目已提供 `main_overseer.go` 文件，使用构建标签来区分：

```bash
# 编译 overseer 版本
go build -tags overseer -o main .

# 或者直接使用（开发环境）
go run -tags overseer .
```

**步骤 3：配置 systemd 服务文件**

在 `scripts/systemd/goravel-admin.service` 中已包含 overseer 支持：

```ini
[Service]
# ... 其他配置 ...

# 支持 overseer 零停机重启
ExecReload=/bin/kill -HUP $MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
```

**步骤 4：部署和启动**

```bash
# 1. 编译应用（使用 overseer 版本）
go build -tags overseer -o main .

# 2. 上传到服务器
scp main user@server:/www/goravel-admin/

# 3. 设置权限
sudo chmod +x /www/goravel-admin/main

# 4. 安装 systemd 服务
sudo cp scripts/systemd/goravel-admin.service /etc/systemd/system/
sudo systemctl daemon-reload

# 5. 启动服务
sudo systemctl start goravel-admin
sudo systemctl enable goravel-admin
```

**步骤 5：使用热更新**

**重要：热更新流程**

**关键点：必须先替换二进制文件，然后才能用 `reload` 热重启！**

1. **重新编译代码**（在开发机器上）：
   ```bash
   go build -tags overseer -o main .
   ```

2. **上传新版本到服务器**：
   ```bash
   scp main user@server:/www/goravel-admin/main.new
   ```

3. **替换可执行文件并触发热更新**：
   ```bash
   # 在服务器上执行
   cd /www/goravel-admin
   mv main main.old          # 备份旧版本
   mv main.new main          # 替换为新版本（重要：必须先替换文件！）
   chmod +x main
   
   # 触发零停机热更新（此时二进制文件已更新，可以用 reload）
   sudo systemctl reload goravel-admin
   ```

**`reload` vs `restart` 的区别：**

| 操作 | 使用场景 | 是否有中断 | 说明 |
|------|----------|------------|------|
| `systemctl reload` | **二进制文件已更新** | ✅ 零中断 | overseer 检测到新文件，fork 子进程实现热更新 |
| `systemctl restart` | 二进制文件未更新，或首次启动 | ❌ 有中断（2-5秒） | 停止旧进程，启动新进程 |

**重要提示：**
- ✅ **如果二进制文件已更新**：使用 `sudo systemctl reload goravel-admin`（零停机）
- ❌ **如果二进制文件未更新**：`reload` 不会加载新代码，必须用 `restart`（有中断）
- ⚠️ **代码更改后**：必须先编译、上传、替换文件，然后才能 `reload`

**或者使用一键部署脚本：**

项目已提供 `scripts/deploy_overseer.sh` 脚本：

```bash
# 使用默认配置
./scripts/deploy_overseer.sh

# 或指定参数
./scripts/deploy_overseer.sh www-data your-server.com /www/goravel-admin
```

脚本会自动完成：
1. 编译应用（使用 overseer 标签）
2. 上传到服务器
3. 备份旧版本
4. 替换新版本
5. 触发热更新
6. 检查服务状态
7. 失败时自动回滚

**热更新命令：**

```bash
# ⚠️ 重要：必须先替换二进制文件，然后才能用 reload！

# 完整流程：
# 1. 编译新版本
go build -tags overseer -o main .

# 2. 上传并替换文件
scp main user@server:/www/goravel-admin/main.new
ssh user@server "cd /www/goravel-admin && mv main main.old && mv main.new main && chmod +x main"

# 3. 热重启（零停机）
sudo systemctl reload goravel-admin  # ✅ 零停机，因为文件已更新

# 如果文件未更新，reload 不会加载新代码，必须用 restart：
sudo systemctl restart goravel-admin  # ❌ 有中断，但会加载新代码
```

**常见错误：**
```bash
# ❌ 错误：代码改了，但只执行 reload，文件没更新
sudo systemctl reload goravel-admin  # 代码还是旧的！

# ✅ 正确：先更新文件，再 reload
mv main.new main && sudo systemctl reload goravel-admin  # 代码已更新，零停机
```

**工作原理：**

1. **初始启动**：systemd 启动应用，overseer 监听指定端口
2. **触发更新**：发送 `SIGHUP` 信号给主进程
3. **零停机切换**：
   - overseer 收到信号后 fork 子进程
   - 子进程继承 socket 文件描述符
   - 子进程开始处理新请求（零中断）
   - 父进程停止接受新连接，等待旧请求完成
   - 父进程退出，完成升级

**验证热更新：**

```bash
# 查看进程信息
ps aux | grep main

# 查看服务状态
sudo systemctl status goravel-admin

# 查看日志
sudo journalctl -u goravel-admin -f

# 测试 API（在更新过程中应该无中断）
watch -n 1 'curl -s http://localhost:3000/api/admin/health'
```

**优势：**

- ✅ **真正的零中断**：新进程启动时，旧进程仍在处理请求
- ✅ **无缝切换**：socket 文件描述符直接传递
- ✅ **简单易用**：只需 `systemctl reload` 即可
- ✅ **自动管理**：systemd 负责进程监控和自动重启
- ✅ **日志完整**：所有日志通过 systemd journal 管理

**注意事项：**

- ⚠️ 确保编译时使用 `-tags overseer` 标签
- ⚠️ 更新代码后需要重新编译并上传
- ⚠️ 建议在生产环境充分测试
- ⚠️ 如果更新失败，overseer 会自动回滚到旧版本

**故障排查：**

```bash
# 如果热更新失败，查看日志
sudo journalctl -u goravel-admin -n 100

# 检查进程
ps aux | grep main

# 手动重启（如果热更新失败）
sudo systemctl restart goravel-admin
```

**关于 `cannot move binary` 警告：**

如果看到以下警告：
```
[overseer] disabled. run failed: cannot move binary (exit status 1)
```

这是**正常现象**，原因：
- overseer 尝试替换正在运行的文件时，文件被占用无法移动
- overseer 会自动回退到普通模式运行
- **服务仍然正常运行**，只是热更新功能暂时不可用

解决方法：
1. **使用正确的部署流程**：先停止服务，替换文件，再启动
2. **或者使用多实例部署**：通过 Nginx 负载均衡实现零停机
3. **或者忽略警告**：如果服务正常运行，可以暂时忽略

##### 其他热更新方案详解

除了 overseer + systemd，还有其他热更新方案可以选择：

**2. endless（经典方案）**

`endless` 是最早的 Go 热更新库之一，稳定可靠。

**安装：**
```bash
go get github.com/fvbock/endless
```

**特点：**
- ✅ 成熟稳定
- ✅ 使用简单
- ⚠️ API 相对简单，功能较少

**3. systemd socket activation（系统级方案）**

通过 systemd 的 socket activation 实现零停机，无需修改代码。

**配置步骤：**

1. 创建 socket 文件 `/etc/systemd/system/goravel-admin.socket`：
```ini
[Unit]
Description=Goravel Admin Socket

[Socket]
ListenStream=3000
Accept=yes

[Install]
WantedBy=sockets.target
```

2. 修改服务文件，添加 `Accept=true`：
```ini
[Service]
Type=notify
ExecStart=/www/goravel-admin/main
StandardOutput=journal
StandardError=journal
```

3. 启用并启动：
```bash
sudo systemctl enable goravel-admin.socket
sudo systemctl start goravel-admin.socket
```

**优点：**
- ✅ 系统级支持，无需修改代码
- ✅ 真正的零中断
- ✅ 自动进程管理

**缺点：**
- ⚠️ 需要应用支持 systemd notify
- ⚠️ 配置较复杂

**4. 容器化方案（现代化推荐）**

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
| **overseer** | ✅ 是 | 中 | 单机部署，需要热更新 | ⭐⭐⭐⭐⭐ |
| **endless** | ✅ 是 | 中 | 单机部署，需要热更新 | ⭐⭐⭐⭐ |
| **systemd socket** | ✅ 是 | 高 | 系统级管理 | ⭐⭐⭐ |
| **Nginx + 多实例** | ✅ 是 | 中 | 生产环境，最可靠 | ⭐⭐⭐⭐⭐ |
| **Docker/K8s** | ✅ 是 | 高 | 容器化环境 | ⭐⭐⭐⭐⭐ |
| **Supervisor** | ⚠️ 需配合 | 低 | 简单场景 | ⭐⭐⭐ |

**推荐选择：**
- **开发环境**：使用 overseer 或 endless
- **生产环境（单机）**：使用 overseer + systemd
- **生产环境（多机）**：使用 Nginx + 多实例 或 Kubernetes

| 方案 | 零中断 | 复杂度 | 适用场景 | 推荐度 |
|------|--------|--------|----------|--------|
| **overseer** | ✅ 是 | 中 | 单机部署，需要热更新 | ⭐⭐⭐⭐⭐ |
| **endless** | ✅ 是 | 中 | 单机部署，需要热更新 | ⭐⭐⭐⭐ |
| **grace** | ✅ 是 | 中 | 单机部署，需要热更新 | ⭐⭐⭐ |
| **systemd socket** | ✅ 是 | 高 | 系统级管理 | ⭐⭐⭐ |
| **Nginx + 多实例** | ✅ 是 | 中 | 生产环境，最可靠 | ⭐⭐⭐⭐⭐ |
| **Docker/K8s** | ✅ 是 | 高 | 容器化环境 | ⭐⭐⭐⭐⭐ |
| **Supervisor** | ⚠️ 需配合 | 低 | 简单场景 | ⭐⭐⭐ |

**推荐选择：**
- **开发环境**：使用 overseer 或 endless
- **生产环境（单机）**：使用 overseer + systemd
- **生产环境（多机）**：使用 Nginx + 多实例 或 Kubernetes

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

