import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  base: process.env.NODE_ENV === 'production' ? '/qq-bot/' : '/',
  title: "QQ Bot",
  description: "qq 机器人，chatgpt/k8s部署/软件自更新/色图/天气等。",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Home', link: '/' },
      { text: '文档', link: '/quick-start' }
    ],

    sidebar: [
      {
        text: '使用教程',
        items: [
          { text: '快速开始', link: '/quick-start' },
          { text: '命令大全', link: '/command' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/Lick-Dog-Club/qq-bot' }
    ]
  }
})
