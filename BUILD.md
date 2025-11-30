#### 使用 systemd 部署（推荐）


```bash
# 常规编译
go build .
```
```bash
# 普通编译（当前平台）
go run . artisan build
```

```bash
# 基础静态编译（当前平台）
go build --ldflags "-extldflags -static" -o main .
# 去除符号表和调试信息
go build -ldflags "-extldflags -static -s -w" -o main .
```

```bash
# Linux 服务器交叉编译（在 Windows/Mac 上编译 Linux 版本）

# Windows PowerShell:
$env:GOOS="linux"; $env:GOARCH="amd64"; go build --ldflags "-extldflags -static" -o main .

# Windows CMD (需要分开执行):
set GOOS=linux
set GOARCH=amd64
go build --ldflags "-extldflags -static" -o main .

# 还原
SET GOOS=windows

# Linux/Mac:
GOOS=linux GOARCH=amd64 go build --ldflags "-extldflags -static" -o main .
```

**重要提示：**
- 如果在 Windows 上编译，但要在 Linux 服务器上运行，必须使用交叉编译
- 使用 `GOOS=linux GOARCH=amd64` 指定目标平台
- 如果服务器是 ARM 架构，使用 `GOARCH=arm64`


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

