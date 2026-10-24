package render

import (
	"sort"
	"strconv"
	"strings"
)

// A cloze field is text with deletions {{cN::answer}} or {{cN::answer::hint}}.
// A deletion may list several numbers, {{c1,2::answer}}, and belongs to
// each of their cards. Deletions may nest.

type clozePiece struct {
	text     string // plain text when ords is empty
	ords     []int
	children []clozePiece
	hint     string
	start    int // byte offset of a deletion's "{{" in the field text
}

// isText reports whether the piece is plain text rather than a deletion.
func (p clozePiece) isText() bool { return len(p.ords) == 0 }

// active reports whether the piece is a deletion for cloze number ord.
func (p clozePiece) active(ord int) bool {
	for _, o := range p.ords {
		if o == ord {
			return true
		}
	}
	return false
}

func parseCloze(s string) []clozePiece {
	pieces, _ := parseClozeRun(s, 0, false)
	return pieces
}

// parseClozeRun parses from pos until the end or, inside a deletion, its
// closing "}}". It returns the pieces and the position after the run.
func parseClozeRun(s string, pos int, inner bool) ([]clozePiece, int) {
	var out []clozePiece
	text := strings.Builder{}
	flush := func() {
		if text.Len() > 0 {
			out = append(out, clozePiece{text: text.String()})
			text.Reset()
		}
	}
	for pos < len(s) {
		if inner && strings.HasPrefix(s[pos:], "}}") {
			flush()
			return out, pos
		}
		if ords, n := clozeOpen(s[pos:]); n > 0 {
			children, end := parseClozeRun(s, pos+n, true)
			if end >= len(s) {
				// Unclosed deletion: keep it as literal text.
				text.WriteString(s[pos : pos+n])
				pos += n
				continue
			}
			flush()
			p := splitHint(ords, children)
			p.start = pos
			out = append(out, p)
			pos = end + 2
			continue
		}
		text.WriteByte(s[pos])
		pos++
	}
	flush()
	return out, pos
}

// clozeOpen recognises "{{cN::" or "{{cN,M,…::" (spaces allowed after the
// commas) and returns the numbers and the opener's length. Every number must
// be positive; an empty entry such as "c1,,2" is not an opener.
func clozeOpen(s string) ([]int, int) {
	if !strings.HasPrefix(s, "{{c") && !strings.HasPrefix(s, "{{C") {
		return nil, 0
	}
	var ords []int
	i := 3
	for {
		start := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		n, err := strconv.Atoi(s[start:i])
		if err != nil || n < 1 {
			return nil, 0
		}
		ords = append(ords, n)
		if strings.HasPrefix(s[i:], "::") {
			return ords, i + 2
		}
		if i >= len(s) || s[i] != ',' {
			return nil, 0
		}
		i++
		for i < len(s) && s[i] == ' ' {
			i++
		}
	}
}

// splitHint moves the text after the first top-level "::" into the hint.
func splitHint(ords []int, children []clozePiece) clozePiece {
	p := clozePiece{ords: ords}
	for i, c := range children {
		if !c.isText() {
			continue
		}
		if before, after, ok := strings.Cut(c.text, "::"); ok {
			p.children = append(append([]clozePiece{}, children[:i]...), clozePiece{text: before})
			var hint strings.Builder
			hint.WriteString(after)
			for _, rest := range children[i+1:] {
				hint.WriteString(plainCloze(rest))
			}
			p.hint = hint.String()
			return p
		}
	}
	p.children = children
	return p
}

// plainCloze renders pieces with every deletion revealed and unmarked.
func plainCloze(p clozePiece) string {
	if p.isText() {
		return p.text
	}
	var sb strings.Builder
	for _, c := range p.children {
		sb.WriteString(plainCloze(c))
	}
	return sb.String()
}

// renderCloze renders text for cloze number ord: on the question, deletion
// ord becomes [...] or [hint]; on the answer it is highlighted. Other
// deletions show their answer text, unhighlighted. Inside a formula the
// deletion is written as TeX (see clozeInMath).
func renderCloze(text string, ord int, question bool) string {
	var sb strings.Builder
	writeCloze(&sb, parseCloze(text), ord, question, mathSpans(text))
	return sb.String()
}

func writeCloze(sb *strings.Builder, pieces []clozePiece, ord int, question bool, math [][2]int) {
	for _, p := range pieces {
		switch {
		case p.isText():
			sb.WriteString(p.text)
		case p.active(ord) && inSpan(math, p.start):
			clozeInMath(sb, p, ord, question, math)
		case p.active(ord) && question:
			label := "..."
			if strings.TrimSpace(p.hint) != "" {
				label = p.hint
			}
			sb.WriteString(`<span class="cloze">[` + label + `]</span>`)
		case p.active(ord):
			sb.WriteString(`<span class="cloze">`)
			writeCloze(sb, p.children, ord, question, math)
			sb.WriteString(`</span>`)
		default:
			writeCloze(sb, p.children, ord, question, math)
		}
	}
}

// clozeAnswers returns the answers of cloze number ord's deletions in text,
// joined with ", ": what {{type:cloze:Field}} expects.
func clozeAnswers(text string, ord int) string {
	var answers []string
	var walk func([]clozePiece)
	walk = func(ps []clozePiece) {
		for _, p := range ps {
			switch {
			case p.active(ord):
				answers = append(answers, plainCloze(p))
			case !p.isText():
				walk(p.children)
			}
		}
	}
	walk(parseCloze(text))
	return strings.Join(answers, ", ")
}

// ClozeNumbers returns the distinct deletion numbers in the given texts,
// ascending. A cloze note gets one card per number.
func ClozeNumbers(texts ...string) []int {
	seen := map[int]bool{}
	var walk func([]clozePiece)
	walk = func(ps []clozePiece) {
		for _, p := range ps {
			for _, o := range p.ords {
				seen[o] = true
			}
			if !p.isText() {
				walk(p.children)
			}
		}
	}
	for _, t := range texts {
		walk(parseCloze(t))
	}
	out := make([]int, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// ClozeFields returns the names of the fields the template passes through
// the cloze filter, e.g. "Text" for {{cloze:Text}}.
func ClozeFields(tmpl string) []string { return clozeFields(tmpl) }

func clozeFields(tmpl string) []string {
	var out []string
	var walk func([]node)
	walk = func(nodes []node) {
		for _, n := range nodes {
			switch n := n.(type) {
			case fieldNode:
				for _, f := range n.filters {
					if f == "cloze" {
						out = append(out, n.name)
					}
				}
			case sectionNode:
				walk(n.body)
			}
		}
	}
	walk(parseTemplate(tmpl))
	return out
}

// StripCloze removes cloze markup from s, keeping each deletion's answer
// and dropping its hint.
func StripCloze(s string) string {
	if !strings.Contains(s, "{{") {
		return s
	}
	var sb strings.Builder
	for _, p := range parseCloze(s) {
		sb.WriteString(plainCloze(p))
	}
	return sb.String()
}
