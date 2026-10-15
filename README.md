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
desktop and a slide-out drawer with focus trapping, Escape/backdrop dismissal,
and scroll locking on smaller screens. Chapter links load ordinary
static pages, with the current page highlighted—no SPA or JavaScript routing.
Navigation remains available without JavaScript.

## Browser features

- **Settings:** System / Light / Dark appearance, optional interface sounds (off
  by default), and a reading-tracking toggle. Native form controls and diagrams
  follow the selected theme, including animated Alpine states.
- **Progress:** the learning map and sidebar distinguish unread, in-progress, and
  read chapters. Section pips show which sections are read. A section is marked
  read after at least three seconds in view and reaching its end; each section
  also has an explicit read/unread button. Marking a section unread keeps it
  unread for the rest of that page visit. Pausing tracking retains existing data.
- **Storage:** IndexedDB database `embedded-visual-guide`, store `state`, holds
  `preferences` and `progress/<chapter-slug>` records. Data stays on this browser
  origin, not a server; clearing site storage removes it. Tabs synchronize via
  BroadcastChannel. A localStorage theme cache is used only before first paint.
  If storage is blocked, controls still work in memory and Settings shows a warning.
- **Reset:** Settings → Reset reading progress, or **Ctrl+Shift+R** (also
  **⌘+Shift+R**), clears reading records while keeping theme, tracking, and sound
  preferences. The shortcut reloads after clearing. Browser-reserved shortcuts
  may not reach the page, so the reset button is the reliable alternative.
- **Sound:** synthesized Web Audio tones for controls, throttled scroll/slider
  feedback, and correct/incorrect answers. Audio requires opt-in and a gesture.
  No audio files, network requests, or autoplay audio. Components can dispatch
  `guide:feedback` with `{ type: "correct" }` or `{ type: "incorrect" }` as detail.
- **Simulations:** UART, clock signals, PWM, boot, and interrupts have Play/Pause,
  Step, Reset, and playback-speed controls. They start paused, retain their
  position when paused, and pause when their tab is hidden. Timings are slowed
  educational models, not hardware emulation. Ambient motion respects the OS
  reduced-motion setting; simulations only animate after an explicit action.

### Optional browser regression tests

Go remains the only build requirement. To run the additional Playwright checks:

```bash
npm install --no-save --package-lock=false playwright
npx playwright install --no-shell chromium
make run                 # leave running in another terminal
SITE_URL=http://localhost:8080 node scripts/browser-test.cjs
```

The checks exercise mobile drawer/focus, themes, persisted preferences and
progress, reset controls/shortcut, storage-denied and no-JS fallbacks, simulation
pause/step/reset, and every chapter at phone, tablet, and desktop widths. Set
`SCREENSHOT_DIR` to an existing directory to save visual snapshots.

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
