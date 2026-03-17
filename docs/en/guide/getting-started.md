# Getting Started

This guide helps you install and configure Podux in under 5 minutes using Docker.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed
- A public server with frps deployed (or a third-party frp service)
- The machine running Podux must have internet access

::: info Don't have a public server yet?
Check out **[RainYun](https://www.rainyun.com/s=ljw_?podux)** — an affordable and reliable cloud server provider, great for hosting frps and other self-hosted services.
:::

## Step 1: Start the Container

```bash
docker run -d \
  --name podux \
  --restart unless-stopped \
  -p 8090:8090 \
  -v podux-data:/app/pb_data \
  ghcr.io/luckjiawei/podux:latest
```

On successful startup, the terminal will display:

```text
App:   http://0.0.0.0:8090
Admin: http://0.0.0.0:8090/_/
```

::: tip Using Docker Compose
You can also manage it with Docker Compose:

```yaml
services:
  podux:
    image: ghcr.io/luckjiawei/podux:latest
    container_name: podux
    restart: unless-stopped
    ports:
      - "8090:8090"
    volumes:
      - podux-data:/app/pb_data

volumes:
  podux-data:
```

```bash
docker compose up -d
```
:::

## Step 2: Create an Account

Open your browser and navigate to the main interface:

```text
http://localhost:8090
```

On the first visit, you will be guided to create an account. Enter your username and password and submit.

## Step 3: Create a Backend Account (Optional)

If you need access to the backend management panel, navigate to the Admin UI:

```text
http://localhost:8090/_/
```

On the first visit, you will be guided to create a backend account. Enter your email and password and submit.

::: warning Security Notice
If Podux is accessible from the public internet, use a strong password and consider restricting access to port `8090` via your firewall.
:::

## Step 4: Open the Main Interface

Visit the application dashboard:

```text
http://localhost:8090
```

You will see the Podux main console.

## Step 5: Add a Server

1. Click **Servers** in the left navigation
2. Click **Add Server** and fill in the following:

   | Field | Description | Example |
   | --- | --- | --- |
   | Name | Display name for this server | `My Server` |
   | Server Address | IP or domain of your frps host | `your-server.com` |
   | Server Port | frps listening port, default `7000` | `7000` |
   | Auth Token | The `auth.token` from your frps config | `your-token` |

3. Enable **Auto Connect** and save — Podux will connect immediately.

Once connected, the server card will show an **Online** status and current latency.

## Step 6: Add a Proxy

1. Click **Proxies**, then **Add Proxy**
2. Choose a proxy type (TCP / UDP / HTTP / HTTPS, etc.) and fill in the config:

   **Example: Expose local SSH (port 22) to the internet**

   | Field | Value |
   | --- | --- |
   | Name | `my-ssh` |
   | Type | `TCP` |
   | Server | Select the server you added |
   | Local IP | `127.0.0.1` |
   | Local Port | `22` |
   | Remote Port | `6022` |

3. Save — Podux will automatically hot-reload the frpc config, no restart needed.

You can then connect via `ssh -p 6022 user@your-server.com`.
