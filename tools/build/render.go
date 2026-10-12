package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark/ast"
)

//go:embed page.html
var pageTemplate string

// uiKeys are the strings of the menus and notes; content/ui/en.json must
// have them all, other languages fall back to English for any they miss.
var uiKeys = []string{
	"site_name", "title_suffix", "skip", "version", "language", "current", "mod_api",
	"pages", "player_guide", "search", "search_hint", "no_results", "fallback",
	"old_version", "old_version_link", "previous", "next", "not_found",
	"not_found_text", "start_page", "mod_list", "site_source",
}

// t is a menu string in a language, with {name} parts filled in.
func (s *site) t(lang, key string, args ...string) string {
	str := s.ui[lang][key]
	if str == "" {
		str = s.ui["en"][key]
	}
	for i := 0; i+1 < len(args); i += 2 {
		str = strings.ReplaceAll(str, "{"+args[i]+"}", args[i+1])
	}
	return str
}

type pageData struct {
	Lang        language
	Dir         string // the page's direction: rtl for Arabic
	VersionID   string
	T           map[string]string
	Title       string
	Description string
	Canonical   string
	NoIndex     bool
	Alternates  []alternate
	Home        string
	VersionMenu []menuItem
	LangMenu    []menuItem
	Nav         []navGroup
	Fallback    bool
	ContentLang string
	ContentDir  string
	OldNote     string
	OldLink     string
	OldLinkText string
	Content     template.HTML
	Prev, Next  *navItem
	SearchIndex string
	SearchBase  string
	NotFound    bool
}

type alternate struct{ Lang, URL string }

type menuItem struct {
	Label, Note, URL, Lang string
	Current                bool
}

type navGroup struct {
	Title string
	Items []*navItem
}

type navItem struct {
	Title, URL, Lang string
	Current          bool
	Heads            []navHead
}

type navHead struct{ Text, URL string }

// write builds the whole site into out and returns how many pages it
// wrote.
func (s *site) write(out, static string) (int, error) {
	if clean := filepath.Clean(out); clean == "." || clean == string(filepath.Separator) || filepath.Dir(clean) == clean {
		return 0, fmt.Errorf("-out %s: give the site a folder of its own", out)
	}
	if err := os.RemoveAll(out); err != nil {
		return 0, err
	}
	if err := copyTree(static, out); err != nil {
		return 0, err
	}
	tmpl, err := template.New("page").Parse(pageTemplate)
	if err != nil {
		return 0, err
	}
	md := newMarkdown()

	var linkErrs []string
	seenErr := map[string]bool{}
	written := 0
	for _, v := range s.versions {
		for lang, pics := range v.pictures {
			for rel, file := range pics {
				if err := copyFile(file, filepath.Join(out, v.id, lang, filepath.FromSlash(rel))); err != nil {
					return 0, err
				}
			}
		}
		for _, l := range s.langs {
			for i, id := range v.order {
				p, translated := v.show(l.Code, id)
				for _, e := range v.resolveLinks(p, l.Code) {
					if !seenErr[e] {
						seenErr[e] = true
						linkErrs = append(linkErrs, e)
					}
				}
				body, err := render(md, p)
				if err != nil {
					return 0, err
				}
				d := s.data(v, l, id)
				d.Title = p.title + " - " + s.t(l.Code, "title_suffix")
				d.Description = description(p)
				d.Content = template.HTML(body)
				d.Fallback = !translated
				d.ContentLang, d.ContentDir = p.lang, s.lang(p.lang).dir()
				d.Canonical = s.base + pageURL(v.id, l.Code, id)
				if d.Fallback {
					// The same text as the English page: search engines
					// should list that one.
					d.Canonical = s.base + pageURL(v.id, "en", id)
				} else {
					for _, al := range s.langs {
						if _, ok := v.show(al.Code, id); ok {
							d.Alternates = append(d.Alternates, alternate{al.Code, s.base + pageURL(v.id, al.Code, id)})
						}
					}
				}
				if i > 0 {
					d.Prev = s.navEntry(v, l.Code, v.order[i-1])
				}
				if i+1 < len(v.order) {
					d.Next = s.navEntry(v, l.Code, v.order[i+1])
				}
				if err := writeTemplate(tmpl, filepath.Join(out, filepath.FromSlash(pageFile(v.id, l.Code, id))), d); err != nil {
					return 0, err
				}
				written++
			}
			// The nearest 404.html up the tree is what Pages shows for a
			// missing page, so each language gets its own.
			if err := s.writeNotFound(tmpl, filepath.Join(out, v.id, l.Code, "404.html"), v, l); err != nil {
				return 0, err
			}
			if err := s.writeSearch(out, v, l); err != nil {
				return 0, err
			}
		}
	}
	if len(linkErrs) > 0 {
		sort.Strings(linkErrs)
		return 0, errors.New("links that do not work:\n  " + strings.Join(linkErrs, "\n  "))
	}
	if err := s.writeNotFound(tmpl, filepath.Join(out, "404.html"), s.version(s.current), s.lang("en")); err != nil {
		return 0, err
	}
	if err := s.writeRedirects(filepath.Join(out, "_redirects")); err != nil {
		return 0, err
	}
	if err := s.writeSitemap(out); err != nil {
		return 0, err
	}
	return written, nil
}

// data fills in what every page of a version and language has: the
// menus, the sidebar, the notes. current is the page being made, or ""
// for the not-found page.
func (s *site) data(v *docVersion, l language, current string) *pageData {
	d := &pageData{
		Lang:      l,
		Dir:       l.dir(),
		VersionID: v.id,
		T:         map[string]string{},
		Home:      pageURL(v.id, l.Code, "index"),
	}
	for _, k := range uiKeys {
		d.T[k] = s.t(l.Code, k)
	}
	target := current
	if target == "" {
		target = "index"
	}
	for _, ov := range s.versions {
		id := target
		if !ov.has(id) {
			id = "index"
		}
		note := ""
		if ov.api != "" {
			note = s.t(l.Code, "mod_api", "api", ov.api)
		}
		label := ov.id
		if ov.id == s.current {
			label += " (" + s.t(l.Code, "current") + ")"
		}
		d.VersionMenu = append(d.VersionMenu, menuItem{Label: label, Note: note, URL: pageURL(ov.id, l.Code, id), Current: ov == v})
	}
	for _, ol := range s.langs {
		d.LangMenu = append(d.LangMenu, menuItem{Label: ol.Name, URL: pageURL(v.id, ol.Code, target), Lang: ol.Code, Current: ol.Code == l.Code})
	}

	var main, player navGroup
	player.Title = s.t(l.Code, "player_guide")
	for _, id := range v.order {
		item := s.navEntry(v, l.Code, id)
		if id == current {
			item.Current = true
			p, _ := v.show(l.Code, id)
			for _, h := range p.heads {
				if h.level == 2 {
					item.Heads = append(item.Heads, navHead{h.text, "#" + h.id})
				}
			}
		}
		if strings.HasPrefix(id, "player/") {
			player.Items = append(player.Items, item)
		} else {
			main.Items = append(main.Items, item)
		}
	}
	d.Nav = append(d.Nav, main)
	if len(player.Items) > 0 {
		d.Nav = append(d.Nav, player)
	}

	if v.id != s.current {
		d.OldNote = s.t(l.Code, "old_version", "version", v.id, "current", s.current)
		cv := s.version(s.current)
		id := target
		if !cv.has(id) {
			id = "index"
		}
		d.OldLink = pageURL(cv.id, l.Code, id)
		d.OldLinkText = s.t(l.Code, "old_version_link", "current", s.current)
	}
	d.SearchIndex, d.SearchBase = s.searchFile(v, l.Code), pageURL(v.id, l.Code, "index")
	return d
}

func (s *site) navEntry(v *docVersion, lang, id string) *navItem {
	p, _ := v.show(lang, id)
	return &navItem{Title: p.title, URL: pageURL(v.id, lang, id), Lang: p.lang}
}

func (s *site) writeNotFound(tmpl *template.Template, file string, v *docVersion, l language) error {
	d := s.data(v, l, "")
	d.NotFound = true
	d.NoIndex = true
	d.Title = s.t(l.Code, "not_found") + " - " + s.t(l.Code, "title_suffix")
	d.ContentLang, d.ContentDir = l.Code, l.dir()
	return writeTemplate(tmpl, file, d)
}

// description is the first real paragraph of a page, for search engines
// and link previews: one long enough to say something, not a "Back to"
// line.
func description(p *page) string {
	for n := p.doc.FirstChild(); n != nil; n = n.NextSibling() {
		if n.Kind() != ast.KindParagraph {
			continue
		}
		t := plainText(n, p.src)
		if len([]rune(t)) < 80 {
			continue
		}
		if r := []rune(t); len(r) > 200 {
			cut := string(r[:200])
			if i := strings.LastIndexAny(cut, " "); i > 100 {
				cut = cut[:i]
			}
			t = strings.TrimRight(cut, ",.;: ") + "..."
		}
		return t
	}
	return ""
}

// searchFile is the search index a page reads. A language with no
// translated page in this version shows the English pages, so it reads
// the English index.
func (s *site) searchFile(v *docVersion, lang string) string {
	if len(v.pages[lang]) == 0 {
		lang = "en"
	}
	return "/" + v.id + "/" + lang + "/search.json"
}

// The search index: for each page its title, its address under the
// language's start page, and its sections as [heading, id, text].
type searchPage struct {
	Title    string      `json:"t"`
	URL      string      `json:"u"`
	Sections [][3]string `json:"s"`
}

func (s *site) writeSearch(out string, v *docVersion, l language) error {
	if s.searchFile(v, l.Code) != "/"+v.id+"/"+l.Code+"/search.json" {
		return nil
	}
	var pages []searchPage
	for _, id := range v.order {
		p, _ := v.show(l.Code, id)
		u := ""
		if id != "index" {
			u = id + "/"
		}
		pages = append(pages, searchPage{Title: p.title, URL: u, Sections: sections(p)})
	}
	b, err := json.Marshal(map[string]any{"pages": pages})
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(out, v.id, l.Code, "search.json"), b)
}

// sections splits a page at its level 2 and 3 headings for search.
func sections(p *page) [][3]string {
	var out [][3]string
	cur := [3]string{}
	var text []string
	flush := func() {
		cur[2] = strings.Join(text, " ")
		if cur[0] != "" || cur[2] != "" {
			out = append(out, cur)
		}
		text = nil
	}
	for n := p.doc.FirstChild(); n != nil; n = n.NextSibling() {
		if h, ok := n.(*ast.Heading); ok {
			if h.Level == 1 {
				continue
			}
			if h.Level <= 3 {
				flush()
				id, _ := h.AttributeString("id")
				cur = [3]string{plainText(h, p.src), stringOf(id), ""}
				continue
			}
		}
		if t := plainText(n, p.src); t != "" {
			text = append(text, t)
		}
	}
	flush()
	return out
}

// writeRedirects writes _redirects: the root, the version folders, and
// addresses without a version or language lead to the docs in English or
// to the current version. A language code in lower case, such as /pt-br/,
// leads to that language. Pages applies a redirect before it looks for a
// file, so no rule here may match an address that has a page.
func (s *site) writeRedirects(file string) error {
	cur := s.current
	var b strings.Builder
	b.WriteString("# Made by tools/build from content/versions.json and languages.json.\n")
	b.WriteString("# Edit those files, not this one. Syntax:\n")
	b.WriteString("# https://developers.cloudflare.com/pages/configuration/redirects/\n\n")
	line := func(from, to string) { fmt.Fprintf(&b, "%s %s 302\n", from, to) }
	both := func(from, to string) {
		line(from, to)
		line(from+"/", to)
	}
	line("/", "/"+cur+"/en/")
	line("/latest", "/"+cur+"/en/")
	line("/latest/*", "/"+cur+"/:splat")
	// A version without a language: its English pages.
	for _, v := range s.versions {
		both("/"+v.id, "/"+v.id+"/en/")
		for _, id := range v.order {
			if id != "index" {
				both("/"+v.id+"/"+id, pageURL(v.id, "en", id))
			}
		}
	}
	// A language without a version: the current version.
	for _, l := range s.langs {
		codes := []string{l.Code}
		if lower := strings.ToLower(l.Code); lower != l.Code {
			codes = append(codes, lower)
		}
		for _, c := range codes {
			line("/"+c, "/"+cur+"/"+l.Code+"/")
			line("/"+c+"/*", "/"+cur+"/"+l.Code+"/:splat")
		}
	}
	// A page alone: the current version in English.
	for _, id := range s.version(cur).order {
		if id != "index" {
			both("/"+id, pageURL(cur, "en", id))
		}
	}
	return writeFile(file, []byte(b.String()))
}

// writeSitemap lists the current version's pages that have their own
// text: the English ones and the translations.
func (s *site) writeSitemap(out string) error {
	v := s.version(s.current)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, l := range s.langs {
		for _, id := range v.order {
			if _, ok := v.show(l.Code, id); ok {
				fmt.Fprintf(&b, "  <url><loc>%s</loc></url>\n", template.HTMLEscapeString(s.base+pageURL(v.id, l.Code, id)))
			}
		}
	}
	b.WriteString("</urlset>\n")
	if err := writeFile(filepath.Join(out, "sitemap.xml"), []byte(b.String())); err != nil {
		return err
	}
	robots := "User-agent: *\nAllow: /\n\nSitemap: " + s.base + "/sitemap.xml\n"
	return writeFile(filepath.Join(out, "robots.txt"), []byte(robots))
}

func writeTemplate(tmpl *template.Template, file string, d *pageData) error {
	var b strings.Builder
	if err := tmpl.Execute(&b, d); err != nil {
		return err
	}
	return writeFile(file, []byte(b.String()))
}

func writeFile(file string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, b, 0o644)
}

// copyTree copies the static folder into the site as it is.
func copyTree(from, to string) error {
	return filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, p)
		dest := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		if path.Base(filepath.ToSlash(p)) == ".gitkeep" {
			return nil
		}
		return copyFile(p, dest)
	})
}

func copyFile(from, to string) error {
	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	dst, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}
