import {defineConfig} from 'vitepress'
import {withMermaid} from 'vitepress-plugin-mermaid'

// https://vitepress.dev/reference/site-config
export default withMermaid(defineConfig({
	title: 'redirects',
	description: 'URL redirect management for Go — automatic and manual redirects with cycle detection and flattening',
	lang: "en-US",
	lastUpdated: true,
	appearance: "dark",
	ignoreDeadLinks: [
		/^http:\/\/localhost/,
	],
	base: '/redirects/',
	cleanUrls: true,
	sitemap: {
		hostname: 'https://foomo.github.io/redirects',
	},
	themeConfig: {
		// https://vitepress.dev/reference/default-theme-config
		logo: '/logo.png',
		outline: [2, 4],
		nav: [
			{text: 'Guide', link: '/guide/getting-started'},
			{text: 'Reference', link: '/reference/configuration'},
			{text: 'godoc', link: 'https://pkg.go.dev/github.com/foomo/redirects/v2'},
		],
		sidebar: [
			{
				text: 'Guide',
				items: [
					{text: 'Getting started', link: '/guide/getting-started'},
					{text: 'Architecture', link: '/guide/architecture'},
					{text: 'Integration', link: '/guide/integration'},
				],
			},
			{
				text: 'Reference',
				items: [
					{text: 'Configuration', link: '/reference/configuration'},
					{text: 'API', link: '/reference/api'},
				],
			},
			{
				text: 'Contributing',
				items: [
					{text: 'Guideline', link: '/CONTRIBUTING'},
					{text: 'Code of conduct', link: '/CODE_OF_CONDUCT'},
					{text: 'Security guidelines', link: '/SECURITY'},
				],
			},
		],
		socialLinks: [
			{icon: 'github', link: 'https://github.com/foomo/redirects'},
		],
		editLink: {
			pattern: 'https://github.com/foomo/redirects/edit/main/docs/:path',
			text: 'Edit this page on GitHub',
		},
		search: {
			provider: 'local',
		},
		footer: {
			message: 'Made with ♥ <a href="https://www.foomo.org">foomo</a> by <a href="https://www.bestbytes.com">bestbytes</a>',
		},
	},
	markdown: {
		// https://github.com/vuejs/vitepress/discussions/3724
		theme: {
			light: 'catppuccin-latte',
			dark: 'catppuccin-frappe',
		}
	},
	mermaid: {},
	mermaidPlugin: {
		class: 'mermaid',
	},
	head: [
		['meta', {name: 'theme-color', content: '#ffffff'}],
		['link', {rel: 'icon', href: '/logo.png'}],
		['meta', {name: 'author', content: 'foomo by bestbytes'}],
		// OpenGraph
		['meta', {property: 'og:title', content: 'foomo/redirects'}],
		[
			'meta',
			{
				property: 'og:image',
				content: 'https://github.com/foomo/redirects/blob/main/docs/public/banner.png?raw=true',
			},
		],
		[
			'meta',
			{
				property: 'og:description',
				content: 'Generic, transport-agnostic messaging patterns for Go',
			},
		],
		['meta', {name: 'twitter:card', content: 'summary_large_image'}],
		[
			'meta',
			{
				name: 'twitter:image',
				content: 'https://github.com/foomo/redirects/blob/main/docs/public/banner.png?raw=true',
			},
		],
		[
			'meta', {
			name: 'viewport', content: 'width=device-width, initial-scale=1.0, viewport-fit=cover',
		},
		],
	]
}))
