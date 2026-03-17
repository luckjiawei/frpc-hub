import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  // base: '/podux-doc/',
  title: "Podux",
  description: "现代化的 Frpc 管理平台 - 基于 PocketBase 和 React 构建的强大内网穿透管理系统",
  head: [
    ['link', { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }],
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
    ['script', { defer: '', src: 'https://umami.jwinks.com/script.js', 'data-website-id': '8af0f36e-650c-45b5-a17e-e104b918f63f' }]
  ],
  metaChunk: true,

  locales: {
    root: {
      label: '简体中文',
      lang: 'zh-CN',
      themeConfig: {
        nav: [
          { text: '首页', link: '/' },
          { text: '指南', link: '/guide/what-is-podux' },
        ],
        sidebar: [
          {
            text: '指南',
            items: [
              { text: '什么是 Podux?', link: '/guide/what-is-podux' },
                        { text: '界面预览', link: '/guide/screenshots' },

              { text: '快速开始', link: '/guide/getting-started' },
            ]
          },
          {
            text: '安装部署',
            items: [
              {
                text: '安装',
                items: [
                  { text: '二进制', link: '/deploy/install/binary' },
                  { text: 'Docker', link: '/deploy/install/docker' },
                ]
              },
              {
                text: '升级',
                items: [
                  { text: '二进制', link: '/deploy/upgrade/binary' },
                  { text: 'Docker', link: '/deploy/upgrade/docker' },
                ]
              }
            ]
          },
          {
            text: '故障排除',
            items: [
              { text: '常见问题', link: '/troubleshoot/faq' },
            ]
          }
        ],
        footer: {
          message: '基于 MIT 协议开源',
          copyright: 'Copyright © 2026-present Podux'
        },
      }
    },
    en: {
      label: 'English',
      lang: 'en-US',
      description: "Modern Frpc Management Platform",
      themeConfig: {
        nav: [
          { text: 'Home', link: '/en/' },
          { text: 'Guide', link: '/en/guide/what-is-podux' }
        ],
        sidebar: [
          {
            text: 'Guide',
            items: [
              { text: 'What is Podux?', link: '/en/guide/what-is-podux' },
              { text: 'Screenshots', link: '/en/guide/screenshots' },
              { text: 'Getting Started', link: '/en/guide/getting-started' }
            ]
          },
          {
            text: 'Deployment',
            items: [
              {
                text: 'Install',
                items: [
                  { text: 'Binary', link: '/en/deploy/install/binary' },
                  { text: 'Docker', link: '/en/deploy/install/docker' }
                ]
              },
              {
                text: 'Upgrade',
                items: [
                  { text: 'Binary', link: '/en/deploy/upgrade/binary' },
                  { text: 'Docker', link: '/en/deploy/upgrade/docker' }
                ]
              }
            ]
          },
          {
            text: 'Troubleshooting',
            items: [
              { text: 'FAQ', link: '/en/troubleshoot/faq' }
            ]
          }
        ],
        footer: {
          message: 'Released under the MIT License.',
          copyright: 'Copyright © 2026-present Podux'
        },
      }
    }
  },

  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: { src: '/logo.svg', width: 24, height: 24 },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/luckjiawei/podux' }
    ],

    search: {
      provider: 'local'
    }
  }
})
