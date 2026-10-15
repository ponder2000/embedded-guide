# Embedded Systems · Visual Guide

A chapter-wise **static website** for learning embedded systems — hand-drawn
SVG diagrams, animated components, and a step-by-step path from *bits* to a
*blinking chip*. Go's `html/template` generates the pages; vendored Alpine.js
keeps the interactive exercises working entirely in the browser.

**No server runtime, JavaScript build tooling, or Go dependencies are required
for deployment.** Go 1.26+ is needed only to build or preview the site.

## Quick start

```bash
make              # generate dist/ (same as make build)
make run          # build, then preview at http://localhost:8080/
make preview      # serve the existing dist/ without rebuilding
make dev          # watch/rebuild with optional Air, otherwise same as make run
```

`make tools` installs the optional Air watcher. Reload your browser after edits;
Air rebuilds HTML/assets and restarts the preview server, not the browser.
Override `PORT=8081` if needed. The preview server never renders templates.

The homepage introduces the guide's purpose, audience, and learning path.
Overview and chapter pages share chapter navigation: a sticky left sidebar on
desktop and a collapsible menu on smaller screens. Chapter links load ordinary
static pages, with the current page highlighted—no SPA or JavaScript routing.
Navigation remains available without JavaScript.

## Build and deploy

```bash
go run ./cmd/build                              # root-hosted site -> dist/
go run ./cmd/build -out public -base-path /guide/ # project-hosted site -> public/
make build BASE_PATH=/visual-guide/             # GitHub Pages project hosting
make run BASE_PATH=/visual-guide/               # preview at /visual-guide/
make -f Makefile.build build                    # equivalent static build entrypoint
```

Deploy the **contents** of the output directory to any static host. The output
contains `index.html`, `chapter/<slug>/index.html` for every registered chapter,
`static/`, `404.html`, and `.nojekyll` for GitHub Pages. Chapter links always end
in `/`; use a host that serves directory indexes and redirects directory URLs
to their trailing-slash form. There is no SPA fallback: unknown paths must
return `404.html` with a 404 status (GitHub Pages does this automatically).

For GitHub Pages, upload `dist/` using the Pages artifact workflow and deploy
that artifact, or publish its contents to your Pages branch. Use
`BASE_PATH=/repository-name/` for a project site and an empty base path for a
user/organization site or custom domain served at `/`. The prefix changes URLs,
not the output directory structure. Rebuild when the deployment prefix changes.
The same prefix must be supplied to the local preview command.

The Go generator uses embedded templates/assets but does not produce a deployed
executable. Typography uses IBM Plex Sans and IBM Plex Mono, loaded from Google
Fonts with local system fallbacks. The Alpine runtime and application scripts are
copied locally.

### Output safety

Use a dedicated output directory. Builds accept an absent/empty directory or a
previous build with its `.visual-guide-manifest.json` checksum manifest intact.
They refuse unknown files, edited generated files, and symlinks rather than
overwriting user data. Only stale files recorded in the previous manifest are
removed, individually; directories are never recursively removed. Keep the
manifest locally for rebuilds; it need not be deployed. Do not edit generated
HTML: edit the templates instead. If a build is interrupted or output was
modified, preserve/inspect it and choose a fresh `OUT` directory to rebuild.
Builds are not transactional; do not run two generators against the same output.
There is deliberately no broad `clean` target touching `bin/` or user files.

## Project layout

```
cmd/build/         static generation CLI
cmd/server/        optional stdlib static preview CLI
assets.go          embeds the first-party templates and static assets
internal/
  content/         chapter registry (order, parts, prev/next, lookup)
  render/          HTTP-independent html/template renderer and partial helper
  site/            page generation, base-path validation, safe output updates
  preview/         static indexes, redirects, and branded HTTP 404 handling
web/
  templates/       layout, home, chapter, and error templates
  templates/chapters/   one HTML partial per chapter
  static/          CSS, application JS, vendored Alpine.js
Makefile           static build, preview, development, and checks
Makefile.build     alternate entrypoint for the same static build targets
```

## Adding a chapter

1. Create `web/templates/chapters/NN-your-slug.html` with a single
   `{{ define "chapter-your-slug" }} … {{ end }}` block.
2. Add an entry to `Chapters` in `internal/content/chapters.go`. Its order defines
   numbering, the table of contents, and previous/next links automatically.
3. Run `make check build` and deploy the rebuilt output.

Interactive components use inline Alpine `x-data`; see
`web/templates/chapters/03-bitwise-and-registers.html`. Use the `url` template
helper for root-relative internal links so project hosting keeps working.

## Development checks

```bash
make fmt          # gofmt
make vet          # go vet ./...
make test         # go test ./...
make check        # format + vet + tests
go build ./...    # compile generator and optional preview
make help
```

## Chapters

**Foundations** — what an embedded system is · bits/bytes/hex/two's complement ·
bitwise ops & registers (interactive).
**The Machine** — memory map & memory-mapped I/O · boot & startup (stepper) ·
clocks & prescalers (calculator).
**Talking to the World** — GPIO (toggle) · interrupts & NVIC (stepper) ·
timers & PWM (dimmer) · UART (frame builder) · SPI & I²C.
**Going Deeper** — ADC (quantisation explorer) · low power, `volatile` & the C runtime.
