package main

import "strings"

// rewriteLinks calls fn for the target of every inline link and image in
// a Markdown text, outside code blocks and code spans. fn returns the new
// target, or keep false to turn the link into its plain text.
func rewriteLinks(text string, fn func(target string) (string, bool)) string {
	var out strings.Builder
	var fence codeFence
	for _, line := range strings.SplitAfter(text, "\n") {
		if !fence.inCode(line) {
			line = rewriteLine(line, fn)
		}
		out.WriteString(line)
	}
	return out.String()
}

// codeFence follows the fenced code blocks of a text, line by line.
type codeFence struct{ open string }

// inCode says whether line belongs to a fenced code block, its fences
// included.
func (c *codeFence) inCode(line string) bool {
	f := fenceOf(line)
	switch {
	case c.open == "" && f != "":
		c.open = f
		return true
	case c.open != "":
		// A closing fence is the same character, at least as many times,
		// and nothing after it.
		if f != "" && f[0] == c.open[0] && len(f) >= len(c.open) && strings.TrimSpace(line) == f {
			c.open = ""
		}
		return true
	}
	return false
}

// linkTargets lists the targets of the inline links in a Markdown text.
func linkTargets(text string) []string {
	var targets []string
	rewriteLinks(text, func(t string) (string, bool) {
		targets = append(targets, t)
		return t, true
	})
	return targets
}

// fenceOf returns the fence a line opens or closes (``` or ~~~, three or
// more), or "".
func fenceOf(line string) string {
	s := strings.TrimRight(line, "\r\n")
	trimmed := strings.TrimLeft(s, " ")
	if len(s)-len(trimmed) > 3 || len(trimmed) < 3 {
		return ""
	}
	c := trimmed[0]
	if c != '`' && c != '~' {
		return ""
	}
	n := 0
	for n < len(trimmed) && trimmed[n] == c {
		n++
	}
	if n < 3 {
		return ""
	}
	return trimmed[:n]
}

func rewriteLine(line string, fn func(string) (string, bool)) string {
	var out strings.Builder
	i := 0
	for i < len(line) {
		switch c := line[i]; {
		case c == '\\' && i+1 < len(line):
			out.WriteString(line[i : i+2])
			i += 2
		case c == '`':
			end := codeSpanEnd(line, i)
			out.WriteString(line[i:end])
			i = end
		case c == '[':
			textEnd, targetStart, targetEnd, ok := parseLink(line, i)
			if !ok {
				out.WriteByte(c)
				i++
				continue
			}
			target := line[targetStart:targetEnd]
			label := line[i+1 : textEnd]
			newTarget, keep := fn(target)
			if keep {
				out.WriteString("[")
				out.WriteString(rewriteLine(label, fn))
				out.WriteString("](")
				out.WriteString(newTarget)
				out.WriteString(line[targetEnd:closeParen(line, targetEnd)])
			} else {
				// The link's text in place of the link. For an image,
				// its alt text without the "!".
				if s := out.String(); strings.HasSuffix(s, "!") {
					out.Reset()
					out.WriteString(s[:len(s)-1])
				}
				out.WriteString(rewriteLine(label, fn))
			}
			i = closeParen(line, targetEnd)
		default:
			out.WriteByte(c)
			i++
		}
	}
	return out.String()
}

// codeSpanEnd returns where the code span that starts at i ends: after the
// closing run of as many backticks. A run with no closing one is text.
func codeSpanEnd(s string, i int) int {
	n := 0
	for i+n < len(s) && s[i+n] == '`' {
		n++
	}
	run := s[i : i+n]
	for j := i + n; j < len(s); {
		k := strings.Index(s[j:], run)
		if k < 0 {
			break
		}
		k += j
		m := 0
		for k+m < len(s) && s[k+m] == '`' {
			m++
		}
		if m == n {
			return k + n
		}
		j = k + m
	}
	return i + n
}

// parseLink reads "[text](target ...)" at i. It returns where the text
// ends (the "]"), and the start and end of the target.
func parseLink(s string, i int) (textEnd, targetStart, targetEnd int, ok bool) {
	depth := 0
	j := i
	for j < len(s) {
		switch s[j] {
		case '\\':
			j += 2
			continue
		case '`':
			j = codeSpanEnd(s, j)
			continue
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				textEnd = j
				goto text
			}
		case '\n':
			return 0, 0, 0, false
		}
		j++
	}
	return 0, 0, 0, false
text:
	if textEnd+1 >= len(s) || s[textEnd+1] != '(' {
		return 0, 0, 0, false
	}
	targetStart = textEnd + 2
	parens := 0
	for k := targetStart; k < len(s); k++ {
		switch s[k] {
		case '(':
			parens++
		case ')':
			if parens == 0 {
				return textEnd, targetStart, k, true
			}
			parens--
		case ' ', '\t', '\n':
			if parens == 0 {
				if closeParen(s, k) > k {
					return textEnd, targetStart, k, true
				}
				return 0, 0, 0, false
			}
		}
	}
	return 0, 0, 0, false
}

// closeParen returns the index after the ")" that closes a link whose
// target ends at i (a title may sit in between).
func closeParen(s string, i int) int {
	for k := i; k < len(s); k++ {
		if s[k] == ')' {
			return k + 1
		}
		if s[k] == '\n' {
			break
		}
	}
	return i
}
