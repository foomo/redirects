# Documentation

This directory holds the [VitePress](https://vitepress.dev/) documentation site
that is published to <https://foomo.github.io/redirects/>.

## Local development

`bun` is provided by `mise` (see `.mise.toml`), so `mise install` from the repo
root gets you the right version. Then, from the repo root:

```sh
make docs        # bun install && bun run dev  → live-reload server
make docs.build  # bun install && bun run build → production build in docs/.vitepress/dist
```

Or directly inside this directory:

```sh
bun install
bun run dev
```

## How it is published

The [`Deploy Docs`](https://github.com/foomo/redirects/blob/main/.github/workflows/docs.yml)
GitHub Actions workflow builds the site and deploys it to GitHub Pages. It runs on every `v*.*.*` tag
push and can also be triggered manually from the **Actions** tab
(`workflow_dispatch`). The `build` job compiles the site with `bun run build`
and uploads `docs/.vitepress/dist`; the `deploy` job publishes it to the
`github-pages` environment.

## One-time GitHub repository settings

The workflow only works once GitHub Pages is switched to the "GitHub Actions"
source. Do this once per repository (requires admin rights):

1. Open the repository on GitHub → **Settings** → **Pages**
   (`https://github.com/foomo/redirects/settings/pages`).
2. Under **Build and deployment → Source**, choose **GitHub Actions**
   (not "Deploy from a branch"). No branch selection is needed.
3. *(Optional, recommended)* Under **Settings → Environments → github-pages**,
   confirm the environment exists and, if you use deployment protection rules,
   add the tags/branches you deploy from (e.g. allow tags matching `v*.*.*`).
   By default no protection rules are required.
4. Trigger a first deployment — either push a `v*.*.*` tag, or open
   **Actions → Deploy Docs → Run workflow** to run it manually via
   `workflow_dispatch`.
5. Once the `deploy` job succeeds, the site is live at
   <https://foomo.github.io/redirects/>. The URL also appears on the
   **Settings → Pages** screen and on the workflow's `deploy` job summary.

No secrets need to be configured: the workflow authenticates to Pages through
the built-in `GITHUB_TOKEN` using the `pages: write` and `id-token: write`
permissions already declared in `docs.yml`.

> The site is served from the `/redirects/` base path (configured as `base` in
> `.vitepress/config.mts`), matching the GitHub Pages project URL. If the
> repository is ever renamed, update `base` accordingly.
