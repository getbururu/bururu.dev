package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// files is a small content folder with two versions; 0.13 has no
// player/feel page.
func files() map[string]string {
	return map[string]string{
		"versions.json": `{"current": "0.14", "versions": [{"id": "0.14", "api": "0.12"}, {"id": "0.13", "api": "0.11"}]}`,

		"0.14/pages.json": `{"pages": ["index", "rules", "player/feel"]}`,
		"0.14/index.md": "# Making mods\n\nA mod tells Bururu about one game: which program it is and what it reads while it runs.\n\n" +
			"## Start here\n\nRead [the rules](rules.md#event-rules), [Effects](player/feel.md) and [below](#start-here).\n\n" +
			"```jsonc\n{\n  // a comment\n  \"id\": \"my-mod\"\n}\n```\n\n<script>alert(1)</script>\n",
		"0.14/rules.md":       "# Rules\n\nBack to the [modding guide](index.md).\n\n## Event rules\n\nText.\n\n## Example\n\n## Example\n\n| Field | What |\n|---|---|\n| `on` | the input |\n",
		"0.14/player/feel.md": "# Effects\n\nSee [Rules](../rules.md) and [the start](../index.md#start-here).\n",

		"0.13/pages.json": `{"pages": ["index", "rules"]}`,
		"0.13/index.md":   "# Making mods\n\nSee [Rules](rules.md).\n",
		"0.13/rules.md":   "# Rules\n\n## Event rules\n",
	}
}

func writeContent(t *testing.T, fs map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, text := range fs {
		file := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// buildSite builds content into a new folder and returns the folder.
func buildSite(t *testing.T, fs map[string]string) (string, error) {
	t.Helper()
	s, err := load(writeContent(t, fs))
	if err != nil {
		return "", err
	}
	s.base = "https://bururu.dev"
	out := filepath.Join(t.TempDir(), "public")
	_, err = s.write(out, filepath.Join("..", "..", "static"))
	return out, err
}

func mustBuild(t *testing.T, fs map[string]string) string {
	t.Helper()
	out, err := buildSite(t, fs)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func read(t *testing.T, out, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func has(t *testing.T, page, text string, want bool) {
	t.Helper()
	if strings.Contains(page, text) != want {
		if want {
			t.Errorf("page lacks %q", text)
		} else {
			t.Errorf("page has %q", text)
		}
	}
}

// isPage says whether an address has a page in the built site.
func isPage(out, addr string) bool {
	name := strings.TrimPrefix(addr, "/")
	if name == "" || strings.HasSuffix(name, "/") {
		name += "index.html"
	}
	st, err := os.Stat(filepath.Join(out, filepath.FromSlash(name)))
	return err == nil && !st.IsDir()
}

type redirect struct{ from, to string }

func redirects(t *testing.T, out string) []redirect {
	t.Helper()
	var rs []redirect
	for _, line := range strings.Split(read(t, out, "_redirects"), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) != 3 || f[2] != "302" {
			t.Fatalf("a rule that is not \"from to 302\": %q", line)
		}
		rs = append(rs, redirect{f[0], f[1]})
	}
	return rs
}

// follow applies the rules to addr the way Cloudflare does (the first
// rule that matches wins, a trailing * matches the rest of the address)
// until none matches, and returns where it ends.
func follow(t *testing.T, rs []redirect, addr string) string {
	t.Helper()
	for hop := 0; hop < 5; hop++ {
		next := ""
		for _, r := range rs {
			if prefix, ok := strings.CutSuffix(r.from, "*"); ok {
				if rest, ok := strings.CutPrefix(addr, prefix); ok {
					next = strings.Replace(r.to, ":splat", rest, 1)
					break
				}
			} else if addr == r.from {
				next = r.to
				break
			}
		}
		if next == "" {
			return addr
		}
		addr = next
	}
	t.Fatalf("%s: more than 5 redirects", addr)
	return ""
}

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Finding a value with brr memory": "finding-a-value-with-brr-memory",
		"The player's overrides":          "the-players-overrides",
		"sensor.set":                      "sensorset",
		"Games on macOS":                  "games-on-macos",
		"Layers and stages (advanced)":    "layers-and-stages-advanced",
		"On Linux":                        "on-linux",
	} {
		if got := slug(in); got != want {
			t.Errorf("slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPagePaths(t *testing.T) {
	out := mustBuild(t, files())
	for _, addr := range []string{"/0.14/", "/0.14/rules/", "/0.14/player/feel/", "/0.13/", "/0.13/rules/"} {
		if !isPage(out, addr) {
			t.Errorf("no page at %s", addr)
		}
	}
	for _, name := range []string{"0.14/en", "0.13/player", "en"} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(name))); err == nil {
			t.Errorf("the site has %s", name)
		}
	}
}

func TestLinksPointAtTheSite(t *testing.T) {
	out := mustBuild(t, files())
	index := read(t, out, "0.14/index.html")
	has(t, index, `href="/0.14/rules/#event-rules"`, true)
	has(t, index, `href="/0.14/player/feel/"`, true)
	has(t, index, `href="#start-here"`, true)
	feel := read(t, out, "0.14/player/feel/index.html")
	has(t, feel, `href="/0.14/rules/"`, true)
	has(t, feel, `href="/0.14/#start-here"`, true)
	has(t, read(t, out, "0.13/index.html"), `href="/0.13/rules/"`, true)
}

func TestHeadingIDs(t *testing.T) {
	out := mustBuild(t, files())
	rules := read(t, out, "0.14/rules/index.html")
	has(t, rules, `<h2 id="event-rules">`, true)
	has(t, rules, `<h2 id="example">`, true)
	has(t, rules, `<h2 id="example-1">`, true)
}

// The pages are in English only: no language menu, no note about a
// missing translation, no links to other languages.
func TestEnglishOnly(t *testing.T) {
	out := mustBuild(t, files())
	rules := read(t, out, "0.14/rules/index.html")
	has(t, rules, `<html lang="en">`, true)
	has(t, rules, `<link rel="canonical" href="https://bururu.dev/0.14/rules/">`, true)
	has(t, rules, `<article class="doc">`, true)
	has(t, rules, `class="note"`, false)
	has(t, rules, `menu-lang`, false)
	has(t, rules, `hreflang`, false)
	has(t, rules, ` dir=`, false)
	has(t, rules, `<a class="skip" href="#main">Skip to content</a>`, true)
	has(t, rules, `<title>Rules - Bururu modding docs</title>`, true)
}

func TestVersions(t *testing.T) {
	out := mustBuild(t, files())
	feel := read(t, out, "0.14/player/feel/index.html")
	// 0.13 has no such page, so its menu entry goes to its start page.
	has(t, feel, `<a href="/0.13/">0.13<span class="item-note">mod API 0.11</span></a>`, true)
	has(t, feel, `<a href="/0.14/player/feel/" aria-current="true">0.14 (current)<span class="item-note">mod API 0.12</span></a>`, true)
	has(t, read(t, out, "0.14/rules/index.html"), `<a href="/0.13/rules/">0.13`, true)

	old := read(t, out, "0.13/rules/index.html")
	has(t, old, `<p class="note">These docs are for Bururu 0.13. The newest docs are for 0.14. <a href="/0.14/rules/">Read the 0.14 docs</a></p>`, true)
	has(t, read(t, out, "0.14/rules/index.html"), "The newest docs are for", false)

	sitemap := read(t, out, "sitemap.xml")
	has(t, sitemap, "<loc>https://bururu.dev/0.14/</loc>", true)
	has(t, sitemap, "<loc>https://bururu.dev/0.14/rules/</loc>", true)
	has(t, sitemap, "<loc>https://bururu.dev/0.14/player/feel/</loc>", true)
	has(t, sitemap, "/0.13/", false)
}

func TestRedirects(t *testing.T) {
	out := mustBuild(t, files())
	text := read(t, out, "_redirects")
	for _, line := range []string{
		"/ /0.14/ 302",
		"/latest /0.14/ 302",
		"/latest/* /0.14/:splat 302",
		"/rules /0.14/rules/ 302",
		"/rules/ /0.14/rules/ 302",
		"/player/feel/ /0.14/player/feel/ 302",
		"/0.14/en /0.14/ 302",
		"/0.14/en/* /0.14/:splat 302",
		"/0.14/pt-BR/* /0.14/:splat 302",
		"/0.14/pt-br/* /0.14/:splat 302",
		"/fr /0.14/ 302",
		"/fr/* /0.14/:splat 302",
		"/pt-br/* /0.14/:splat 302",
	} {
		has(t, text, line+"\n", true)
	}
	// Only the 0.14 docs had addresses with a language.
	has(t, text, "/0.13/en", false)

	rs := redirects(t, out)
	for _, r := range rs {
		if strings.HasSuffix(r.from, "*") {
			continue // a splat: under a language or "latest", where no page is
		}
		// Cloudflare applies a redirect before it looks for a file.
		if isPage(out, r.from) || isPage(out, r.from+"/") {
			t.Errorf("the redirect %s covers a page", r.from)
		}
		if !isPage(out, r.to) {
			t.Errorf("the redirect %s goes to %s, which has no page", r.from, r.to)
		}
	}
}

// Every address the site had before it lost its languages leads to the
// same page in English.
func TestOldAddressesLeadToThePage(t *testing.T) {
	out := mustBuild(t, files())
	rs := redirects(t, out)
	for from, want := range map[string]string{
		"/":                        "/0.14/",
		"/latest":                  "/0.14/",
		"/latest/":                 "/0.14/",
		"/latest/rules/":           "/0.14/rules/",
		"/latest/en/rules/":        "/0.14/rules/",
		"/rules":                   "/0.14/rules/",
		"/rules/":                  "/0.14/rules/",
		"/player/feel":             "/0.14/player/feel/",
		"/0.14/en":                 "/0.14/",
		"/0.14/en/":                "/0.14/",
		"/0.14/en/rules/":          "/0.14/rules/",
		"/0.14/en/player/feel/":    "/0.14/player/feel/",
		"/0.14/fr/rules/":          "/0.14/rules/",
		"/0.14/ar/":                "/0.14/",
		"/0.14/pt-BR/rules/":       "/0.14/rules/",
		"/0.14/pt-br/player/feel/": "/0.14/player/feel/",
		"/0.14/zh-TW/rules/":       "/0.14/rules/",
		"/fr":                      "/0.14/",
		"/fr/":                     "/0.14/",
		"/fr/rules/":               "/0.14/rules/",
		"/en/player/feel/":         "/0.14/player/feel/",
		"/es-419/rules/":           "/0.14/rules/",
		"/es-es/rules/":            "/0.14/rules/",
		"/0.14/":                   "/0.14/",
		"/0.14/rules/":             "/0.14/rules/",
		"/0.13/rules/":             "/0.13/rules/",
	} {
		got := follow(t, rs, from)
		if got != want {
			t.Errorf("%s leads to %s, want %s", from, got, want)
		}
		if !isPage(out, got) {
			t.Errorf("%s leads to %s, which has no page", from, got)
		}
	}
}

// Cloudflare counts every rule after the first dynamic one as dynamic
// and takes at most 100 of them, so the static rules come first. Checked
// on the real content, with all its versions.
func TestRedirectsStaticFirst(t *testing.T) {
	s, err := load(filepath.Join("..", "..", "content"))
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "public")
	if _, err := s.write(out, filepath.Join("..", "..", "static")); err != nil {
		t.Fatal(err)
	}
	statics, dynamics := 0, 0
	for _, r := range redirects(t, out) {
		switch {
		case isDynamic(r.from):
			dynamics++
		case dynamics > 0:
			t.Fatalf("a static rule after a dynamic one: %s", r.from)
		default:
			statics++
		}
	}
	if dynamics > 100 || statics > 2000 {
		t.Errorf("%d static and %d dynamic rules; Cloudflare takes 2000 and 100", statics, dynamics)
	}
}

// The old language codes and "latest" are redirects, so no page can take
// one of those names.
func TestPageNamesTakenByRedirects(t *testing.T) {
	for name, want := range map[string]string{
		"0.14/pt-br.md":    `content/0.14/pt-br.md: "pt-br" is an old language code`,
		"0.14/en/rules.md": `content/0.14/en/rules.md: "en" is an old language code`,
		"0.14/latest.md":   `"latest" is an old language code or "latest"`,
	} {
		fs := files()
		fs[name] = "# A page with a taken name\n"
		if _, err := buildSite(t, fs); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err = %v, want one with %q", name, err, want)
		}
	}
}

func TestPictures(t *testing.T) {
	fs := files()
	fs["0.14/rules.md"] += "\n![The rules page](shots/rules.png)\n\n[The full picture](shots/rules.png)\n"
	fs["0.14/shots/rules.png"] = "png"
	fs["0.14/player/feel.md"] += "\n![Effects](effects.svg)\n"
	fs["0.14/player/effects.svg"] = "<svg/>"
	out := mustBuild(t, fs)
	rules := read(t, out, "0.14/rules/index.html")
	has(t, rules, `<img src="/0.14/shots/rules.png" alt="The rules page">`, true)
	has(t, rules, `<a href="/0.14/shots/rules.png">The full picture</a>`, true)
	has(t, read(t, out, "0.14/player/feel/index.html"), `<img src="/0.14/player/effects.svg" alt="Effects">`, true)
	if got := read(t, out, "0.14/shots/rules.png"); got != "png" {
		t.Errorf("the copied picture holds %q", got)
	}
	if got := read(t, out, "0.14/player/effects.svg"); got != "<svg/>" {
		t.Errorf("the copied picture holds %q", got)
	}
}

func TestSearchIndex(t *testing.T) {
	out := mustBuild(t, files())
	index := read(t, out, "0.14/search.json")
	has(t, index, `"t":"Rules","u":"rules/"`, true)
	has(t, index, `"t":"Effects","u":"player/feel/"`, true)
	has(t, index, `["Event rules","event-rules","Text."]`, true)
	has(t, read(t, out, "0.14/rules/index.html"), `data-index="/0.14/search.json" data-base="/0.14/"`, true)
	has(t, read(t, out, "0.13/rules/index.html"), `data-index="/0.13/search.json" data-base="/0.13/"`, true)
}

func TestCodeAndRawHTML(t *testing.T) {
	out := mustBuild(t, files())
	index := read(t, out, "0.14/index.html")
	has(t, index, `<pre class="code" tabindex="0"><code class="language-jsonc">`, true)
	has(t, index, `<span class="nt">&#34;id&#34;</span>`, true)
	has(t, index, `<span class="c1">// a comment`, true)
	has(t, index, "<script>alert(1)</script>", false)
	rules := read(t, out, "0.14/rules/index.html")
	has(t, rules, `<div class="table" tabindex="0"><table>`, true)
}

func TestNotFoundPages(t *testing.T) {
	out := mustBuild(t, files())
	root := read(t, out, "404.html")
	has(t, root, "<h1>Page not found</h1>", true)
	has(t, root, `<a class="button" href="/0.14/">`, true)
	nf := read(t, out, "0.13/404.html")
	has(t, nf, `<meta name="robots" content="noindex">`, true)
	has(t, nf, `<a class="button" href="/0.13/">`, true)
}

func TestBrokenLinksStopTheBuild(t *testing.T) {
	for name, c := range map[string]struct{ file, text, want string }{
		"missing page":    {"0.14/rules.md", "# Rules\n\n[x](nope.md)\n\n## Event rules\n\n## Example\n\n## Example\n", "content/0.14/rules.md: the link to \"nope.md\" goes to a page that does not exist"},
		"missing heading": {"0.14/player/feel.md", "# Effects\n\n[x](../rules.md#nope)\n", "content/0.14/player/feel.md: the link to \"../rules.md#nope\" goes to a heading that page does not have"},
		"own heading":     {"0.14/player/feel.md", "# Effects\n\n[x](#nope)\n", "goes to a heading this page does not have"},
		"not a page":      {"0.14/player/feel.md", "# Effects\n\n[x](../notes.txt)\n", "goes to a file the site does not carry"},
		"no picture":      {"0.14/player/feel.md", "# Effects\n\n[x](../shot.png)\n", `the link to "../shot.png" goes to a picture that is not there`},
	} {
		t.Run(name, func(t *testing.T) {
			fs := files()
			fs[c.file] = c.text
			_, err := buildSite(t, fs)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v\nwant one with %q", err, c.want)
			}
		})
	}
}

func TestContentChecks(t *testing.T) {
	for name, c := range map[string]struct {
		edit func(map[string]string)
		want string
	}{
		"no title":             {func(fs map[string]string) { fs["0.14/rules.md"] = "Text without a title.\n" }, "content/0.14/rules.md: the page needs a title"},
		"listed, no file":      {func(fs map[string]string) { fs["0.14/pages.json"] = `{"pages": ["index", "gone"]}` }, "lists gone, which has no content/0.14/gone.md"},
		"bad current":          {func(fs map[string]string) { fs["versions.json"] = `{"current": "9.9", "versions": [{"id": "0.14"}]}` }, "the current version \"9.9\""},
		"other site's picture": {func(fs map[string]string) { fs["0.14/rules.md"] = "# Rules\n\n![a](https://example.com/a.png)\n" }, "the site's Content-Security-Policy (img-src 'self' in static/_headers) blocks those"},
		"missing picture":      {func(fs map[string]string) { fs["0.14/rules.md"] = "# Rules\n\n![a](shot.png)\n" }, `content/0.14/rules.md: the picture "shot.png" goes to a picture that is not there`},
		"picture outside":      {func(fs map[string]string) { fs["0.14/rules.md"] = "# Rules\n\n![a](../x.png)\n" }, "goes outside the version's folder"},
	} {
		t.Run(name, func(t *testing.T) {
			fs := files()
			c.edit(fs)
			_, err := buildSite(t, fs)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v\nwant one with %q", err, c.want)
			}
		})
	}
}

func TestWarnings(t *testing.T) {
	fs := files()
	fs["0.14/unlisted.md"] = "# Unlisted\n"
	s, err := load(writeContent(t, fs))
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(s.warnings, "\n")
	if w := "content/0.14/pages.json does not list unlisted"; !strings.Contains(all, w) {
		t.Errorf("no warning %q in:\n%s", w, all)
	}
	if order := s.version("0.14").order; order[len(order)-1] != "unlisted" {
		t.Errorf("order = %q: an unlisted page goes last", order)
	}
}

// The real content builds: every link and heading in it works.
func TestRealContent(t *testing.T) {
	s, err := load(filepath.Join("..", "..", "content"))
	if err != nil {
		t.Fatal(err)
	}
	s.base = "https://bururu.dev"
	if _, err := s.write(filepath.Join(t.TempDir(), "public"), filepath.Join("..", "..", "static")); err != nil {
		t.Fatal(err)
	}
}
