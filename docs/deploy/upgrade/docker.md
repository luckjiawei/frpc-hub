# Docker 升级

升级 Docker 部署的 Podux，只需拉取新镜像并重建容器，数据通过挂载 volume 自动保留。

## Docker Compose 升级（推荐）

```bash
# 拉取最新镜像
docker compose pull

# 重建并启动容器
docker compose up -d

# 确认运行状态
docker compose ps
```

旧镜像不会自动删除，可手动清理：

```bash
docker image prune -f
```

## docker run 升级

```bash
# 拉取最新镜像
docker pull ghcr.io/luckjiawei/podux:latest

# 停止并删除旧容器（数据存储在 volume 中，不会丢失）
docker stop podux
docker rm podux

# 启动新容器
docker run -d \
  --name podux \
  --restart unless-stopped \
  -p 8090:8090 \
  -v ./podux-data:/app/pb_data \
  ghcr.io/luckjiawei/podux:latest
```

## 升级到指定版本

如需升级到特定版本而非 `latest`，修改镜像标签：

```bash
# Docker Compose：设置环境变量
VERSION=v1.2.0 docker compose up -d

# 或编辑 docker-compose.yml 中的 image 字段
image: ghcr.io/luckjiawei/podux:v1.2.0
```

## 回滚

如果升级后出现问题，切换回旧版本：

```bash
# 停止当前容器
docker compose down

# 修改 docker-compose.yml 中的版本号为旧版本，然后启动
VERSION=v1.1.0 docker compose up -d
```

::: tip
数据存储在 volume `podux-data` 中，回滚容器不影响已有数据。
:::
