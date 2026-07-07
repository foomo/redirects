import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

export default withMermaid(defineConfig({
  title: 'redirects',
  description: 'URL redirect management for Go — automatic and manual redirects with cycle detection and flattening',
  base: '/redirects/',
  cleanUrls: true,
  markdown: {
    lineNumbers: true,
  },
  ignoreDeadLinks: [
    /^http:\/\/localhost/,
  ],
  themeConfig: {
    logo: '/logo.png',
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Reference', link: '/reference/configuration' },
      { text: 'godoc', link: 'https://pkg.go.dev/github.com/foomo/redirects/v2' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting started', link: '/guide/getting-started' },
            { text: 'Architecture', link: '/guide/architecture' },
            { text: 'Integration', link: '/guide/integration' },
          ],
        },
      ],
      '/reference/': [
        {
          text: 'Reference',
          items: [
            { text: 'Configuration', link: '/reference/configuration' },
            { text: 'API', link: '/reference/api' },
          ],
        },
      ],
      '/': [
        {
          text: 'Contributing',
          items: [
            { text: 'Contributing', link: '/CONTRIBUTING' },
            { text: 'Code of Conduct', link: '/CODE_OF_CONDUCT' },
            { text: 'Security', link: '/SECURITY' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/foomo/redirects' },
    ],
    editLink: {
      pattern: 'https://github.com/foomo/redirects/edit/main/docs/:path',
      text: 'Edit this page on GitHub',
    },
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Made with ♥ by foomo / bestbytes',
    },
  },
  mermaid: {},
  mermaidPlugin: {
    class: 'mermaid',
  },
}))
