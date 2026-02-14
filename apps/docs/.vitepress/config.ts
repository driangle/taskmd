import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'taskmd',
  description: 'Manage tasks with markdown files',
  base: '/taskmd/',

  head: [
    ['meta', { name: 'theme-color', content: '#5b7ee5' }],
  ],

  themeConfig: {
    nav: [
      { text: 'Guide', link: '/getting-started/' },
      { text: 'CLI Reference', link: '/guide/cli' },
      { text: 'Specification', link: '/reference/specification' },
      { text: 'FAQ', link: '/faq' },
    ],

    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Quick Start', link: '/getting-started/' },
          { text: 'Installation', link: '/getting-started/installation' },
          { text: 'Core Concepts', link: '/getting-started/concepts' },
        ],
      },
      {
        text: 'User Guide',
        items: [
          { text: 'CLI Guide', link: '/guide/cli' },
          { text: 'Web Interface', link: '/guide/web' },
        ],
      },
      {
        text: 'Reference',
        items: [
          { text: 'Task Specification', link: '/reference/specification' },
          { text: 'Configuration', link: '/reference/configuration' },
        ],
      },
      {
        text: 'More',
        items: [
          { text: 'FAQ', link: '/faq' },
          { text: 'Contributing', link: '/contributing/' },
          { text: 'Releasing', link: '/contributing/releasing' },
        ],
      },
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/driangle/taskmd' },
    ],

    editLink: {
      pattern: 'https://github.com/driangle/taskmd/edit/main/apps/docs/:path',
      text: 'Edit this page on GitHub',
    },

    search: {
      provider: 'local',
    },

    footer: {
      message: 'Released under the MIT License.',
    },
  },
})
