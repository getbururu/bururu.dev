package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// language is one entry of content/languages.json.
type language struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Dir  string `json:"dir,omitempty"` // "rtl" for a language written right to left
}

func (l language) dir() string {
	if l.Dir == "rtl" {
		return "rtl"
	}
	return "ltr"
}

type versionsFile struct {
	Current  string `json:"current"`
	Versions []struct {
		ID  string `json:"id"`
		API string `json:"api"`
	} `json:"versions"`
}

type site struct {
	base     string
	langs    []language
	versions []*docVersion // newest first, as versions.json lists them
	current  string
	ui       map[string]map[string]string // language -> key -> text
	warnings []string
}

// docVersion is the docs of one app version, such as 0.14.
type docVersion struct {
	id       string
	api      string
	order    []string                     // page ids, as the sidebar lists them
	pages    map[string]map[string]*page  // language -> page id -> page; only pages with a file
	pictures map[string]map[string]string // language -> path in its folder ("player/shot.png") -> the file
}

// page is one Markdown file, parsed.
type page struct {
	id     string // "index", "rules", "player/feel"
	lang   string
	file   string // for messages, such as content/0.14/fr/rules.md
	src    []byte
	doc    ast.Node
	title  string
	heads  []heading
	ids    map[string]bool
	links  []*ast.Link
	dests  [][]byte // each link's destination as written in the file
	images []*ast.Image
}

type heading struct {
	level int
	text  string
	id    string
}

func (s *site) lang(code string) language {
	for _, l := range s.langs {
		if l.Code == code {
			return l
		}
	}
	return language{Code: code, Name: code}
}

func (s *site) version(id string) *docVersion {
	for _, v := range s.versions {
		if v.id == id {
			return v
		}
	}
	return nil
}

// show returns the page that shows for id in a language: its translation,
// or else the English page.
func (v *docVersion) show(lang, id string) (p *page, translated bool) {
	if p := v.pages[lang][id]; p != nil {
		return p, true
	}
	return v.pages["en"][id], lang == "en"
}

func (v *docVersion) has(id string) bool {
	return v.pages["en"][id] != nil
}

func load(content string) (*site, error) {
	s := &site{ui: map[string]map[string]string{}}
	if err := readJSON(filepath.Join(content, "languages.json"), &s.langs); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, l := range s.langs {
		if l.Code == "" || l.Name == "" || strings.ContainsAny(l.Code, "/\\ ") {
			return nil, fmt.Errorf("languages.json: each language needs a code and a name: %+v", l)
		}
		if seen[l.Code] {
			return nil, fmt.Errorf("languages.json: %s is there twice", l.Code)
		}
		if l.Dir != "" && l.Dir != "rtl" && l.Dir != "ltr" {
			return nil, fmt.Errorf("languages.json: %s: dir is rtl or ltr, not %q", l.Code, l.Dir)
		}
		seen[l.Code] = true
	}
	if !seen["en"] {
		return nil, errors.New("languages.json: en is missing; every page falls back to it")
	}

	for _, l := range s.langs {
		strs := map[string]string{}
		err := readJSON(filepath.Join(content, "ui", l.Code+".json"), &strs)
		if errors.Is(err, fs.ErrNotExist) && l.Code != "en" {
			s.warnings = append(s.warnings, fmt.Sprintf("content/ui/%s.json is missing; %s shows its menus in English", l.Code, l.Name))
		} else if err != nil {
			return nil, err
		}
		s.ui[l.Code] = strs
	}
	for _, key := range uiKeys {
		if s.ui["en"][key] == "" {
			return nil, fmt.Errorf("content/ui/en.json: %q is missing", key)
		}
	}

	var vf versionsFile
	if err := readJSON(filepath.Join(content, "versions.json"), &vf); err != nil {
		return nil, err
	}
	s.current = vf.Current
	for _, fv := range vf.Versions {
		v, err := s.loadVersion(content, fv.ID)
		if err != nil {
			return nil, err
		}
		v.api = fv.API
		s.versions = append(s.versions, v)
	}
	if s.version(s.current) == nil {
		return nil, fmt.Errorf("versions.json: the current version %q is not in its list", s.current)
	}
	return s, nil
}

func (s *site) loadVersion(content, id string) (*docVersion, error) {
	v := &docVersion{id: id, pages: map[string]map[string]*page{}, pictures: map[string]map[string]string{}}
	dir := filepath.Join(content, id)
	var pj struct {
		Pages []string `json:"pages"`
	}
	if err := readJSON(filepath.Join(dir, "pages.json"), &pj); err != nil {
		return nil, err
	}

	md := newMarkdown()
	// files reads a language's pages, and notes its pictures in v.
	files := func(lang string) (map[string][]byte, error) {
		out := map[string][]byte{}
		root := filepath.Join(dir, lang)
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			if isPicture(p) {
				if v.pictures[lang] == nil {
					v.pictures[lang] = map[string]string{}
				}
				v.pictures[lang][filepath.ToSlash(rel)] = p
				return nil
			}
			if filepath.Ext(p) != ".md" {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			out[strings.TrimSuffix(filepath.ToSlash(rel), ".md")] = b
			return nil
		})
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return out, err
	}

	en, err := files("en")
	if err != nil {
		return nil, err
	}
	if len(en) == 0 {
		return nil, fmt.Errorf("content/%s/en has no pages", id)
	}
	v.pages["en"] = map[string]*page{}
	for pid := range en {
		// /<version>/<page>/ and /<page>/ lead to the English page, so a
		// page cannot take a name those addresses already use.
		first, _, _ := strings.Cut(pid, "/")
		if first == "latest" || s.isLangCode(first) {
			return nil, fmt.Errorf("content/%s/en/%s.md: %q is a language or \"latest\" in the site's addresses; give the page another name", id, pid, first)
		}
	}
	for pid, src := range en {
		p, err := parsePage(md, pid, "en", fmt.Sprintf("content/%s/en/%s.md", id, pid), src, nil)
		if err != nil {
			return nil, err
		}
		v.pages["en"][pid] = p
	}

	// The order: pages.json, then English pages it does not list.
	listed := map[string]bool{}
	for _, pid := range pj.Pages {
		if !v.has(pid) {
			return nil, fmt.Errorf("content/%s/pages.json lists %s, which has no content/%s/en/%s.md", id, pid, id, pid)
		}
		if !listed[pid] {
			listed[pid] = true
			v.order = append(v.order, pid)
		}
	}
	var extra []string
	for pid := range en {
		if !listed[pid] {
			extra = append(extra, pid)
		}
	}
	sort.Strings(extra)
	for _, pid := range extra {
		s.warnings = append(s.warnings, fmt.Sprintf("content/%s/pages.json does not list %s; it goes at the end of the sidebar", id, pid))
		v.order = append(v.order, pid)
	}

	known := map[string]bool{}
	for _, l := range s.langs {
		known[l.Code] = true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() != "en" && !known[e.Name()] {
			s.warnings = append(s.warnings, fmt.Sprintf("content/%s/%s is not a language in languages.json; it is left out", id, e.Name()))
		}
	}

	for _, l := range s.langs {
		if l.Code == "en" {
			continue
		}
		tr, err := files(l.Code)
		if err != nil {
			return nil, err
		}
		for pid, src := range tr {
			if !v.has(pid) {
				s.warnings = append(s.warnings, fmt.Sprintf("content/%s/%s/%s.md has no English page; it is left out", id, l.Code, pid))
				continue
			}
			p, err := parsePage(md, pid, l.Code, fmt.Sprintf("content/%s/%s/%s.md", id, l.Code, pid), src, v.pages["en"][pid])
			if err != nil {
				return nil, err
			}
			if v.pages[l.Code] == nil {
				v.pages[l.Code] = map[string]*page{}
			}
			v.pages[l.Code][pid] = p
		}
	}

	// The pictures: now that every language's files are known.
	for _, l := range s.langs {
		for _, p := range v.pages[l.Code] {
			for _, img := range p.images {
				dest := string(img.Destination)
				to, err := v.picture(p, dest)
				if err != "" {
					return nil, fmt.Errorf("%s: the picture %q %s", p.file, dest, err)
				}
				img.Destination = []byte(to)
			}
		}
	}
	return v, nil
}

// isLangCode says whether s is a language's code, in any case: addresses
// such as /pt-br/ lead to that language too.
func (s *site) isLangCode(code string) bool {
	for _, l := range s.langs {
		if strings.EqualFold(l.Code, code) {
			return true
		}
	}
	return false
}

// pageURL is the address of a page: /0.14/en/ for the start page,
// /0.14/en/rules/ for the others.
func pageURL(version, lang, id string) string {
	if id == "index" {
		return "/" + version + "/" + lang + "/"
	}
	return "/" + version + "/" + lang + "/" + id + "/"
}

// pageFile is where pageURL's page goes in the output folder.
func pageFile(version, lang, id string) string {
	if id == "index" {
		return path.Join(version, lang, "index.html")
	}
	return path.Join(version, lang, id, "index.html")
}

func readJSON(file string, v any) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", filepath.ToSlash(file), err)
	}
	return nil
}
