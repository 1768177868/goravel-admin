# 构建说明

## 多环境配置

为了避免每次构建都要手动修改 `.env` 文件，项目支持使用不同的环境配置文件。

### 1. 创建环境配置文件

在项目根目录创建不同环境的 `.env` 文件：

- `.env.local` - 本地开发环境配置
- `.env.production` - 生产环境配置
- `.env.staging` - 测试环境配置

### 2. 使用构建脚本

#### Linux/Mac

```bash
# 给脚本添加执行权限
chmod +x build.sh

# 使用本地配置构建
./build.sh .env.local

# 使用生产配置构建
./build.sh .env.production

# 使用默认 .env 构建
./build.sh
```

#### Windows

```batch
# 使用本地配置构建
build.bat .env.local

# 使用生产配置构建
build.bat .env.production

# 使用默认 .env 构建
build.bat
```

### 3. 直接使用 Go 命令（不推荐）

如果你不想使用构建脚本，也可以手动复制文件：

```bash
# 复制生产环境配置
cp .env.production .env

# 构建
go build --ldflags "-extldflags -static" -o main .

# 恢复本地配置（可选）
cp .env.local .env
```

### 4. 环境文件示例

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

### 5. 注意事项

1. **不要提交敏感信息**：`.env.local`、`.env.production` 等文件已在 `.gitignore` 中，不会被提交到 Git
2. **创建 .env.example**：可以创建一个 `.env.example` 文件作为模板，包含所有配置项但不包含敏感信息
3. **构建时自动复制**：构建脚本会自动将指定的环境文件复制为 `.env`，构建完成后不会恢复原文件

### 6. 在 CI/CD 中使用

在持续集成/部署中，可以通过环境变量或构建参数指定环境文件：

```bash
# 示例：在 CI 中构建生产版本
./build.sh .env.production
```

### 7. Docker 构建

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

