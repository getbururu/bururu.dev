package main

import (
	"bytes"
	"fmt"
	"html"
	"path"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// newMarkdown is GitHub's Markdown as far as these docs use it: tables,
// strikethrough and {#id} after a heading. Raw HTML in a page is not
// passed on, so a page cannot add scripts.
func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.Table, extension.Strikethrough),
		goldmark.WithParserOptions(parser.WithAttribute()),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(util.Prioritized(nodeRenderer{}, 100)),
		),
	)
}

// parsePage parses a page and gives its headings their ids. A translation
// (en is its English page) takes the English ids when it has the same
// headings at the same levels, so links into it keep working.
func parsePage(md goldmark.Markdown, id, lang, file string, src []byte, en *page) (*page, error) {
	p := &page{id: id, lang: lang, file: file, src: src, ids: map[string]bool{}}
	p.doc = md.Parser().Parse(text.NewReader(src))

	var nodes []*ast.Heading
	err := ast.Walk(p.doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Heading:
			nodes = append(nodes, n)
		case *ast.Link:
			p.links = append(p.links, n)
			p.dests = append(p.dests, append([]byte(nil), n.Destination...))
		case *ast.Image:
			if !isExternal(string(n.Destination)) {
				return ast.WalkStop, fmt.Errorf("%s: the picture %s: pictures must come from another site for now", file, n.Destination)
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	sameShape := en != nil && len(en.heads) == len(nodes)
	if sameShape {
		for i, n := range nodes {
			if en.heads[i].level != n.Level {
				sameShape = false
				break
			}
		}
	}
	counts := map[string]int{}
	for i, n := range nodes {
		h := heading{level: n.Level, text: plainText(n, src)}
		if v, ok := n.AttributeString("id"); ok {
			h.id = fmt.Sprint(stringOf(v))
		} else if sameShape {
			h.id = en.heads[i].id
		} else {
			h.id = slug(h.text)
		}
		if h.id == "" {
			h.id = "section"
		}
		if p.ids[h.id] {
			// GitHub's way: the second "Example" is example-1.
			for {
				counts[h.id]++
				next := h.id + "-" + strconv.Itoa(counts[h.id])
				if !p.ids[next] {
					h.id = next
					break
				}
			}
		}
		p.ids[h.id] = true
		n.SetAttributeString("id", []byte(h.id))
		p.heads = append(p.heads, h)
		if n.Level == 1 && p.title == "" {
			p.title = h.text
		}
	}
	if p.title == "" {
		return nil, fmt.Errorf("%s: the page needs a title: a first line such as \"# Rules\"", file)
	}
	return p, nil
}

func stringOf(v any) string {
	switch v := v.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	}
	return fmt.Sprint(v)
}

// slug makes a heading's id the way GitHub does: lower case, spaces to
// "-", and everything but letters, digits, "-" and "_" left out.
func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r == ' ':
			b.WriteByte('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// plainText is a node's text without Markdown: what a heading says, what
// search finds.
func plainText(n ast.Node, src []byte) string {
	var b strings.Builder
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		switch n := n.(type) {
		case *ast.Text:
			b.Write(n.Segment.Value(src))
			if n.SoftLineBreak() || n.HardLineBreak() {
				b.WriteByte(' ')
			}
			return
		case *ast.String:
			b.Write(n.Value)
			return
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				seg := lines.At(i)
				b.Write(seg.Value(src))
			}
			b.WriteByte(' ')
			return
		case *ast.RawHTML, *ast.HTMLBlock:
			return
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			walk(c)
			if c.Type() == ast.TypeBlock {
				b.WriteByte(' ')
			}
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

var schemeRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:`)

func isExternal(dest string) bool {
	return schemeRe.MatchString(dest) || strings.HasPrefix(dest, "//")
}

// resolveLinks sets every link of p to its address on the site, for the
// page as it shows in lang. It returns what does not resolve.
func (v *docVersion) resolveLinks(p *page, lang string) []string {
	var errs []string
	for i, l := range p.links {
		dest := string(p.dests[i])
		to, err := v.linkTarget(p, lang, dest)
		if err != "" {
			errs = append(errs, fmt.Sprintf("%s: the link to %q %s", p.file, dest, err))
			to = dest
		}
		l.Destination = []byte(to)
	}
	return errs
}

// linkTarget turns a link as written in a page into its address on the
// site: rules.md#event-rules becomes /0.14/fr/rules/#event-rules on a
// French page. err says why a link does not work.
func (v *docVersion) linkTarget(p *page, lang, dest string) (to, err string) {
	if dest == "" || isExternal(dest) {
		return dest, ""
	}
	file, frag, _ := strings.Cut(dest, "#")
	shown, _ := v.show(lang, p.id)
	if file == "" {
		if frag != "" && !shown.ids[frag] {
			return "", "goes to a heading this page does not have"
		}
		return dest, ""
	}
	if path.Ext(file) != ".md" {
		return "", "goes to a file the site does not carry; link to .md pages only"
	}
	id := strings.TrimSuffix(path.Clean(path.Join(path.Dir(p.id), file)), ".md")
	if !v.has(id) {
		return "", "goes to a page that does not exist"
	}
	to = pageURL(v.id, lang, id)
	if frag != "" {
		target, translated := v.show(lang, id)
		if !target.ids[frag] {
			if translated && lang != "en" {
				return "", fmt.Sprintf("goes to a heading %s does not have: give that heading the English id, as in \"## Heading {#%s}\"", target.file, frag)
			}
			return "", "goes to a heading that page does not have"
		}
		to += "#" + frag
	}
	return to, ""
}

// render turns a parsed page into HTML, for the language its links were
// last resolved for.
func render(md goldmark.Markdown, p *page) (string, error) {
	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, p.src, p.doc); err != nil {
		return "", fmt.Errorf("%s: %w", p.file, err)
	}
	out := buf.String()
	// Wide tables scroll on their own, not the page.
	out = strings.ReplaceAll(out, "<table>", `<div class="table" tabindex="0"><table>`)
	out = strings.ReplaceAll(out, "</table>", "</table></div>")
	return out, nil
}

// nodeRenderer draws headings with a link to themselves, and code blocks
// with their syntax in colour.
type nodeRenderer struct{}

func (nodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, renderHeading)
	reg.Register(ast.KindFencedCodeBlock, renderCode)
	reg.Register(ast.KindCodeBlock, renderCode)
}

func renderHeading(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	id, _ := n.AttributeString("id")
	if entering {
		fmt.Fprintf(w, "<h%d id=\"%s\">", n.Level, html.EscapeString(stringOf(id)))
		return ast.WalkContinue, nil
	}
	if n.Level > 1 {
		fmt.Fprintf(w, "<a class=\"anchor\" href=\"#%s\" aria-hidden=\"true\" tabindex=\"-1\">#</a>", html.EscapeString(stringOf(id)))
	}
	fmt.Fprintf(w, "</h%d>\n", n.Level)
	return ast.WalkContinue, nil
}

// Code block languages the guide uses, by the name chroma knows them.
var lexerNames = map[string]string{"jsonc": "json", "json5": "json", "ps": "powershell", "ps1": "powershell", "sh": "bash", "shell": "bash"}

func renderCode(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	var code strings.Builder
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		code.Write(seg.Value(src))
	}
	lang := ""
	if f, ok := node.(*ast.FencedCodeBlock); ok {
		lang = strings.ToLower(string(f.Language(src)))
	}
	w.WriteString(`<pre class="code" tabindex="0" dir="ltr"><code`)
	if lang != "" {
		fmt.Fprintf(w, ` class="language-%s"`, html.EscapeString(lang))
	}
	w.WriteString(">")
	w.WriteString(highlight(lang, code.String()))
	w.WriteString("</code></pre>\n")
	return ast.WalkSkipChildren, nil
}

// highlight marks up code for its language with chroma's short class
// names (k for a keyword, s2 for a string); site.css colours them.
// Code in a language chroma does not know stays plain.
func highlight(lang, code string) string {
	key := lang + "\n" + code
	if out, ok := highlighted[key]; ok {
		return out
	}
	out := highlightCode(lang, code)
	highlighted[key] = out
	return out
}

// highlighted keeps what highlight made: each page is made once per
// language, mostly with the same code blocks.
var highlighted = map[string]string{}

func highlightCode(lang, code string) string {
	if name, ok := lexerNames[lang]; ok {
		lang = name
	}
	var lexer chroma.Lexer
	if lang != "" {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		return html.EscapeString(code)
	}
	it, err := chroma.Coalesce(lexer).Tokenise(nil, code)
	if err != nil {
		return html.EscapeString(code)
	}
	var b strings.Builder
	for _, tok := range it.Tokens() {
		class := chroma.StandardTypes[tok.Type]
		value := html.EscapeString(tok.Value)
		if class == "" || tok.Type == chroma.Text || tok.Type == chroma.TextWhitespace {
			b.WriteString(value)
			continue
		}
		b.WriteString(`<span class="` + class + `">` + value + `</span>`)
	}
	return b.String()
}
