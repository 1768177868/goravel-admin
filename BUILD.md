# 构建说明

## 多环境配置

为了避免每次构建都要手动修改 `.env` 文件，项目支持使用不同的环境配置文件。

### 1. 创建环境配置文件

在项目根目录创建不同环境的 `.env` 文件：

- `.env.local` - 本地开发环境配置
- `.env.production` - 生产环境配置
- `.env.staging` - 测试环境配置

### 2. 使用构建脚本

构建脚本支持跨平台编译，默认生成 Linux 二进制文件，适合部署到服务器。

#### 脚本参数

```
build.sh [env_file] [output_name] [target_os] [target_arch]
build.bat [env_file] [output_name] [target_os] [target_arch]
```

- `env_file`: 环境配置文件（如 `.env.production`）
- `output_name`: 输出文件名（默认：`main`）
- `target_os`: 目标操作系统（默认：`linux`）
- `target_arch`: 目标架构（默认：`amd64`）

#### Linux/Mac

```bash
# 给脚本添加执行权限
chmod +x build.sh

# 使用生产配置构建 Linux 二进制文件（默认）
./build.sh .env.production
# 输出: main (Linux 二进制文件)

# 使用本地配置构建
./build.sh .env.local

# 指定输出文件名
./build.sh .env.production app

# 生成 Windows 可执行文件
./build.sh .env.production app.exe windows amd64

# 生成 macOS 二进制文件
./build.sh .env.production app darwin amd64

# 生成 Linux ARM64 二进制文件
./build.sh .env.production app linux arm64

# 使用默认 .env 构建
./build.sh
```

#### Windows

```batch
# 使用生产配置构建 Linux 二进制文件（默认）
build.bat .env.production
# 输出: main (Linux 二进制文件)

# 使用本地配置构建
build.bat .env.local

# 指定输出文件名
build.bat .env.production main

# 生成 Windows 可执行文件
build.bat .env.production main.exe windows amd64

# 生成 macOS 二进制文件
build.bat .env.production main darwin amd64

# 生成 Linux ARM64 二进制文件
build.bat .env.production main linux arm64

# 使用默认 .env 构建
build.bat
```

### 3. 跨平台编译说明

构建脚本默认生成 **Linux 二进制文件**，适合部署到 Linux 服务器。脚本会自动：

1. **备份本地 .env 文件**：如果存在 `.env`，会先备份为 `.env.bak`
2. **切换环境配置**：将指定的环境文件复制为 `.env`
3. **设置编译环境**：临时设置 `GOOS`、`GOARCH`、`CGO_ENABLED` 环境变量
4. **执行构建**：使用静态链接编译
5. **恢复环境**：构建完成后自动恢复 `.env` 文件和 Go 环境变量

#### 支持的平台

- **Linux**: `linux/amd64`, `linux/arm64`
- **Windows**: `windows/amd64`
- **macOS**: `darwin/amd64`, `darwin/arm64`

### 4. 直接使用 Go 命令（不推荐）

如果你不想使用构建脚本，也可以手动操作：

```bash
# 备份本地 .env
cp .env .env.bak

# 复制生产环境配置
cp .env.production .env

# 设置跨平台编译环境变量
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

# 构建 Linux 二进制文件
go build --ldflags "-extldflags -static" -o main .

# 恢复本地配置
mv .env.bak .env

# 恢复 Go 环境变量（如果需要）
unset GOOS GOARCH CGO_ENABLED
```

### 5. 环境文件示例

#### .env.local (本地开发)
```env
APP_ENV=local
APP_DEBUG=true
APP_URL=http://localhost:3000
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=goravel_local
DB_USERNAME=root
DB_PASSWORD=
```

#### .env.production (生产环境)
```env
APP_ENV=production
APP_DEBUG=false
APP_URL=https://your-domain.com
DB_CONNECTION=mysql
DB_HOST=your-db-host
DB_PORT=3306
DB_DATABASE=goravel_prod
DB_USERNAME=your-username
DB_PASSWORD=your-password
```

### 6. 注意事项

1. **不要提交敏感信息**：`.env.local`、`.env.production`、`.env.bak` 等文件已在 `.gitignore` 中，不会被提交到 Git
2. **创建 .env.example**：可以创建一个 `.env.example` 文件作为模板，包含所有配置项但不包含敏感信息
3. **自动备份和恢复**：构建脚本会自动备份本地 `.env` 文件为 `.env.bak`，构建完成后自动恢复
4. **环境变量自动还原**：构建脚本会自动还原 `GOOS`、`GOARCH`、`CGO_ENABLED` 环境变量，不会影响后续的 Go 命令
5. **默认生成 Linux 二进制**：默认生成 Linux 二进制文件（`main`），适合部署到服务器
6. **静态链接编译**：使用 `-extldflags -static` 进行静态链接，生成的二进制文件不依赖系统库

### 7. 在 CI/CD 中使用

在持续集成/部署中，可以通过环境变量或构建参数指定环境文件：

```bash
# 示例：在 CI 中构建生产版本（Linux 二进制文件）
./build.sh .env.production

# 示例：在 CI 中构建并指定输出文件名
./build.sh .env.production goravel-admin

# 示例：在 Windows CI 中构建 Linux 二进制文件
build.bat .env.production main linux amd64
```

### 8. Docker 构建

如果使用 Docker，可以在构建时指定环境文件：

```dockerfile
# 在 Dockerfile 中
ARG ENV_FILE=.env.production
COPY ${ENV_FILE} .env
```

构建时：
```bash
docker build --build-arg ENV_FILE=.env.production -t your-app .
```

### 9. 部署二进制文件

构建完成后，将生成的二进制文件上传到服务器即可运行：

```bash
# 上传到服务器
scp main user@server:/path/to/app/

# 在服务器上运行
chmod +x main
./main
```

**注意**：确保服务器上有所需的配置文件（`.env`）和其他必要的文件（如 `database/`、`storage/`、`resources/` 等）。

