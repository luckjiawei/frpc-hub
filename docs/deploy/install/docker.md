# Docker 安装

使用 Docker 部署 Podux，适合容器化环境或希望快速隔离部署的场景。

## 前置条件

- Docker 20.10+
- （可选）Docker Compose v2

## 镜像信息

官方镜像托管在 GitHub Container Registry：

```text
ghcr.io/luckjiawei/podux:latest
```

支持以下架构：`linux/amd64`、`linux/arm64`、`linux/arm/v7`

## 快速启动

```bash
docker run -d \
  --name podux \
  --restart unless-stopped \
  -p 8090:8090 \
  -v ./podux-data:/app/pb_data \
  ghcr.io/luckjiawei/podux:latest
```

| 参数 | 说明 |
| --- | --- |
| `-p 8090:8090` | 将容器端口映射到宿主机 |
| `-v ./podux-data:/app/pb_data` | 持久化数据目录 |
| `--restart unless-stopped` | 宿主机重启后自动恢复 |

启动后访问 `http://localhost:8090` 进入主界面。

## Docker Compose（推荐）

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  podux:
    image: ghcr.io/luckjiawei/podux:${VERSION:-latest}
    container_name: podux
    restart: unless-stopped

    ports:
      - "8090:8090"

    volumes:
      - podux-data:/app/pb_data

    environment:
      - TZ=Asia/Shanghai

    # Resource limits (optional)
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 256M

    # Network mode (optional)
    # Uncomment if you need host network for better frpc performance
    # network_mode: host

    # Logging
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8090/api/health"]
      interval: 30s
      timeout: 3s
      start_period: 10s
      retries: 3

volumes:
  podux-data:
    driver: local

networks:
  default:
    name: podux-network
```

启动：

```bash
docker compose up -d
```

查看日志：

```bash
docker compose logs -f podux
```

## 数据持久化

容器内数据存储在 `/app/pb_data`，包含数据库文件和配置信息。**必须挂载此目录**，否则容器重建后数据丢失。

Docker Compose 配置使用具名 volume `podux-data`，数据由 Docker 统一管理：

```bash
# 查看 volume 位置
docker volume inspect podux-data
```

## 自定义端口

如需修改监听端口，同步修改宿主机映射端口和容器内端口：

```bash
docker run -d \
  --name podux \
  --restart unless-stopped \
  -p 18090:18090 \
  -v ./podux-data:/app/pb_data \
  ghcr.io/luckjiawei/podux:latest \
  serve --http 0.0.0.0:18090
```

## 常用命令

```bash
# 停止
docker compose stop podux

# 重启
docker compose restart podux

# 查看日志
docker compose logs -f podux

# 进入容器
docker compose exec podux sh
```

