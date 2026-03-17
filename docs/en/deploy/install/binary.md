# Binary Installation

::: warning Work in Progress
This page is still being improved. Some details may not be fully accurate — please refer to the actual behavior when in doubt.
:::

Install Podux using pre-compiled binaries — ideal for running directly on a server or local machine.

## Download

Go to the [GitHub Releases](https://github.com/luckjiawei/podux/releases/latest) page and download the package for your platform:

| OS | Architecture | Filename |
| --- | --- | --- |
| Linux | x86_64 (amd64) | `podux-linux-amd64.tar.gz` |
| Linux | ARM64 | `podux-linux-arm64.tar.gz` |
| Linux | ARMv7 | `podux-linux-armv7.tar.gz` |
| macOS | Intel | `podux-darwin-amd64.tar.gz` |
| macOS | Apple Silicon | `podux-darwin-arm64.tar.gz` |
| Windows | x86_64 | `podux-windows-amd64.zip` |

Or download directly from the command line (Linux amd64 example):

```bash
curl -LO https://github.com/luckjiawei/podux/releases/latest/download/podux-linux-amd64.tar.gz
```

## Install

### Linux / macOS

```bash
# Extract
tar -xzf podux-linux-amd64.tar.gz

# Move to system path (optional)
sudo mv podux /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/podux

# Verify installation
podux --version
```

### Windows

Extract the `.zip` file to any directory, then open PowerShell or Command Prompt inside that directory to use it.

## Start

```bash
podux serve --http 0.0.0.0:8090
```

After starting, visit `http://localhost:8090` for the main interface and `http://localhost:8090/_/` for the admin panel.

::: tip
On first launch, the database is initialized automatically. Data is stored in a `pb_data/` folder in the current working directory. Fix the working directory to avoid data scatter.
:::

## Auto-start on Boot (Linux)

Use systemd to manage the Podux process.

**1. Create a dedicated user (optional but recommended)**

```bash
sudo useradd -r -s /bin/false podux
```

**2. Create the data directory**

```bash
sudo mkdir -p /var/lib/podux
sudo chown podux:podux /var/lib/podux
```

**3. Create the systemd service file**

```bash
sudo tee /etc/systemd/system/podux.service > /dev/null <<EOF
[Unit]
Description=Podux Service
After=network.target

[Service]
Type=simple
User=podux
WorkingDirectory=/var/lib/podux
ExecStart=/usr/local/bin/podux serve --http 0.0.0.0:8090
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
```

**4. Enable and start the service**

```bash
sudo systemctl daemon-reload
sudo systemctl enable podux
sudo systemctl start podux

# Check status
sudo systemctl status podux
```

## View Logs

```bash
sudo journalctl -u podux -f
```

## Next Steps

- [Binary Upgrade](/en/deploy/upgrade/binary) — Upgrade to a new version
- [Getting Started](/en/guide/getting-started) — Add servers and proxies
