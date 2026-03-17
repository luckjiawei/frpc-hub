# 二进制升级

::: warning 文档待完善
本页内容正在持续完善中，部分细节可能不够准确，请以实际情况为准。
:::

升级 Podux 二进制版本前，建议先备份数据，再替换可执行文件。

## 升级步骤

### 1. 备份数据

数据默认存储在工作目录下的 `pb_data/`，升级前先备份：

```bash
cp -r /var/lib/podux/pb_data /var/lib/podux/pb_data.bak
```

### 2. 停止服务

```bash
sudo systemctl stop podux
```

若未使用 systemd，手动终止进程：

```bash
pkill podux
```

### 3. 下载新版本

前往 [GitHub Releases](https://github.com/luckjiawei/podux/releases/latest) 查看最新版本，或通过命令行下载（以 Linux amd64 为例）：

```bash
curl -LO https://github.com/luckjiawei/podux/releases/latest/download/podux-linux-amd64.tar.gz
tar -xzf podux-linux-amd64.tar.gz
```

### 4. 替换二进制文件

```bash
sudo mv podux /usr/local/bin/podux
sudo chmod +x /usr/local/bin/podux
```

### 5. 重启服务

```bash
sudo systemctl start podux

# 确认服务正常运行
sudo systemctl status podux
```


## 回滚

如果升级后出现问题，恢复备份数据并重新安装旧版本：

```bash
# 停止服务
sudo systemctl stop podux

# 恢复数据备份
rm -rf /var/lib/podux/pb_data
cp -r /var/lib/podux/pb_data.bak /var/lib/podux/pb_data

# 从 GitHub Releases 下载对应旧版本并替换二进制文件，然后启动
sudo systemctl start podux
```
