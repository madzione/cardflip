package render

import (
	"html"
	"strings"
)

// MathJax formulas, \(…\) inline and \[…\] on their own line, become
// <cf-math display="inline|block"> elements holding the TeX, which the app
// typesets. Only balanced pairs outside tags count; an unbalanced delimiter
// stays text. Tags inside a formula are removed and entities decoded, since
// fields edited in a rich editor often contain <br> or &nbsp; in formulas.

// clozeColorMacro marks a revealed cloze inside math. RenderCard replaces it
// with \color{…} when the note type's CSS colours .cloze; otherwise the app
// colours it like other cloze highlights.
const clozeColorMacro = `\cfcloze`

// wrapMath replaces the formulas in rendered HTML with <cf-math> elements.
func wrapMath(src string) string {
	spans := mathSpans(src)
	if len(spans) == 0 {
		return src
	}
	var out strings.Builder
	last := 0
	for _, sp := range spans {
		display := "inline"
		if src[sp[0]+1] == '[' {
			display = "block"
		}
		tex, tags := mathTeX(src[sp[0]+2 : sp[1]-2])
		out.WriteString(src[last:sp[0]])
		out.WriteString(`<cf-math display="` + display + `">`)
		out.WriteString(html.EscapeString(tex))
		out.WriteString(`</cf-math>`)
		out.WriteString(tags)
		last = sp[1]
	}
	out.WriteString(src[last:])
	return out.String()
}

// mathOpener recognises \( and \[ and returns the matching closer.
func mathOpener(s string) (closer, display string, ok bool) {
	switch {
	case strings.HasPrefix(s, `\(`):
		return `\)`, "inline", true
	case strings.HasPrefix(s, `\[`):
		return `\]`, "block", true
	}
	return "", "", false
}

// findMathEnd returns the position of closer at or after pos, skipping tags,
// or -1. A formula never extends into a script or style element.
func findMathEnd(src string, pos int, closer string) int {
	for pos < len(src) {
		if src[pos] == '<' {
			if rawTextElement(src[pos:]) != "" {
				return -1
			}
			pos = skipMarkup(src, pos)
			continue
		}
		if strings.HasPrefix(src[pos:], closer) {
			return pos
		}
		pos++
	}
	return -1
}

// skipMarkup returns the position after the tag, comment or raw-text element
// (script, style) starting at src[i] == '<'. A '<' that starts no tag is
// skipped as text.
func skipMarkup(src string, i int) int {
	rest := src[i:]
	if strings.HasPrefix(rest, "<!--") {
		if end := strings.Index(rest[4:], "-->"); end >= 0 {
			return i + 4 + end + 3
		}
		return len(src)
	}
	if len(rest) < 2 || !(isLetter(rest[1]) || rest[1] == '/' || rest[1] == '!') {
		return i + 1
	}
	end := tagEnd(src, i)
	if name := rawTextElement(rest); name != "" {
		if close := strings.Index(strings.ToLower(src[end:]), "</"+name); close >= 0 {
			return tagEnd(src, end+close)
		}
		return len(src)
	}
	return end
}

// tagEnd returns the position after the '>' closing the tag at src[i],
// respecting quoted attribute values.
func tagEnd(src string, i int) int {
	var quote byte
	for j := i + 1; j < len(src); j++ {
		c := src[j]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			return j + 1
		}
	}
	return len(src)
}

// rawTextElement returns "script" or "style" when s opens such an element.
func rawTextElement(s string) string {
	for _, name := range []string{"script", "style"} {
		if hasTagPrefix(s, "<"+name) {
			return name
		}
	}
	return ""
}

// mathTeX turns the HTML between math delimiters into TeX: line breaks
// become spaces, other tags are removed, entities are decoded and \ce{} and
// \pu{} are translated. Tags that do not close inside the formula are
// returned, closing tags first, so the caller can keep the HTML balanced.
func mathTeX(s string) (tex, tags string) {
	var sb, closers strings.Builder
	var open []string // unclosed start tags
	for i := 0; i < len(s); {
		if s[i] != '<' {
			sb.WriteByte(s[i])
			i++
			continue
		}
		end := skipMarkup(s, i)
		tag := s[i:end]
		switch {
		case end == i+1:
			sb.WriteByte('<')
		case hasTagPrefix(tag, "<br"):
			sb.WriteByte(' ')
		case strings.HasPrefix(tag, "</"):
			if n := len(open); n > 0 && hasTagPrefix(open[n-1], "<"+tagName(tag[2:])) {
				open = open[:n-1]
			} else {
				closers.WriteString(tag)
			}
		case isLetter(tag[1]) && !strings.HasSuffix(tag, "/>") && !voidElements[strings.ToLower(tagName(tag[1:]))]:
			open = append(open, tag)
		}
		i = end
	}
	tex = strings.ReplaceAll(html.UnescapeString(sb.String()), "\u00a0", " ")
	return translateChem(tex), closers.String() + strings.Join(open, "")
}

// tagName returns the element name at the start of s.
func tagName(s string) string {
	end := 0
	for end < len(s) && (isLetter(s[end]) || isDigit(s[end]) || s[end] == '-') {
		end++
	}
	return s[:end]
}

var voidElements = map[string]bool{"br": true, "hr": true, "img": true, "wbr": true, "input": true, "col": true, "source": true, "area": true, "embed": true, "meta": true, "link": true, "base": true, "param": true, "track": true}

// clozeInMath writes a cloze deletion that lies inside a formula: on the
// question \text{[...]} (or the hint), on the answer the revealed TeX in a
// coloured group.
func clozeInMath(sb *strings.Builder, p clozePiece, ord int, question bool, math [][2]int) {
	if question {
		label := "..."
		if strings.TrimSpace(p.hint) != "" {
			label = p.hint
		}
		sb.WriteString(`\text{[` + label + `]}`)
		return
	}
	sb.WriteString(clozeColorMacro + "{")
	writeCloze(sb, p.children, ord, question, math)
	sb.WriteString("}")
}

// mathSpans returns the [start, end) byte ranges of the balanced formulas in
// a field's text, delimiters included.
func mathSpans(src string) [][2]int {
	var spans [][2]int
	for i := 0; i < len(src); {
		if src[i] == '<' {
			i = skipMarkup(src, i)
			continue
		}
		if closer, _, ok := mathOpener(src[i:]); ok {
			if end := findMathEnd(src, i+2, closer); end >= 0 {
				spans = append(spans, [2]int{i, end + 2})
				i = end + 2
				continue
			}
			i += 2
			continue
		}
		i++
	}
	return spans
}

// inSpan reports whether pos lies inside one of the spans.
func inSpan(spans [][2]int, pos int) bool {
	for _, sp := range spans {
		if pos >= sp[0] && pos < sp[1] {
			return true
		}
	}
	return false
}
