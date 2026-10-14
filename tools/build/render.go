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

// titleSuffix goes after each page's title, in the browser's tab.
const titleSuffix = " - Bururu modding docs"

type pageData struct {
	VersionID   string
	Title       string
	Description string
	Canonical   string
	NoIndex     bool
	Home        string
	VersionMenu []menuItem
	Nav         []navGroup
	OldVersion  string
	OldLink     string
	Content     template.HTML
	Prev, Next  *navItem
	SearchIndex string
	SearchBase  string
	NotFound    bool
}

type menuItem struct {
	Label, Note, URL string
	Current          bool
}

type navGroup struct {
	Title string
	Items []*navItem
}

type navItem struct {
	Title, URL string
	Current    bool
	Heads      []navHead
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
	written := 0
	for _, v := range s.versions {
		for rel, file := range v.pictures {
			if err := copyFile(file, filepath.Join(out, v.id, filepath.FromSlash(rel))); err != nil {
				return 0, err
			}
		}
		for i, id := range v.order {
			p := v.pages[id]
			linkErrs = append(linkErrs, v.resolveLinks(p)...)
			body, err := render(md, p)
			if err != nil {
				return 0, err
			}
			d := s.data(v, id)
			d.Title = p.title + titleSuffix
			d.Description = description(p)
			d.Content = template.HTML(body)
			d.Canonical = s.base + pageURL(v.id, id)
			if i > 0 {
				d.Prev = navEntry(v, v.order[i-1])
			}
			if i+1 < len(v.order) {
				d.Next = navEntry(v, v.order[i+1])
			}
			if err := writeTemplate(tmpl, filepath.Join(out, filepath.FromSlash(pageFile(v.id, id))), d); err != nil {
				return 0, err
			}
			written++
		}
		// The nearest 404.html up the tree is what Pages shows for a
		// missing page, so each version gets its own.
		if err := s.writeNotFound(tmpl, filepath.Join(out, v.id, "404.html"), v); err != nil {
			return 0, err
		}
		if err := writeSearch(out, v); err != nil {
			return 0, err
		}
	}
	if len(linkErrs) > 0 {
		sort.Strings(linkErrs)
		return 0, errors.New("links that do not work:\n  " + strings.Join(linkErrs, "\n  "))
	}
	if err := s.writeNotFound(tmpl, filepath.Join(out, "404.html"), s.version(s.current)); err != nil {
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

// data fills in what every page of a version has: the version menu, the
// sidebar, the note on an older version. current is the page being
// made, or "" for the not-found page.
func (s *site) data(v *docVersion, current string) *pageData {
	d := &pageData{
		VersionID: v.id,
		Home:      pageURL(v.id, "index"),
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
			note = "mod API " + ov.api
		}
		label := ov.id
		if ov.id == s.current {
			label += " (current)"
		}
		d.VersionMenu = append(d.VersionMenu, menuItem{Label: label, Note: note, URL: pageURL(ov.id, id), Current: ov == v})
	}

	var main, player navGroup
	player.Title = "From the player guide"
	for _, id := range v.order {
		item := navEntry(v, id)
		if id == current {
			item.Current = true
			for _, h := range v.pages[id].heads {
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
		d.OldVersion = s.current
		cv := s.version(s.current)
		id := target
		if !cv.has(id) {
			id = "index"
		}
		d.OldLink = pageURL(cv.id, id)
	}
	d.SearchIndex, d.SearchBase = "/"+v.id+"/search.json", pageURL(v.id, "index")
	return d
}

func navEntry(v *docVersion, id string) *navItem {
	return &navItem{Title: v.pages[id].title, URL: pageURL(v.id, id)}
}

func (s *site) writeNotFound(tmpl *template.Template, file string, v *docVersion) error {
	d := s.data(v, "")
	d.NotFound = true
	d.NoIndex = true
	d.Title = "Page not found" + titleSuffix
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

// The search index of a version: for each page its title, its address
// under the version's start page, and its sections as [heading, id, text].
type searchPage struct {
	Title    string      `json:"t"`
	URL      string      `json:"u"`
	Sections [][3]string `json:"s"`
}

func writeSearch(out string, v *docVersion) error {
	var pages []searchPage
	for _, id := range v.order {
		p := v.pages[id]
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
	return writeFile(filepath.Join(out, v.id, "search.json"), b)
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

// writeRedirects writes _redirects: the root and /latest/ lead to the
// current version, a page alone (/rules/) to that page in the current
// version, and the old addresses with a language (/0.14/fr/rules/,
// /fr/rules/) to the English page. Pages applies a redirect before it
// looks for a file, so no rule here may match an address that has a page.
//
// Cloudflare takes at most 100 dynamic rules (with a splat or a
// placeholder) and counts every rule after the first dynamic one as
// dynamic too, so the static rules all come first.
func (s *site) writeRedirects(file string) error {
	cur := s.current
	var static, dynamic strings.Builder
	line := func(from, to string) {
		b := &static
		if isDynamic(from) {
			b = &dynamic
		}
		fmt.Fprintf(b, "%s %s 302\n", from, to)
	}
	line("/", "/"+cur+"/")
	line("/latest", "/"+cur+"/")
	line("/latest/*", "/"+cur+"/:splat")
	for _, l := range oldLangs {
		codes := []string{l}
		if lower := strings.ToLower(l); lower != l {
			codes = append(codes, lower)
		}
		for _, c := range codes {
			for _, vid := range oldLangVersions {
				if s.version(vid) != nil {
					line("/"+vid+"/"+c, "/"+vid+"/")
					line("/"+vid+"/"+c+"/*", "/"+vid+"/:splat")
				}
			}
			line("/"+c, "/"+cur+"/")
			line("/"+c+"/*", "/"+cur+"/:splat")
		}
	}
	for _, id := range s.version(cur).order {
		if id != "index" {
			line("/"+id, pageURL(cur, id))
			line("/"+id+"/", pageURL(cur, id))
		}
	}
	var b strings.Builder
	b.WriteString("# Made by tools/build from content/versions.json. Edit that file, not\n")
	b.WriteString("# this one. Syntax:\n")
	b.WriteString("# https://developers.cloudflare.com/pages/configuration/redirects/\n")
	b.WriteString("# The static rules come first: Cloudflare counts every rule after the\n")
	b.WriteString("# first one with a splat or a placeholder as dynamic (at most 100).\n\n")
	b.WriteString(static.String())
	b.WriteString(dynamic.String())
	return writeFile(file, []byte(b.String()))
}

// isDynamic reports whether a redirect's source has a splat or a
// placeholder, which makes it one of Cloudflare's dynamic rules.
func isDynamic(from string) bool {
	return strings.Contains(from, "*") || strings.Contains(from, "/:")
}

// writeSitemap lists the current version's pages.
func (s *site) writeSitemap(out string) error {
	v := s.version(s.current)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, id := range v.order {
		fmt.Fprintf(&b, "  <url><loc>%s</loc></url>\n", template.HTMLEscapeString(s.base+pageURL(v.id, id)))
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
