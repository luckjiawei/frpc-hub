---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "Podux"
  text: "现代化的 Frpc 管理平台"
  tagline: 一站式内网穿透服务管理平台，让 frpc 配置和监控变得简单高效
  image:
    src: /logo.svg
    alt: Podux Dashboard
  actions:
    - theme: brand
      text: 什么是 Podux?
      link: /guide/what-is-podux
    - theme: alt
      text: 快速开始
      link: /guide/getting-started
    - theme: alt
      text: GitHub
      link: https://github.com/luckjiawei/podux

features:
  - icon: 🖥️
    title: 服务器管理
    details: 轻松创建、配置和管理多个 frpc 服务器，支持自动连接和状态监控
  - icon: 🔌
    title: 代理配置
    details: 支持 HTTP、HTTPS、TCP、UDP 等多种代理类型，配置简单直观
  - icon: 📊
    title: 实时监控
    details: 实时网络延迟监控、地理位置展示，一目了然
  - icon: 🔄
    title: 热重载
    details: 配置更改后自动重载 frpc，无需手动重启服务
  - icon: 🌐
    title: 多语言支持
    details: 内置国际化支持，轻松切换中英文界面
  - icon: 🎨
    title: 现代化界面
    details: 基于 Radix UI 和 Tailwind CSS，提供优雅的用户体验
  - icon: 🚀
    title: 高性能
    details: 基于 Go 和 React，提供快速的开发和运行体验
  - icon: 📦
    title: 开箱即用
    details: 跨平台支持，单一可执行文件，无需复杂配置即可快速部署到 Linux、Windows、macOS
---
