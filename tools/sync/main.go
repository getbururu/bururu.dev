// Command sync copies the modding guide from a Bururu checkout into
// content/<version>, with its links rewritten for this site.
//
// From the bururu.dev folder:
//
//	go run ./tools/sync -from ../bururu -version 0.14
//
// It copies docs/modding/*.md (README.md becomes index.md), the pages of
// docs/ that those pages link to (into player/), and the pictures they
// show. A link to a page it does not copy becomes plain text. The "Back to
// the ..." line at the top of a page and HTML comments are left out: the
// sidebar does the first job, and the comments are notes for the app's
// developers. It also writes the page order to content/<version>/pages.json
// and adds the version to content/versions.json.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	from := flag.String("from", "", "the Bururu checkout to copy the guide from (the folder with docs/modding)")
	version := flag.String("version", "", "the app version the guide belongs to: 0.14, 0.14.13 or 0.14.13-test all give 0.14")
	api := flag.String("api", "", "the mod API version; read from the guide when empty")
	content := flag.String("content", "content", "the content folder of this site")
	force := flag.Bool("force", false, "replace a docs version even when the guide now names another mod API for it")
	flag.Parse()
	if *from == "" || *version == "" {
		flag.Usage()
		os.Exit(2)
	}
	v, err := docsVersion(*version)
	if err != nil {
		log.Fatal(err)
	}
	res, err := sync(os.DirFS(filepath.Join(*from, "docs")), *content, v, *api, *force)
	if err != nil {
		log.Fatal(err)
	}
	for _, w := range res.warnings {
		log.Print(w)
	}
	fmt.Printf("copied %d pages into %s\n", len(res.pages), filepath.Join(*content, v))
	for _, p := range res.pages {
		if strings.HasPrefix(p, "player/") {
			fmt.Printf("  %s comes from the player guide: read it before you commit, the site is public\n", p)
		}
	}
}

// docsVersion turns an app version into the docs version: its major and
// minor number. "v0.14.13", "0.14.13-test" and "0.14" all give "0.14".
func docsVersion(s string) (string, error) {
	m := regexp.MustCompile(`^v?(\d+)\.(\d+)(\.\d+)?([-+].*)?$`).FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return "", fmt.Errorf("version %q: want major.minor, such as 0.14 or 0.14.13-test", s)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	return fmt.Sprintf("%d.%d", major, minor), nil
}

type result struct {
	pages    []string // page ids in order
	warnings []string
}

// sync reads the guide from docs (the app's docs folder) and writes it
// into content/<v>. force lets it replace a version whose mod API
// changed.
func sync(docs fs.FS, content, v, api string, force bool) (*result, error) {
	names, err := fs.Glob(docs, "modding/*.md")
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, errors.New("no docs/modding/*.md in that folder")
	}

	// The pages to copy: source path in docs -> content path.
	dest := map[string]string{}
	src := map[string]string{}
	for _, n := range names {
		b, err := fs.ReadFile(docs, n)
		if err != nil {
			return nil, err
		}
		src[n] = string(b)
		dest[n] = contentPath(n)
	}
	// One step out: the pages of docs/ the guide links to.
	for _, n := range names {
		for _, target := range linkTargets(src[n]) {
			p, _ := splitTarget(target)
			if p == "" || isExternal(p) {
				continue
			}
			full := path.Clean(path.Join(path.Dir(n), p))
			if path.Dir(full) != "." || path.Ext(full) != ".md" || full == "README.md" {
				continue
			}
			if _, ok := dest[full]; ok {
				continue
			}
			b, err := fs.ReadFile(docs, full)
			if err != nil {
				return nil, fmt.Errorf("%s links to %s: %w", n, full, err)
			}
			src[full] = string(b)
			dest[full] = contentPath(full)
		}
	}
	// The pictures those pages show, wherever in docs/ they are.
	pictures := map[string][]byte{} // content path -> the file
	for n, text := range src {
		for _, target := range linkTargets(text) {
			p, _ := splitTarget(target)
			if p == "" || isExternal(p) || !isPicture(p) {
				continue
			}
			full := path.Clean(path.Join(path.Dir(n), p))
			if _, ok := dest[full]; ok {
				continue
			}
			if !fs.ValidPath(full) {
				return nil, fmt.Errorf("%s shows the picture %s, which is outside docs/", n, p)
			}
			b, err := fs.ReadFile(docs, full)
			if err != nil {
				return nil, fmt.Errorf("%s shows the picture %s: %w", n, p, err)
			}
			dest[full] = contentPath(full)
			pictures[dest[full]] = b
		}
	}

	res := &result{}
	out := map[string]string{}
	for n, text := range src {
		text = stripComments(dropBackLines(text))
		text = rewriteLinks(text, func(target string) (string, bool) {
			p, frag := splitTarget(target)
			if p == "" || isExternal(p) {
				return target, true
			}
			to, ok := resolve(n, target, dest)
			if !ok {
				res.warnings = append(res.warnings, fmt.Sprintf("%s: the site does not carry %s, so the link is plain text now", n, p))
				return "", false
			}
			return relPath(dest[n], to) + frag, true
		})
		if err := checkPrivate(n, text); err != nil {
			return nil, err
		}
		out[dest[n]] = text
	}
	sort.Strings(res.warnings)

	if api == "" {
		api = modAPI(src["modding/sharing.md"])
		if api == "" {
			res.warnings = append(res.warnings, "could not find the mod API version in sharing.md; pass -api")
		}
	}

	versions := filepath.Join(content, "versions.json")
	if err := checkAPI(versions, v, api, force); err != nil {
		return nil, err
	}

	// Write the pages. The version's old folder goes first, so a page the
	// guide dropped leaves the site too.
	dir := filepath.Join(content, v)
	if err := os.RemoveAll(dir); err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	for p, text := range out {
		files[p] = []byte(text)
	}
	for p, b := range pictures {
		files[p] = b
	}
	for p, b := range files {
		file := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(file, b, 0o644); err != nil {
			return nil, err
		}
	}

	res.pages = pageOrder(out)
	if err := writeJSON(filepath.Join(content, v, "pages.json"), map[string]any{"pages": res.pages}); err != nil {
		return nil, err
	}
	if err := addVersion(versions, v, api); err != nil {
		return nil, err
	}
	return res, nil
}

// The kinds of pictures a page can show.
var pictureExt = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true}

func isPicture(p string) bool {
	return pictureExt[strings.ToLower(path.Ext(p))]
}

// contentPath is where a page of the app's docs goes in content/<v>.
func contentPath(docsPath string) string {
	if rest, ok := strings.CutPrefix(docsPath, "modding/"); ok {
		if rest == "README.md" {
			return "index.md"
		}
		return rest
	}
	return "player/" + docsPath
}

// resolve finds the content path for a link in the page at docs path n.
func resolve(n, target string, dest map[string]string) (string, bool) {
	p, _ := splitTarget(target)
	full := path.Clean(path.Join(path.Dir(n), p))
	to, ok := dest[full]
	return to, ok
}

func splitTarget(t string) (p, frag string) {
	if i := strings.IndexByte(t, '#'); i >= 0 {
		return t[:i], t[i:]
	}
	return t, ""
}

var scheme = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:`)

func isExternal(p string) bool {
	return scheme.MatchString(p) || strings.HasPrefix(p, "//")
}

// relPath is the link from the page at content path from to the one at to.
func relPath(from, to string) string {
	fromDir := strings.Split(path.Dir(from), "/")
	toParts := strings.Split(to, "/")
	if fromDir[0] == "." {
		fromDir = nil
	}
	i := 0
	for i < len(fromDir) && i < len(toParts)-1 && fromDir[i] == toParts[i] {
		i++
	}
	var b strings.Builder
	for range fromDir[i:] {
		b.WriteString("../")
	}
	b.WriteString(strings.Join(toParts[i:], "/"))
	return b.String()
}

var backLine = regexp.MustCompile(`^Back to the \[[^\]]*\]\([^)\s]*\)\.\s*$`)

// dropBackLines leaves out the "Back to the [X](target)." lines, with the
// blank line after them. On the site the sidebar is the way back.
func dropBackLines(text string) string {
	lines := strings.Split(text, "\n")
	var out []string
	var fence codeFence
	for i := 0; i < len(lines); i++ {
		if !fence.inCode(lines[i]) && backLine.MatchString(lines[i]) {
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "" {
				i++
			}
			continue
		}
		out = append(out, lines[i])
	}
	return strings.Join(out, "\n")
}

// stripComments leaves out HTML comments outside code blocks and code
// spans. A line that held only a comment goes too, and so does the blank
// line after it when the line before it is blank, so no gap is left.
func stripComments(text string) string {
	var out []string
	var fence codeFence
	inComment, skipBlank := false, false
	for _, line := range strings.SplitAfter(text, "\n") {
		if skipBlank {
			skipBlank = false
			if strings.TrimSpace(line) == "" {
				continue
			}
		}
		if !inComment && fence.inCode(line) {
			out = append(out, line)
			continue
		}
		stripped, still, cut := stripCommentsInLine(line, inComment)
		inComment = still
		if cut && strings.TrimSpace(stripped) == "" {
			if len(out) == 0 || strings.TrimSpace(out[len(out)-1]) == "" {
				skipBlank = true
			}
			continue
		}
		out = append(out, stripped)
	}
	return strings.Join(out, "")
}

// stripCommentsInLine takes the comments out of one line. in says whether
// the line starts inside a comment, still whether it ends inside one, and
// cut whether anything was taken out.
func stripCommentsInLine(line string, in bool) (out string, still, cut bool) {
	var b strings.Builder
	i := 0
	for i < len(line) {
		if in {
			end := strings.Index(line[i:], "-->")
			if end < 0 {
				return strings.TrimRight(b.String(), " \t"), true, true
			}
			i += end + len("-->")
			in, cut = false, true
			continue
		}
		switch {
		case line[i] == '\\' && i+1 < len(line):
			b.WriteString(line[i : i+2])
			i += 2
		case line[i] == '`':
			end := codeSpanEnd(line, i)
			b.WriteString(line[i:end])
			i = end
		case strings.HasPrefix(line[i:], "<!--"):
			in, cut = true, true
			i += len("<!--")
		default:
			b.WriteByte(line[i])
			i++
		}
	}
	s := b.String()
	if cut {
		// The spaces that stood before a comment at the end of the line.
		body := strings.TrimRight(s, "\r\n")
		s = strings.TrimRight(body, " \t") + s[len(body):]
	}
	return s, in, cut
}

// modAPI reads the mod API version from the sharing page: 'This Bururu
// speaks `"0.12"`'.
func modAPI(sharing string) string {
	m := regexp.MustCompile("speaks `\"(\\d+\\.\\d+)\"`").FindStringSubmatch(sharing)
	if m == nil {
		return ""
	}
	return m[1]
}

// Paths of someone's own PC have no place on a public site.
var privatePath = regexp.MustCompile(`(?i)[a-z]:\\users\\|/users/[a-z]|/home/[a-z]`)

func checkPrivate(name, text string) error {
	if loc := privatePath.FindStringIndex(text); loc != nil {
		line := strings.Count(text[:loc[0]], "\n") + 1
		return fmt.Errorf("%s:%d: a path of someone's PC (%q); take it out of the guide first", name, line, text[loc[0]:loc[1]])
	}
	return nil
}

// pageOrder is the sidebar's order: the start page, then the pages in the
// order its "Chapters" list links them, then the rest by name, then the
// pages from the player guide.
func pageOrder(pages map[string]string) []string {
	order := []string{"index"}
	seen := map[string]bool{"index": true}
	add := func(p string) {
		id := strings.TrimSuffix(p, ".md")
		if !seen[id] {
			seen[id] = true
			order = append(order, id)
		}
	}
	if index, ok := pages["index.md"]; ok {
		for _, t := range linkTargets(chaptersSection(index)) {
			p, _ := splitTarget(t)
			if _, ok := pages[p]; ok && p != "" {
				add(p)
			}
		}
	}
	var rest, player []string
	for p := range pages {
		if strings.HasPrefix(p, "player/") {
			player = append(player, p)
		} else {
			rest = append(rest, p)
		}
	}
	sort.Strings(rest)
	sort.Strings(player)
	for _, p := range append(rest, player...) {
		add(p)
	}
	return order
}

// chaptersSection is the text under the "## Chapters" heading, up to the
// next heading of that level or higher.
func chaptersSection(text string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) != "## Chapters" {
			continue
		}
		var b strings.Builder
		for _, l := range lines[i+1:] {
			if strings.HasPrefix(l, "# ") || strings.HasPrefix(l, "## ") {
				break
			}
			b.WriteString(l)
			b.WriteByte('\n')
		}
		return b.String()
	}
	return ""
}

type versionsFile struct {
	Current  string    `json:"current"`
	Versions []version `json:"versions"`
}

type version struct {
	ID  string `json:"id"`
	API string `json:"api,omitempty"`
}

func readVersions(file string) (versionsFile, error) {
	var vf versionsFile
	if b, err := os.ReadFile(file); err == nil {
		if err := json.Unmarshal(b, &vf); err != nil {
			return vf, fmt.Errorf("%s: %w", file, err)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return vf, err
	}
	return vf, nil
}

// checkAPI stops a sync that would change the mod API of a docs version
// the site already has. A docs version covers every patch release of the
// app's minor version, so when a patch moves the mod API, players on the
// earlier patch would read about mods their Bururu refuses.
func checkAPI(file, v, api string, force bool) error {
	vf, err := readVersions(file)
	if err != nil || api == "" || force {
		return err
	}
	for _, have := range vf.Versions {
		if have.ID == v && have.API != "" && have.API != api {
			return fmt.Errorf("the %s docs are for mod API %s, and this guide is for %s. "+
				"A patch release changed the mod API, so these docs would no longer fit the earlier %s releases. "+
				"Pass -force to replace them anyway", v, have.API, api, v)
		}
	}
	return nil
}

// addVersion records v in versions.json, newest first. The newest
// version is the current one.
func addVersion(file, v, api string) error {
	vf, err := readVersions(file)
	if err != nil {
		return err
	}
	found := false
	for i := range vf.Versions {
		if vf.Versions[i].ID == v {
			found = true
			if api != "" {
				vf.Versions[i].API = api
			}
		}
	}
	if !found {
		vf.Versions = append(vf.Versions, version{ID: v, API: api})
	}
	sort.Slice(vf.Versions, func(i, j int) bool { return newer(vf.Versions[i].ID, vf.Versions[j].ID) })
	if vf.Current == "" || newer(v, vf.Current) {
		vf.Current = v
	}
	return writeJSON(file, vf)
}

// newer says whether major.minor version a is newer than b.
func newer(a, b string) bool {
	amaj, amin, _ := strings.Cut(a, ".")
	bmaj, bmin, _ := strings.Cut(b, ".")
	x, _ := strconv.Atoi(amaj)
	y, _ := strconv.Atoi(bmaj)
	if x != y {
		return x > y
	}
	x, _ = strconv.Atoi(amin)
	y, _ = strconv.Atoi(bmin)
	return x > y
}

func writeJSON(file string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, append(b, '\n'), 0o644)
}
