# 端口配置改进说明

## ✅ 改进内容

现在所有端口配置统一在 **`scripts/deploy/deploy-config.sh`** 文件中管理，只需修改一个地方即可。

## 📝 修改端口的方法

### 方法一：修改配置文件（推荐）

编辑 `scripts/deploy/deploy-config.sh`：

```bash
vim scripts/deploy/deploy-config.sh

# 修改为：
export BLUE_PORT="${BLUE_PORT:-4000}"      # 蓝环境端口
export GREEN_PORT="${GREEN_PORT:-4001}"    # 绿环境端口
export CONTAINER_PORT="${CONTAINER_PORT:-3000}"  # 容器内部端口（通常不改）
```

### 方法二：使用环境变量（临时）

```bash
export BLUE_PORT=4000
export GREEN_PORT=4001
./scripts/deploy/docker-blue-green.sh
```

## 🔄 自动应用的文件

修改配置后，以下文件会自动使用新端口：

1. ✅ `docker-compose.blue.yml` - 蓝环境配置
2. ✅ `docker-compose.green.yml` - 绿环境配置
3. ✅ `scripts/deploy/docker-blue-green.sh` - 部署脚本
4. ✅ `scripts/deploy/rollback.sh` - 回滚脚本
5. ✅ Nginx 配置更新（如果配置了）

## 📋 配置变量说明

- **BLUE_PORT**: 蓝环境在宿主机上映射的端口（默认 3000）
- **GREEN_PORT**: 绿环境在宿主机上映射的端口（默认 3001）
- **CONTAINER_PORT**: 容器内部应用监听的端口（默认 3000，通常不需要修改）

## ⚠️ 注意事项

1. 修改端口后需要确保端口未被占用
2. 如果使用 Nginx，需要更新 Nginx 配置中的 upstream
3. 修改端口后需要重新构建和启动容器

## 🔍 检查端口占用

```bash
# 检查端口是否被占用
netstat -tlnp | grep 4000
ss -tlnp | grep 4000
```

