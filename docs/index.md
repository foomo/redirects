---
layout: home

hero:
  name: redirects
  text: URL redirect management for Go
  tagline: Authoring is separated from serving — a domain service creates, validates and stores redirects; a gateway matches and redirects; a NATS signal keeps them in sync.
  image:
    src: /logo.png
    alt: redirects
  actions:
    - theme: brand
      text: Get started
      link: /guide/getting-started
    - theme: alt
      text: Integration guide
      link: /guide/integration
    - theme: alt
      text: View on GitHub
      link: https://github.com/foomo/redirects

features:
  - title: Automatic & manual redirects
    details: Automatic redirects are generated from content-server updates with cycle detection; manual ones are authored through a frontend widget with validation.
  - title: Cycle detection & flattening
    details: Cyclic chains are detected and disabled; multi-step chains are flattened to a single hop so visitors and crawlers are redirected exactly once.
  - title: Decoupled by a NATS signal
    details: The service persists to MongoDB and publishes a change signal; the gateway holds definitions in memory and reloads on the signal. No shared database on the hot path.
  - title: Safe on the request path
    details: The gateway middleware only acts on GET, swallows errors, and falls through when nothing matches — redirects can never break the gateway.
---
