# Docker Installation

Deploy Podux with Docker — ideal for containerized environments or quick isolated deployments.

## Prerequisites

- Docker 20.10+
- Docker Compose v2 (optional)

## Image

The official image is hosted on GitHub Container Registry:

```text
ghcr.io/luckjiawei/podux:latest
```

Supported architectures: `linux/amd64`, `linux/arm64`, `linux/arm/v7`

## Quick Start

```bash
docker run -d \
  --name podux \
  --restart unless-stopped \
  -p 8090:8090 \
  -v ./podux-data:/app/pb_data \
  ghcr.io/luckjiawei/podux:latest
```

| Flag | Description |
| --- | --- |
| `-p 8090:8090` | Map container port to host |
| `-v ./podux-data:/app/pb_data` | Persist data directory |
| `--restart unless-stopped` | Auto-recover after host reboot |

After starting, visit `http://localhost:8090`.

## Docker Compose (Recommended)

Create a `docker-compose.yml` file:

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

Start:

```bash
docker compose up -d
```

View logs:

```bash
docker compose logs -f podux
```

## Data Persistence

Container data is stored in `/app/pb_data`, which includes the database and configuration. **This directory must be mounted**, otherwise data will be lost when the container is recreated.

Docker Compose uses a named volume `podux-data` managed by Docker:

```bash
# Inspect volume location
docker volume inspect podux-data
```

## Custom Port

To change the listening port, update both the host mapping and the container port:

```bash
docker run -d \
  --name podux \
  --restart unless-stopped \
  -p 18090:18090 \
  -v ./podux-data:/app/pb_data \
  ghcr.io/luckjiawei/podux:latest \
  serve --http 0.0.0.0:18090
```

## Common Commands

```bash
# Stop
docker compose stop podux

# Restart
docker compose restart podux

# View logs
docker compose logs -f podux

# Enter container shell
docker compose exec podux sh
```

## Next Steps

- [Docker Upgrade](/en/deploy/upgrade/docker) — Upgrade to a new version
- [Getting Started](/en/guide/getting-started) — Add servers and proxies
