# bururu.dev

The source of https://bururu.dev, the guide to making mods for Bururu. The pages are Markdown files. A small Go program turns them into a static site in `public/`, and Cloudflare Pages serves that folder.

The English pages are copied from the modding guide that ships with the app, so fix the English text there and sync it again. Translations live only here.

## What is where

| Path | What |
|---|---|
| `content/versions.json` | the docs versions, newest first, and which one is current |
| `content/languages.json` | the languages: code, name in that language, and `"dir": "rtl"` for one written right to left |
| `content/ui/<lang>.json` | the menu and note strings of one language |
| `content/<version>/pages.json` | the order of the pages in the sidebar |
| `content/<version>/en/*.md` | the English pages, as `tools/sync` copied them |
| `content/<version>/<lang>/*.md` | translations, with the same file names |
| `static/` | copied into the site as it is: `site.css`, `site.js`, the logo, the icons, `_headers` |
| `tools/build/` | makes the site; `page.html` is the page template |
| `tools/sync/` | copies a new version of the guide from the app |
| `tools/serve/` | shows the built site on your PC the way Cloudflare Pages does |

## Build and look at it

You need Go 1.22 or newer. Nothing else: the tools are plain Go and build with `CGO_ENABLED=0`.

```
go run ./tools/build
go run ./tools/serve
```

Then open http://127.0.0.1:8788/. The build empties `public/` and writes it again. It stops with a list when a link goes to a page or a heading that does not exist, in any language.

`go test ./...` runs the tests of both tools.

The site follows the system's light or dark mode. To see the other one, switch Windows' app mode, or in the browser's DevTools open **Rendering** and set **Emulate CSS media feature prefers-color-scheme**.

## Versions

A docs version is the app's major and minor number. Bururu 0.14.0, 0.14.13 and a test build such as 0.14.13-test all read the 0.14 docs, because patch releases do not change how mods are made. Each version also names the mod API it speaks (the `api` field of a mod's manifest), which the version menu shows.

Every page lives at `/<version>/<language>/<page>/`, such as `/0.14/fr/rules/`. The address keeps the version and the language, so a link someone shares opens the same page in the same language. The start page is `/0.14/en/`. Addresses without a version or a language go to the current version: `/` and `/latest/` go to `/0.14/en/`, `/fr/` to `/0.14/fr/`, `/rules/` to `/0.14/en/rules/`. The build writes these redirects into `public/_redirects` from `versions.json`, so do not edit that file by hand.

An older version gets a note at the top of each page that links to the same page in the current version.

## Sync a new version of the guide

When the app's modding guide changes, copy it again. `-from` is the folder of the app's repository, here next to this one:

```
go run ./tools/sync -from ../bururu -version 0.15
```

`-version` takes `0.15`, `0.15.2` or `0.15.2-test` alike. The sync:

- copies `docs/modding/*.md` into `content/0.15/en/` (`README.md` becomes `index.md`) and the pages of `docs/` the guide links to into `content/0.15/en/player/`,
- rewrites the links for this site; a link to a page it does not copy turns into plain text, and a "Back to the README" line that points outside the guide is left out,
- writes the sidebar order into `content/0.15/pages.json` (the start page, then the guide's "Chapters" list, then the rest),
- adds 0.15 to `versions.json` with its mod API, read from `sharing.md`, and makes it current when it is the newest,
- stops when a page holds a path of someone's PC, such as `C:\Users\...`.

Syncing the same version again replaces its English pages, so a page the guide dropped goes too. Older versions stay as they are.

Then build, look at the pages it lists as coming from the player guide, read the diff (this repository is public), and commit.

Translations do not move to a new version by themselves. To keep one, copy `content/0.14/<lang>/` to `content/0.15/<lang>/` and bring each page up to date with the new English one.

## Add a translation

A page with no translation shows the English page, with a note at the top in the reader's language. So a language can start with one page.

1. Make the folder `content/<version>/<lang>/`, with `<lang>` as in `languages.json` (`fr`, `pt-BR`, `zh-TW`...).
2. Copy an English page into it under the same name, such as `rules.md`, and translate it. Translate the text and leave code blocks, file names, commands and field names as they are.
3. Keep the links as the English page has them (`sharing.md#the-manifest`). They go to the page in the same language, or to the English page when there is no translation.
4. Keep the headings: the same number, at the same levels, in the same order. Then each translated heading gets the English heading's id, and links into the page keep working. When a translation needs other headings, put the English id after each heading that other pages link to, as in `## Le manifeste {#the-manifest}`. The build tells you which ones are missing.
5. Build and look at the page.

The menus and notes come from `content/ui/<lang>.json`. A string it lacks shows in English.

To add a language, add it to `content/languages.json` with its code and its name in that language, add `"dir": "rtl"` if it is written right to left, and add `content/ui/<lang>.json`.

Pages in a right-to-left language get `dir="rtl"`: the sidebar moves to the right and the page mirrors. Code stays left to right. `site.css` uses only the inline-start and inline-end forms of margins, paddings, borders and positions for that reason, and a test checks it.

## Deploy on Cloudflare Pages

1. In the Cloudflare dashboard, create a Pages project and connect it to the `getbururu/bururu.dev` repository on GitHub, production branch `main`.
2. Build settings: framework preset **None**, build command `go run ./tools/build`, build output directory `public`.
3. Go is in Cloudflare's build image (1.24 in the v3 image), and this repository needs 1.22 or newer. To pin a version, set the environment variable `GO_VERSION`, for example to `1.24.3`.
4. Deploy. Every push to `main` builds and updates the site. Other branches get a preview address on `pages.dev`, which `_headers` keeps out of search engines.
5. `bururu.dev` is a bare domain, and Pages takes a bare domain only when it is a zone in the same Cloudflare account: add `bururu.dev` as a site in Cloudflare and set Cloudflare's name servers at the registrar. Then, in the project's **Custom domains**, add `bururu.dev`. Pages adds the DNS record and the certificate itself. A `.dev` domain works over HTTPS only, which Pages already does.

After a deploy, check the headers and a redirect:

```
curl -sI https://bururu.dev/0.14/en/
curl -sI https://bururu.dev/
```

The first should list the `Content-Security-Policy` and the other headers from `static/_headers`; the second should answer 302 with `Location: /0.14/en/`.

## Scripts and the headers

The pages work without JavaScript. `static/assets/site.js` does two things: it closes the version, language and page menus when you click elsewhere or press Escape, and it runs the search. The search reads `search.json` of the version and language (a language with no translations reads the English one) and looks through it in the browser, so it needs no server.

`static/_headers` sets a strict Content-Security-Policy: scripts, styles and pictures only from the site itself, `connect-src 'self'` for the search index, and no inline scripts or styles anywhere. If you add a script, a font or a file from another host, add it to the policy, or browsers will block it. `tools/serve` applies the same file, so a blocked file shows in DevTools before you deploy.

Markdown in a page cannot add HTML: the build leaves raw HTML out, so a translation cannot bring in a script.

## Look

The colours, type and shapes are the app window's: the same tokens as Bururu's own window, with Sun Clay `#E37C45` for the brand and the current page only, Night Ink `#16151B`, and Segoe UI Variable with the system's fonts as fallback. The logo files come from the brand kit.
