# 快速开始

本指南帮助你在 5 分钟内通过 Docker 完成 Podux 的安装和初始配置。

## 前置条件

- 已安装 [Docker](https://docs.docker.com/get-docker/)
- 一台已部署 frps 的公网服务器（或使用第三方 frp 服务）
- 运行 Podux 的机器能够访问公网

::: info 还没有公网服务器？
推荐使用 **[雨云](https://www.rainyun.com/s=ljw_?podux)** —— 价格实惠、稳定可靠的国内云服务器，适合部署 frps 和各类自托管服务。
:::

## 第一步：启动容器

```bash
docker run -d \
  --name podux \
  --restart unless-stopped \
  -p 8090:8090 \
  -v podux-data:/app/pb_data \
  ghcr.io/luckjiawei/podux:latest
```

启动成功后，终端会输出类似以下内容：

```text
App:   http://0.0.0.0:8090
Admin: http://0.0.0.0:8090/_/
```

::: tip 使用 Docker Compose
也可以通过 Docker Compose 管理：

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

## 第二步：创建账号

打开浏览器访问主界面：

```text
http://localhost:8090
```

首次访问时，系统会引导你创建账号，填写用户名和密码后提交即可。

## 第三步：创建后台账号（可选）

如需使用后台管理功能，可访问 Admin UI：

```text
http://localhost:8090/_/
```

首次访问时，系统会引导你创建后台账号，填写邮箱和密码后提交即可。

::: warning 安全提示
如果 Podux 运行在公网可访问的服务器上，请务必设置强密码，并考虑通过防火墙限制 `8090` 端口的访问来源。
:::

## 第四步：进入主界面

访问应用主界面：

```text
http://localhost:8090
```

你将看到 Podux 的主控制台。

## 第五步：添加服务器

1. 点击左侧导航栏的 **服务器** 菜单
2. 点击 **添加服务器**，填写以下信息：

   | 字段 | 说明 | 示例 |
   | --- | --- | --- |
   | 名称 | 服务器的显示名称 | `我的服务器` |
   | 服务器地址 | frps 所在主机的 IP 或域名 | `your-server.com` |
   | 服务器端口 | frps 监听端口，默认 `7000` | `7000` |
   | 认证令牌 | frps 配置的 `auth.token` | `your-token` |

3. 开启 **自动连接** 后保存，Podux 会立即尝试连接。

连接成功后，服务器卡片将显示 **在线** 状态和当前网络延迟。

## 第六步：添加代理

1. 点击 **代理** 菜单，然后点击 **添加代理**
2. 选择代理类型（TCP / UDP / HTTP / HTTPS 等）并填写配置：

   **示例：将本机 SSH (22 端口) 映射到公网**

   | 字段 | 值 |
   | --- | --- |
   | 名称 | `my-ssh` |
   | 类型 | `TCP` |
   | 服务器 | 选择已添加的服务器 |
   | 本地 IP | `127.0.0.1` |
   | 本地端口 | `22` |
   | 远程端口 | `6022` |

3. 保存后，Podux 会自动热重载 frpc 配置，无需重启。

之后即可通过 `ssh -p 6022 user@your-server.com` 访问本机。
