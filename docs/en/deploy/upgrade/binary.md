# Binary Upgrade

::: warning Work in Progress
This page is still being improved. Some details may not be fully accurate — please refer to the actual behavior when in doubt.
:::

Before upgrading the Podux binary, back up your data first, then replace the executable.

## Upgrade Steps

### 1. Back Up Data

Data is stored in `pb_data/` under the working directory. Back it up before upgrading:

```bash
cp -r /var/lib/podux/pb_data /var/lib/podux/pb_data.bak
```

### 2. Stop the Service

```bash
sudo systemctl stop podux
```

If not using systemd, kill the process manually:

```bash
pkill podux
```

### 3. Download the New Version

Visit [GitHub Releases](https://github.com/luckjiawei/podux/releases/latest) for the latest version, or download via command line (Linux amd64 example):

```bash
curl -LO https://github.com/luckjiawei/podux/releases/latest/download/podux-linux-amd64.tar.gz
tar -xzf podux-linux-amd64.tar.gz
```

### 4. Replace the Binary

```bash
sudo mv podux /usr/local/bin/podux
sudo chmod +x /usr/local/bin/podux
```

### 5. Restart the Service

```bash
sudo systemctl start podux

# Confirm service is running
sudo systemctl status podux
```

## Rollback

If issues occur after upgrading, restore the backup and reinstall the old version:

```bash
# Stop service
sudo systemctl stop podux

# Restore data backup
rm -rf /var/lib/podux/pb_data
cp -r /var/lib/podux/pb_data.bak /var/lib/podux/pb_data

# Download the old version from GitHub Releases, replace the binary, then start
sudo systemctl start podux
```
