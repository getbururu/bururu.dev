// Command serve shows the built site on this PC the way Cloudflare Pages
// serves it, close enough to check pages before a deploy: clean URLs
// (/0.14/en/rules/ for 0.14/en/rules/index.html, /x for x.html), the
// nearest 404.html up the tree for missing pages, and the rules in
// _headers and _redirects. It reads both files on every request, so a new
// build shows without a restart.
//
// From the bururu.dev folder, after go run ./tools/build:
//
//	go run ./tools/serve
//	go run ./tools/serve -addr 127.0.0.1:8080 -dir public
package main

import (
	"bufio"
	"errors"
	"flag"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	dir := flag.String("dir", "public", "the folder to serve")
	addr := flag.String("addr", "127.0.0.1:8788", "the address to listen on")
	flag.Parse()

	if _, err := os.Stat(*dir); err != nil {
		log.Fatal(err)
	}
	log.Printf("serving %s on http://%s/", *dir, *addr)
	log.Fatal(http.ListenAndServe(*addr, &site{dir: *dir, files: os.DirFS(*dir)}))
}

type site struct {
	dir   string
	files fs.FS
}

// rule is one entry of _headers or _redirects.
type rule struct {
	match   *regexp.Regexp
	names   []string    // the placeholders in match, in order; "splat" for *
	headers http.Header // _headers: the headers to set
	detach  []string    // _headers: "! Name" lines
	to      string      // _redirects: the target
	status  int         // _redirects: the status code
}

// Pages never serves its own config files.
var hidden = map[string]bool{"_headers": true, "_redirects": true, "_routes.json": true, "_worker.js": true}

func (s *site) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	urlPath := r.URL.Path

	for _, rd := range s.rules("_redirects") {
		if to, ok := rd.apply(urlPath, rd.to); ok {
			http.Redirect(w, r, to, rd.status)
			return
		}
	}

	file, redirect := s.resolve(urlPath)
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusPermanentRedirect)
		return
	}
	status := http.StatusOK
	if file == "" {
		log.Printf("404 %s", urlPath)
		status, file = http.StatusNotFound, s.notFound(urlPath)
	}
	s.setHeaders(w, r.Host, urlPath)
	s.send(w, r, file, status)
}

// resolve finds the file for a URL path the way Pages does. It returns
// the file (relative to the folder), or a clean URL to redirect to, or
// neither when nothing matches.
func (s *site) resolve(urlPath string) (file, redirect string) {
	clean := path.Clean("/" + urlPath)
	name := strings.TrimPrefix(clean, "/")
	if hidden[path.Base(name)] {
		return "", ""
	}
	switch {
	case strings.HasSuffix(clean, "/index.html"):
		if s.isFile(name) {
			return "", strings.TrimSuffix(clean, "index.html")
		}
	case strings.HasSuffix(clean, ".html"):
		if s.isFile(name) {
			return "", strings.TrimSuffix(clean, ".html")
		}
	case strings.HasSuffix(urlPath, "/"):
		if s.isFile(path.Join(name, "index.html")) {
			return path.Join(name, "index.html"), ""
		}
		if name != "" && s.isFile(name+".html") {
			return "", clean
		}
	default:
		if s.isFile(name) {
			return name, ""
		}
		if s.isFile(name + ".html") {
			return name + ".html", ""
		}
		if s.isFile(path.Join(name, "index.html")) {
			return "", clean + "/"
		}
	}
	return "", ""
}

// notFound is the 404.html Pages shows for a missing path: the one in the
// same folder, or else the nearest one up the tree.
func (s *site) notFound(urlPath string) string {
	dir := path.Dir(path.Clean("/" + urlPath))
	if strings.HasSuffix(urlPath, "/") {
		dir = path.Clean("/" + urlPath)
	}
	for {
		name := strings.TrimPrefix(path.Join(dir, "404.html"), "/")
		if s.isFile(name) || dir == "/" {
			return name
		}
		dir = path.Dir(dir)
	}
}

func (s *site) isFile(name string) bool {
	if name == "" {
		name = "."
	}
	if !fs.ValidPath(name) {
		return false
	}
	st, err := fs.Stat(s.files, name)
	return err == nil && st.Mode().IsRegular()
}

func (s *site) setHeaders(w http.ResponseWriter, host, urlPath string) {
	var detach []string
	for _, h := range s.rules("_headers") {
		if _, ok := h.apply(urlPath, ""); !ok {
			if _, ok := h.apply("https://"+host+urlPath, ""); !ok {
				continue
			}
		}
		for name, values := range h.headers {
			for _, v := range values {
				if old := w.Header().Get(name); old != "" {
					v = old + ", " + v
				}
				w.Header().Set(name, v)
			}
		}
		detach = append(detach, h.detach...)
	}
	for _, name := range detach {
		w.Header().Del(name)
	}
}

// Content types by extension; Windows' registry is not always right.
var types = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "text/javascript; charset=utf-8",
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".webp": "image/webp",
	".ico":  "image/x-icon",
	".txt":  "text/plain; charset=utf-8",
	".xml":  "application/xml",
	".json": "application/json",
}

func (s *site) send(w http.ResponseWriter, r *http.Request, name string, status int) {
	f, err := s.files.Open(name)
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	defer f.Close()
	ext := strings.ToLower(path.Ext(name))
	ctype := types[ext]
	if ctype == "" {
		ctype = mime.TypeByExtension(ext)
	}
	if ctype != "" {
		w.Header().Set("Content-Type", ctype)
	}
	if status != http.StatusOK {
		w.WriteHeader(status)
		if r.Method == http.MethodGet {
			io.Copy(w, f)
		}
		return
	}
	st, err := f.Stat()
	rs, ok := f.(io.ReadSeeker)
	if err != nil || !ok {
		http.Error(w, "cannot read "+name, http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, name, st.ModTime(), rs)
}

// rules reads _headers or _redirects from the folder. A missing file
// gives no rules.
func (s *site) rules(file string) []rule {
	f, err := s.files.Open(file)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			log.Printf("%s: %v", file, err)
		}
		return nil
	}
	defer f.Close()

	var out []rule
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		text := strings.TrimSpace(line)
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		indented := line[0] == ' ' || line[0] == '\t'
		if file == "_headers" && indented {
			if len(out) == 0 {
				log.Printf("%s:%d: header before any path", file, n)
				continue
			}
			h := &out[len(out)-1]
			if name, ok := strings.CutPrefix(text, "!"); ok {
				h.detach = append(h.detach, strings.TrimSpace(name))
			} else if name, value, ok := strings.Cut(text, ":"); ok {
				h.headers.Add(strings.TrimSpace(name), strings.TrimSpace(value))
			} else {
				log.Printf("%s:%d: not a header: %s", file, n, text)
			}
			continue
		}
		fields := strings.Fields(text)
		re, names := pattern(fields[0])
		r := rule{match: re, names: names, headers: http.Header{}}
		if file == "_redirects" {
			if len(fields) < 2 {
				log.Printf("%s:%d: no target", file, n)
				continue
			}
			r.to, r.status = fields[1], http.StatusFound
			if len(fields) > 2 {
				r.status, _ = strconv.Atoi(fields[2])
			}
			if r.status < 300 || r.status > 399 {
				log.Printf("%s:%d: only redirects (3xx) work here, not %s", file, n, fields[2])
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

var placeholder = regexp.MustCompile(`^:[A-Za-z][A-Za-z0-9_]*`)

// pattern turns a Pages path pattern into a regexp: * matches anything
// (once, as :splat), :name matches up to the next / (or the next . in
// the host of an https:// pattern).
func pattern(p string) (*regexp.Regexp, []string) {
	hostEnd := 0
	if rest, ok := strings.CutPrefix(p, "https://"); ok {
		hostEnd = len(p)
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			hostEnd = len("https://") + i
		}
	}
	var names []string
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(p); {
		if p[i] == '*' {
			b.WriteString("(.*)")
			names = append(names, "splat")
			i++
			continue
		}
		if m := placeholder.FindString(p[i:]); m != "" {
			if i < hostEnd {
				b.WriteString(`([^.]+)`)
			} else {
				b.WriteString(`([^/]+)`)
			}
			names = append(names, m[1:])
			i += len(m)
			continue
		}
		b.WriteString(regexp.QuoteMeta(p[i : i+1]))
		i++
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String()), names
}

// apply matches a path against the rule and fills :splat and the named
// placeholders into target.
func (r rule) apply(p, target string) (string, bool) {
	m := r.match.FindStringSubmatch(p)
	if m == nil {
		return "", false
	}
	for i, name := range r.names {
		target = strings.ReplaceAll(target, ":"+name, m[i+1])
	}
	return target, true
}
