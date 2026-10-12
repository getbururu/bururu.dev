package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// files is a small content folder: two versions, English, a French
// translation of one page, an Arabic one of another, and German with no
// translation at all.
func files() map[string]string {
	return map[string]string{
		"languages.json": `[
			{"code": "en", "name": "English"},
			{"code": "fr", "name": "Français"},
			{"code": "ar", "name": "العربية", "dir": "rtl"},
			{"code": "de", "name": "Deutsch"}
		]`,
		"versions.json": `{"current": "0.14", "versions": [{"id": "0.14", "api": "0.12"}, {"id": "0.13", "api": "0.11"}]}`,
		"ui/en.json":    uiJSON("en"),
		"ui/fr.json":    `{"fallback": "Pas encore traduite.", "current": "actuelle"}`,
		"ui/ar.json":    `{"fallback": "غير مترجمة بعد."}`,

		"0.14/pages.json": `{"pages": ["index", "rules", "player/feel"]}`,
		"0.14/en/index.md": "# Making mods\n\nA mod tells Bururu about one game: which program it is and what it reads while it runs.\n\n" +
			"## Start here\n\nRead [the rules](rules.md#event-rules), [Effects](player/feel.md) and [below](#start-here).\n\n" +
			"```jsonc\n{\n  // a comment\n  \"id\": \"my-mod\"\n}\n```\n\n<script>alert(1)</script>\n",
		"0.14/en/rules.md":       "# Rules\n\nBack to the [modding guide](index.md).\n\n## Event rules\n\nText.\n\n## Example\n\n## Example\n\n| Field | What |\n|---|---|\n| `on` | the input |\n",
		"0.14/en/player/feel.md": "# Effects\n\nSee [Rules](../rules.md) and [the start](../index.md#start-here).\n",
		"0.14/fr/rules.md":       "# Règles\n\nRetour au [guide](index.md).\n\n## Règles d'événement\n\nTexte.\n\n## Exemple\n\n## Exemple\n\nFin.\n",
		"0.14/ar/index.md":       "# صنع المودات\n\n## ابدأ هنا {#start-here}\n\nاقرأ [القواعد](rules.md#event-rules).\n",

		"0.13/pages.json":  `{"pages": ["index", "rules"]}`,
		"0.13/en/index.md": "# Making mods\n\nSee [Rules](rules.md).\n",
		"0.13/en/rules.md": "# Rules\n\n## Event rules\n",
	}
}

func uiJSON(lang string) string {
	var b strings.Builder
	b.WriteString("{")
	for i, k := range uiKeys {
		if i > 0 {
			b.WriteString(",")
		}
		v := lang + ":" + k
		switch k {
		case "old_version":
			v = "These docs are for {version}; the newest are for {current}."
		case "mod_api":
			v = "mod API {api}"
		case "fallback":
			v = "Not translated yet."
		}
		b.WriteString(`"` + k + `": "` + v + `"`)
	}
	b.WriteString("}")
	return b.String()
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

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Finding a value with brr memory": "finding-a-value-with-brr-memory",
		"The player's overrides":          "the-players-overrides",
		"sensor.set":                      "sensorset",
		"Games on macOS":                  "games-on-macos",
		"Layers and stages (advanced)":    "layers-and-stages-advanced",
		"On Linux":                        "on-linux",
		"على Linux":                       "على-linux",
		"Règles d'événement":              "règles-dévénement",
	} {
		if got := slug(in); got != want {
			t.Errorf("slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLinksPointAtTheSite(t *testing.T) {
	out := mustBuild(t, files())
	en := read(t, out, "0.14/en/index.html")
	has(t, en, `href="/0.14/en/rules/#event-rules"`, true)
	has(t, en, `href="/0.14/en/player/feel/"`, true)
	has(t, en, `href="#start-here"`, true)
	feel := read(t, out, "0.14/en/player/feel/index.html")
	has(t, feel, `href="/0.14/en/rules/"`, true)
	has(t, feel, `href="/0.14/en/#start-here"`, true)

	// The same English page shown in French links to the French pages.
	fr := read(t, out, "0.14/fr/index.html")
	has(t, fr, `href="/0.14/fr/rules/#event-rules"`, true)
	has(t, fr, `href="/0.14/fr/player/feel/"`, true)
}

func TestHeadingIDs(t *testing.T) {
	out := mustBuild(t, files())
	rules := read(t, out, "0.14/en/rules/index.html")
	has(t, rules, `<h2 id="example">`, true)
	has(t, rules, `<h2 id="example-1">`, true)
	// The French page has the same headings, so it keeps the English ids
	// and links into it still work.
	fr := read(t, out, "0.14/fr/rules/index.html")
	has(t, fr, `<h2 id="event-rules">Règles d'événement`, true)
	has(t, fr, `<h2 id="example-1">`, true)
}

func TestFallbackShowsEnglish(t *testing.T) {
	out := mustBuild(t, files())
	fr := read(t, out, "0.14/fr/index.html")
	has(t, fr, `<html lang="fr" dir="ltr">`, true)
	has(t, fr, `<p class="note">Pas encore traduite.</p>`, true)
	has(t, fr, `<article class="doc" lang="en" dir="ltr">`, true)
	has(t, fr, `<link rel="canonical" href="https://bururu.dev/0.14/en/">`, true)
	has(t, fr, "A mod tells Bururu about one game", true)
	// Menu strings fr.json lacks come from en.json.
	has(t, fr, "en:search_hint", true)

	tr := read(t, out, "0.14/fr/rules/index.html")
	has(t, tr, `class="note"`, false)
	has(t, tr, "<h1 id=\"rules\">Règles", true)
	has(t, tr, `<link rel="alternate" hreflang="fr" href="https://bururu.dev/0.14/fr/rules/">`, true)
	has(t, tr, `<link rel="alternate" hreflang="en" href="https://bururu.dev/0.14/en/rules/">`, true)
	// The sidebar names untranslated pages in English, marked as such.
	has(t, tr, `<li lang="en">`, true)
}

func TestVersions(t *testing.T) {
	out := mustBuild(t, files())
	redirects := read(t, out, "_redirects")
	for _, line := range []string{
		"/ /0.14/en/ 302",
		"/latest/* /0.14/:splat 302",
		"/0.13/ /0.13/en/ 302",
		"/fr/* /0.14/fr/:splat 302",
		"/rules /0.14/en/rules/ 302",
		"/player/feel/ /0.14/en/player/feel/ 302",
		// a version without a language
		"/0.14/rules/ /0.14/en/rules/ 302",
		"/0.14/player/feel /0.14/en/player/feel/ 302",
		"/0.13/rules/ /0.13/en/rules/ 302",
	} {
		has(t, redirects, line+"\n", true)
	}
	has(t, redirects, "/0.13/player/feel", false)

	feel := read(t, out, "0.14/de/player/feel/index.html")
	// 0.13 has no such page, so its menu entry goes to its start page.
	has(t, feel, `<a href="/0.13/de/">0.13<span class="item-note">mod API 0.11</span></a>`, true)
	has(t, feel, `<a href="/0.14/de/player/feel/" aria-current="true">0.14 (en:current)<span class="item-note">mod API 0.12</span></a>`, true)

	old := read(t, out, "0.13/fr/rules/index.html")
	has(t, old, `These docs are for 0.13; the newest are for 0.14. <a href="/0.14/fr/rules/">`, true)
	has(t, read(t, out, "0.14/en/rules/index.html"), "the newest are for", false)
	if _, err := os.Stat(filepath.Join(out, "0.13", "en", "player")); err == nil {
		t.Error("0.13 got a page it does not have")
	}

	sitemap := read(t, out, "sitemap.xml")
	has(t, sitemap, "<loc>https://bururu.dev/0.14/fr/rules/</loc>", true)
	has(t, sitemap, "<loc>https://bururu.dev/0.14/fr/</loc>", false)
	has(t, sitemap, "/0.13/", false)
}

// Cloudflare counts every rule after the first dynamic one as dynamic
// and takes at most 100 of them, so the static rules come first. Checked
// on the real content, with all its versions and languages.
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
	for _, line := range strings.Split(read(t, out, "_redirects"), "\n") {
		from, _, ok := strings.Cut(line, " ")
		if !ok || strings.HasPrefix(line, "#") {
			continue
		}
		switch {
		case isDynamic(from):
			dynamics++
		case dynamics > 0:
			t.Fatalf("a static rule after a dynamic one: %q", line)
		default:
			statics++
		}
	}
	if dynamics > 100 || statics > 2000 {
		t.Errorf("%d static and %d dynamic rules; Cloudflare takes 2000 and 100", statics, dynamics)
	}
}

// A language code in lower case leads to that language, and no redirect
// covers an address that has a page.
func TestRedirectsForLanguageCodes(t *testing.T) {
	fs := files()
	fs["languages.json"] = `[{"code": "en", "name": "English"}, {"code": "pt-BR", "name": "Português (Brasil)"}]`
	out := mustBuild(t, fs)
	redirects := read(t, out, "_redirects")
	for _, line := range []string{
		"/pt-BR/* /0.14/pt-BR/:splat 302",
		"/pt-br /0.14/pt-BR/ 302",
		"/pt-br/* /0.14/pt-BR/:splat 302",
	} {
		has(t, redirects, line+"\n", true)
	}
	for _, line := range strings.Split(redirects, "\n") {
		from, _, ok := strings.Cut(line, " ")
		if !ok || strings.HasPrefix(line, "#") {
			continue
		}
		file := strings.TrimPrefix(strings.TrimSuffix(from, "*"), "/")
		if strings.HasSuffix(from, "*") {
			continue // a splat: under a language or "latest", where no page is
		}
		for _, f := range []string{filepath.Join(file, "index.html"), file + ".html"} {
			if st, err := os.Stat(filepath.Join(out, filepath.FromSlash(f))); err == nil && !st.IsDir() {
				t.Errorf("the redirect %q covers the page %s", line, f)
			}
		}
	}

	fs["0.14/en/pt-br.md"] = "# A page named like a language\n"
	if _, err := buildSite(t, fs); err == nil || !strings.Contains(err.Error(), `"pt-br" is a language`) {
		t.Errorf("err = %v, want one about the page's name", err)
	}
}

func TestPictures(t *testing.T) {
	fs := files()
	fs["0.14/en/rules.md"] += "\n![The rules page](shots/rules.png)\n\n[The full picture](shots/rules.png)\n"
	fs["0.14/en/shots/rules.png"] = "en"
	fs["0.14/en/player/feel.md"] += "\n![Effects](effects.svg)\n"
	fs["0.14/en/player/effects.svg"] = "<svg/>"
	fs["0.14/ar/player/effects.svg"] = "<svg>ar</svg>"
	fs["0.14/fr/rules.md"] += "\n![Les règles](shots/rules.png)\n"
	out := mustBuild(t, fs)
	has(t, read(t, out, "0.14/en/rules/index.html"), `<img src="/0.14/en/shots/rules.png" alt="The rules page">`, true)
	has(t, read(t, out, "0.14/en/rules/index.html"), `<a href="/0.14/en/shots/rules.png">The full picture</a>`, true)
	// The French page has no picture of its own, so it shows the English one.
	has(t, read(t, out, "0.14/fr/rules/index.html"), `<img src="/0.14/en/shots/rules.png" alt="Les règles">`, true)
	// The English page shown in Arabic shows the English picture.
	has(t, read(t, out, "0.14/ar/player/feel/index.html"), `<img src="/0.14/en/player/effects.svg" alt="Effects">`, true)
	if got := read(t, out, "0.14/en/shots/rules.png"); got != "en" {
		t.Errorf("the copied picture holds %q", got)
	}
	if got := read(t, out, "0.14/ar/player/effects.svg"); got != "<svg>ar</svg>" {
		t.Errorf("the Arabic picture holds %q", got)
	}
}

func TestRightToLeft(t *testing.T) {
	out := mustBuild(t, files())
	ar := read(t, out, "0.14/ar/index.html")
	has(t, ar, `<html lang="ar" dir="rtl">`, true)
	has(t, ar, `<article class="doc">`, true) // Arabic text: the page's own direction
	has(t, ar, `<h2 id="start-here">`, true)
	has(t, ar, `href="/0.14/ar/rules/#event-rules"`, true)

	// An English page in the Arabic docs: the menus right to left, the
	// text left to right.
	rules := read(t, out, "0.14/ar/rules/index.html")
	has(t, rules, `<html lang="ar" dir="rtl">`, true)
	has(t, rules, `<p class="note">غير مترجمة بعد.</p>`, true)
	has(t, rules, `<article class="doc" lang="en" dir="ltr">`, true)
	has(t, read(t, out, "0.14/en/index.html"), `<pre class="code" tabindex="0" dir="ltr">`, true)
}

func TestNoPhysicalSidesInCSS(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "static", "assets", "site.css"))
	if err != nil {
		t.Fatal(err)
	}
	css := regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(string(b), "")
	physical := regexp.MustCompile(`(?m)(^|[\s;{])((margin|padding|border)-(left|right)[a-z-]*|left|right|float|clear)\s*:|text-align\s*:\s*(left|right)`)
	for i, line := range strings.Split(css, "\n") {
		if m := physical.FindString(line); m != "" {
			t.Errorf("site.css line %d: %q uses a physical side; use the inline-start or inline-end form so right-to-left pages mirror", i+1, strings.TrimSpace(line))
		}
	}
}

func TestSearchIndex(t *testing.T) {
	out := mustBuild(t, files())
	en := read(t, out, "0.14/en/search.json")
	has(t, en, `"t":"Rules","u":"rules/"`, true)
	has(t, en, `["Event rules","event-rules","Text."]`, true)
	// French has a translated page, so it gets its own index; German has
	// none and reads the English one.
	has(t, read(t, out, "0.14/fr/search.json"), `"t":"Règles"`, true)
	if _, err := os.Stat(filepath.Join(out, "0.14", "de", "search.json")); err == nil {
		t.Error("German got its own search index")
	}
	de := read(t, out, "0.14/de/rules/index.html")
	has(t, de, `data-index="/0.14/en/search.json" data-base="/0.14/de/"`, true)
}

func TestCodeAndRawHTML(t *testing.T) {
	out := mustBuild(t, files())
	en := read(t, out, "0.14/en/index.html")
	has(t, en, `<code class="language-jsonc">`, true)
	has(t, en, `<span class="nt">&#34;id&#34;</span>`, true)
	has(t, en, `<span class="c1">// a comment`, true)
	has(t, en, "<script>alert(1)</script>", false)
	rules := read(t, out, "0.14/en/rules/index.html")
	has(t, rules, `<div class="table" tabindex="0"><table>`, true)
}

func TestNotFoundPages(t *testing.T) {
	out := mustBuild(t, files())
	has(t, read(t, out, "404.html"), "en:not_found", true)
	nf := read(t, out, "0.14/ar/404.html")
	has(t, nf, `<html lang="ar" dir="rtl">`, true)
	has(t, nf, `<meta name="robots" content="noindex">`, true)
	has(t, nf, `<a class="button" href="/0.14/ar/">`, true)
}

func TestBrokenLinksStopTheBuild(t *testing.T) {
	for name, c := range map[string]struct{ file, text, want string }{
		"missing page":    {"0.14/en/rules.md", "# Rules\n\n[x](nope.md)\n\n## Event rules\n\n## Example\n\n## Example\n", "content/0.14/en/rules.md: the link to \"nope.md\" goes to a page that does not exist"},
		"missing heading": {"0.14/en/player/feel.md", "# Effects\n\n[x](../rules.md#nope)\n", "content/0.14/en/player/feel.md: the link to \"../rules.md#nope\" goes to a heading that page does not have"},
		"own heading":     {"0.14/en/player/feel.md", "# Effects\n\n[x](#nope)\n", "goes to a heading this page does not have"},
		"not a page":      {"0.14/en/player/feel.md", "# Effects\n\n[x](../notes.txt)\n", "goes to a file the site does not carry"},
		"no picture":      {"0.14/en/player/feel.md", "# Effects\n\n[x](../shot.png)\n", `the link to "../shot.png" goes to a picture that is not there`},
		// The French page's headings changed shape, so it lost the
		// English ids, and the English page's link into it breaks.
		"translated heading": {"0.14/fr/rules.md", "# Règles\n\n## Règles d'événement\n\nTexte.\n", "content/0.14/fr/rules.md does not have: give that heading the English id, as in \"## Heading {#event-rules}\""},
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
		"no title":             {func(fs map[string]string) { fs["0.14/en/rules.md"] = "Text without a title.\n" }, "content/0.14/en/rules.md: the page needs a title"},
		"listed, no file":      {func(fs map[string]string) { fs["0.14/pages.json"] = `{"pages": ["index", "gone"]}` }, "lists gone, which has no content/0.14/en/gone.md"},
		"no english":           {func(fs map[string]string) { fs["languages.json"] = `[{"code": "fr", "name": "Français"}]` }, "en is missing"},
		"bad current":          {func(fs map[string]string) { fs["versions.json"] = `{"current": "9.9", "versions": [{"id": "0.14"}]}` }, "the current version \"9.9\""},
		"ui key missing":       {func(fs map[string]string) { fs["ui/en.json"] = `{"skip": "Skip"}` }, "content/ui/en.json: \"site_name\" is missing"},
		"other site's picture": {func(fs map[string]string) { fs["0.14/en/rules.md"] = "# Rules\n\n![a](https://example.com/a.png)\n" }, "the site's Content-Security-Policy (img-src 'self' in static/_headers) blocks those"},
		"missing picture":      {func(fs map[string]string) { fs["0.14/en/rules.md"] = "# Rules\n\n![a](shot.png)\n" }, `content/0.14/en/rules.md: the picture "shot.png" goes to a picture that is not there`},
		"picture outside":      {func(fs map[string]string) { fs["0.14/en/rules.md"] = "# Rules\n\n![a](../../x.png)\n" }, "goes outside the language's folder"},
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
	fs["0.14/fr/extra.md"] = "# Extra\n"
	fs["0.14/xx/index.md"] = "# Unknown language\n"
	fs["0.14/en/unlisted.md"] = "# Unlisted\n"
	s, err := load(writeContent(t, fs))
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(s.warnings, "\n")
	for _, w := range []string{
		"content/0.14/fr/extra.md has no English page",
		"content/0.14/xx is not a language in languages.json",
		"content/0.14/pages.json does not list unlisted",
		"content/ui/de.json is missing",
	} {
		if !strings.Contains(all, w) {
			t.Errorf("no warning %q in:\n%s", w, all)
		}
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
