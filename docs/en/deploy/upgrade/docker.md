# Docker Upgrade

To upgrade a Docker-deployed Podux, pull the new image and recreate the container. Data in the mounted volume is preserved automatically.

## Docker Compose Upgrade (Recommended)

```bash
# Pull the latest image
docker compose pull

# Recreate and start the container
docker compose up -d

# Confirm running status
docker compose ps
```

Old images are not removed automatically. Clean them up manually:

```bash
docker image prune -f
```

## docker run Upgrade

```bash
# Pull the latest image
docker pull ghcr.io/luckjiawei/podux:latest

# Stop and remove the old container (data in volume is preserved)
docker stop podux
docker rm podux

# Start a new container
docker run -d \
  --name podux \
  --restart unless-stopped \
  -p 8090:8090 \
  -v ./podux-data:/app/pb_data \
  ghcr.io/luckjiawei/podux:latest
```

## Upgrade to a Specific Version

To target a specific version instead of `latest`, set the image tag:

```bash
# Docker Compose: set via environment variable
VERSION=v1.2.0 docker compose up -d

# Or edit the image field in docker-compose.yml
image: ghcr.io/luckjiawei/podux:v1.2.0
```

## Rollback

If issues occur after upgrading, switch back to the previous version:

```bash
# Stop the current container
docker compose down

# Set the old version and start
VERSION=v1.1.0 docker compose up -d
```

::: tip
Data is stored in the `podux-data` volume. Rolling back the container does not affect existing data.
:::
