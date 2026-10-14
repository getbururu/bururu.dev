package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark/ast"
)

type versionsFile struct {
	Current  string `json:"current"`
	Versions []struct {
		ID  string `json:"id"`
		API string `json:"api"`
	} `json:"versions"`
}

type site struct {
	base     string
	versions []*docVersion // newest first, as versions.json lists them
	current  string
	warnings []string
}

// docVersion is the docs of one app version, such as 0.14.
type docVersion struct {
	id       string
	api      string
	order    []string          // page ids, as the sidebar lists them
	pages    map[string]*page  // page id -> page
	pictures map[string]string // path in the version's folder ("player/shot.png") -> the file
}

// page is one Markdown file, parsed.
type page struct {
	id     string // "index", "rules", "player/feel"
	file   string // for messages, such as content/0.14/rules.md
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

// Until October 2026 every page had a language in its address, such as
// /0.14/fr/rules/, and /fr/rules/ led to the current version. Only the
// 0.14 docs had those addresses. They all lead to the English page now,
// so a page cannot take one of these names.
var (
	oldLangVersions = []string{"0.14"}
	oldLangs        = []string{"en", "fr", "it", "de", "es-ES", "ar", "cs", "hu", "ja", "ko", "pl", "pt-BR", "ru", "zh-TW", "tr", "zh-CN", "es-419", "uk", "th", "vi"}
)

// isOldLang says whether code is one of oldLangs, in any case: /pt-br/
// led to pt-BR too.
func isOldLang(code string) bool {
	for _, l := range oldLangs {
		if strings.EqualFold(l, code) {
			return true
		}
	}
	return false
}

func (s *site) version(id string) *docVersion {
	for _, v := range s.versions {
		if v.id == id {
			return v
		}
	}
	return nil
}

func (v *docVersion) has(id string) bool {
	return v.pages[id] != nil
}

func load(content string) (*site, error) {
	s := &site{}
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
	v := &docVersion{id: id, pages: map[string]*page{}, pictures: map[string]string{}}
	dir := filepath.Join(content, id)
	var pj struct {
		Pages []string `json:"pages"`
	}
	if err := readJSON(filepath.Join(dir, "pages.json"), &pj); err != nil {
		return nil, err
	}

	md := newMarkdown()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		if isPicture(p) {
			v.pictures[rel] = p
			return nil
		}
		if filepath.Ext(p) != ".md" {
			return nil
		}
		pid := strings.TrimSuffix(rel, ".md")
		// The old addresses with a language and the short ones such as
		// /rules/ are redirects, so a page cannot take a name they use.
		first, _, _ := strings.Cut(pid, "/")
		if first == "latest" || isOldLang(first) {
			return fmt.Errorf("content/%s/%s.md: %q is an old language code or \"latest\" in the site's addresses; give the page another name", id, pid, first)
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		pg, err := parsePage(md, pid, fmt.Sprintf("content/%s/%s.md", id, pid), src)
		if err != nil {
			return err
		}
		v.pages[pid] = pg
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(v.pages) == 0 {
		return nil, fmt.Errorf("content/%s has no pages", id)
	}

	// The order: pages.json, then pages it does not list.
	listed := map[string]bool{}
	for _, pid := range pj.Pages {
		if !v.has(pid) {
			return nil, fmt.Errorf("content/%s/pages.json lists %s, which has no content/%s/%s.md", id, pid, id, pid)
		}
		if !listed[pid] {
			listed[pid] = true
			v.order = append(v.order, pid)
		}
	}
	var extra []string
	for pid := range v.pages {
		if !listed[pid] {
			extra = append(extra, pid)
		}
	}
	sort.Strings(extra)
	for _, pid := range extra {
		s.warnings = append(s.warnings, fmt.Sprintf("content/%s/pages.json does not list %s; it goes at the end of the sidebar", id, pid))
		v.order = append(v.order, pid)
	}

	for _, p := range v.pages {
		for _, img := range p.images {
			dest := string(img.Destination)
			to, err := v.picture(p, dest)
			if err != "" {
				return nil, fmt.Errorf("%s: the picture %q %s", p.file, dest, err)
			}
			img.Destination = []byte(to)
		}
	}
	return v, nil
}

// pageURL is the address of a page: /0.14/ for the start page,
// /0.14/rules/ for the others.
func pageURL(version, id string) string {
	if id == "index" {
		return "/" + version + "/"
	}
	return "/" + version + "/" + id + "/"
}

// pageFile is where pageURL's page goes in the output folder.
func pageFile(version, id string) string {
	if id == "index" {
		return path.Join(version, "index.html")
	}
	return path.Join(version, id, "index.html")
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
