// Package render turns a note and one of its note type's templates into the
// question and answer HTML of a card, using the subset of Anki's template
// language described in the card-rendering spec, and prepares that HTML for
// the app: sanitized, with the note type's CSS inlined.
package render

import (
	"cardflip/golang/internal/usermsg"
	"fmt"
	"html"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// node is one piece of a parsed template.
type node interface{ isNode() }

type textNode string

// fieldNode is {{filter2:filter1:Field}}; filters apply right to left.
type fieldNode struct {
	name    string
	filters []string
}

// sectionNode is {{#Field}}…{{/Field}} (or {{^Field}} when inverted).
type sectionNode struct {
	name     string
	inverted bool
	body     []node
}

// errorNode marks a template mistake, shown inline instead of failing.
type errorNode string

func (textNode) isNode()    {}
func (fieldNode) isNode()   {}
func (sectionNode) isNode() {}
func (errorNode) isNode()   {}

// parseTemplate parses tmpl. It never fails: mistakes such as an unclosed
// section become error nodes so that the card still renders.
func parseTemplate(tmpl string) []node {
	p := &parser{src: tmpl, crossed: map[string]int{}}
	nodes, _ := p.parse()
	return nodes
}

type parser struct {
	src string
	pos int
	// open holds the sections being parsed, outermost first.
	open []openSection
	// crossed counts the sections closed early by the closing tag of a
	// section around them; their own closing tag, when it follows, is
	// already reported.
	crossed map[string]int
}

// openSection is a section whose closing tag has not been read yet.
type openSection struct {
	mark byte // '#' or '^'
	name string
}

// parse reads nodes until the end or the closing tag of an open section,
// whose name it returns: the innermost section's own or, when sections
// cross, an outer one's.
func (p *parser) parse() ([]node, string) {
	var out []node
	for p.pos < len(p.src) {
		start := strings.Index(p.src[p.pos:], "{{")
		if start < 0 {
			out = append(out, textNode(p.src[p.pos:]))
			p.pos = len(p.src)
			break
		}
		if start > 0 {
			out = append(out, textNode(p.src[p.pos:p.pos+start]))
		}
		p.pos += start
		end := strings.Index(p.src[p.pos+2:], "}}")
		next := strings.Index(p.src[p.pos+2:], "{{")
		if end < 0 || next >= 0 && next < end {
			// {{ without }}: reported, and kept as text up to the next {{.
			stop := len(p.src)
			if next >= 0 {
				stop = p.pos + 2 + next
			}
			raw := p.src[p.pos:stop]
			out = append(out, errorNode(usermsg.Text("tagNotClosed", "%s has no closing }}", unclosedTag(raw))), textNode(raw))
			p.pos = stop
			continue
		}
		raw := p.src[p.pos : p.pos+2+end+2]
		tag := strings.TrimSpace(p.src[p.pos+2 : p.pos+2+end])
		p.pos += 2 + end + 2
		if tag != "" && (tag[0] == '#' || tag[0] == '^' || tag[0] == '/') {
			name := strings.TrimSpace(tag[1:])
			if name == "" {
				out = append(out, errorNode(usermsg.Text("tagNoFieldName", "%s has no field name", "{{"+tag[:1]+"}}")), textNode(raw))
				continue
			}
			if tag[0] == '/' {
				if p.isOpen(name) {
					return out, name
				}
				if p.crossed[name] > 0 {
					p.crossed[name]--
					continue
				}
				out = append(out, errorNode(usermsg.Text("sectionNotOpened", "%s has no matching %s", "{{/"+name+"}}", "{{#"+name+"}}")))
				continue
			}
			p.open = append(p.open, openSection{mark: tag[0], name: name})
			body, closing := p.parse()
			p.open = p.open[:len(p.open)-1]
			out = append(out, sectionNode{name: name, inverted: tag[0] == '^', body: body})
			switch {
			case closing == name:
			case closing == "":
				out = append(out, errorNode(usermsg.Text("sectionNotClosed", "%s is not closed", "{{"+tag[:1]+name+"}}")))
			default:
				// The closing tag of a section around this one.
				outer := p.open[slices.IndexFunc(p.open, func(o openSection) bool { return o.name == closing })]
				out = append(out, errorNode(usermsg.Text("sectionsCrossed", "%s closes %s before %s is closed",
					"{{/"+closing+"}}", "{{"+string(outer.mark)+closing+"}}", "{{"+tag[:1]+name+"}}")))
				p.crossed[name]++
				return out, closing
			}
			continue
		}
		parts := strings.Split(tag, ":")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) == 2 && parts[0] == "tts-voices" && parts[1] == "" {
			out = append(out, fieldNode{filters: []string{"tts-voices"}})
			continue
		}
		if parts[len(parts)-1] == "" {
			out = append(out, errorNode(usermsg.Text("tagNoFieldName", "%s has no field name", "{{"+tag+"}}")), textNode(raw))
			continue
		}
		out = append(out, fieldNode{name: parts[len(parts)-1], filters: parts[:len(parts)-1]})
	}
	return out, ""
}

// isOpen reports whether a section named name is being parsed.
func (p *parser) isOpen(name string) bool {
	return slices.ContainsFunc(p.open, func(o openSection) bool { return o.name == name })
}

// unclosedTag shortens a {{ without }} for messages: up to the first space,
// line break or HTML tag, and at most 30 characters.
func unclosedTag(raw string) string {
	rest := strings.TrimLeft(raw[2:], " \t")
	if i := strings.IndexAny(rest, " \t\r\n<"); i >= 0 {
		rest = rest[:i]
	}
	if r := []rune(rest); len(r) > 30 {
		rest = string(r[:30]) + "…"
	}
	return "{{" + rest
}

// context holds what a template is rendered against.
type context struct {
	fields   map[string]string
	special  map[string]string
	question bool // rendering the question side
	clozeOrd int  // cloze number being studied (1-based), 0 for standard
	// usedNonEmpty is set when a non-empty field is substituted; a card is
	// only generated when its question uses at least one.
	usedNonEmpty bool
	// typed is what the user typed for {{type:…}}, nil when the answer is
	// shown without typing (previews). typeUsed is set once the first
	// type: filter is rendered; later ones render nothing.
	typed    *string
	typeUsed bool
	// typeExpected and typeNC are the expected text and nc: modifier of the
	// first type: filter on the question, for an answer that only shows it
	// through {{FrontSide}}.
	typeExpected string
	typeNC       bool
}

func (c *context) lookup(name string) (string, bool) {
	if v, ok := c.fields[name]; ok {
		return v, true
	}
	if v, ok := c.special[name]; ok {
		return v, true
	}
	return "", false
}

func (c *context) render(nodes []node, sb *strings.Builder) {
	for _, n := range nodes {
		switch n := n.(type) {
		case textNode:
			sb.WriteString(string(n))
		case errorNode:
			sb.WriteString(errorMarker(usermsg.Plain(string(n))))
		case sectionNode:
			if num, ok := c.clozeCondition(n.name); ok {
				if (num == c.clozeOrd) != n.inverted {
					c.render(n.body, sb)
				}
				continue
			}
			v, ok := c.lookup(n.name)
			if !ok {
				sb.WriteString(errorMarker(fmt.Sprintf("unknown field %s", n.name)))
			}
			if nonEmpty(v) != n.inverted {
				c.render(n.body, sb)
			}
		case fieldNode:
			sb.WriteString(c.field(n))
		}
	}
}

// clozeCondition reports whether a section named name is a cloze-number
// condition ({{#c2}}…{{/c2}} shows only on cloze card 2) and returns the
// number. That needs a cloze card and no field of that name, which wins.
func (c *context) clozeCondition(name string) (int, bool) {
	if c.clozeOrd == 0 {
		return 0, false
	}
	if _, isField := c.fields[name]; isField {
		return 0, false
	}
	return clozeNumber(name)
}

// clozeNumber returns N for a name of the form cN with N ≥ 1.
func clozeNumber(name string) (int, bool) {
	if len(name) < 2 || name[0] != 'c' {
		return 0, false
	}
	num, err := strconv.Atoi(name[1:])
	if err != nil || num < 1 || "c"+strconv.Itoa(num) != name {
		return 0, false
	}
	return num, true
}

func (c *context) field(n fieldNode) string {
	if n.name == "" && slices.Equal(n.filters, []string{"tts-voices"}) {
		return ttsVoicesPlaceholder
	}
	v, ok := c.lookup(n.name)
	if !ok {
		return errorMarker(fmt.Sprintf("unknown field %s", n.name))
	}
	if _, isField := c.fields[n.name]; isField && nonEmpty(v) {
		c.usedNonEmpty = true
	}
	if len(n.filters) > 0 && n.filters[0] == "type" {
		return c.typeAnswer(n.filters[1:], v)
	}
	for i := len(n.filters) - 1; i >= 0; i-- {
		if n.filters[i] == "cloze" && i > 0 && isTTSFilter(n.filters[i-1]) && c.clozeOrd > 0 {
			// Spoken, the active deletion is read as "blank".
			v = clozeForSpeech(v, c.clozeOrd, c.question)
			continue
		}
		v = c.filter(n.filters[i], n.name, v)
	}
	return v
}

func (c *context) filter(name, field, v string) string {
	switch {
	case name == "text":
		return html.EscapeString(StripHTML(v))
	case name == "hint":
		if !nonEmpty(v) {
			return ""
		}
		return fmt.Sprintf(`<details class="hint"><summary>%s</summary>%s</details>`, html.EscapeString(field), v)
	case name == "cloze":
		if c.clozeOrd == 0 {
			return v
		}
		if hasOcclusions(v) {
			return occlusionElement(v, c.clozeOrd, c.question)
		}
		return renderCloze(v, c.clozeOrd, c.question)
	case name == "furigana":
		return furigana(v)
	case name == "kana":
		return kana(v)
	case name == "kanji":
		return kanji(v)
	case name == "cloze-only":
		if c.clozeOrd == 0 {
			return v
		}
		return clozeAnswers(v, c.clozeOrd)
	case isTTSFilter(name):
		return ttsFilter(strings.TrimPrefix(name, "tts"), v)
	}
	// Unknown filters are ignored.
	return v
}

// typeAnsPlaceholder marks where the app shows the answer input.
const typeAnsPlaceholder = "<cf-typeans></cf-typeans>"

// typeAnswer renders {{type:Field}}, {{type:nc:Field}} and
// {{type:cloze:Field}}, with mods the filters after type:. Only the first
// one on a side is active. The question gets the input placeholder; the
// answer compares the typed text with the expected one, or shows the
// expected text when nothing was typed in (previews).
func (c *context) typeAnswer(mods []string, v string) string {
	if c.typeUsed {
		return ""
	}
	c.typeUsed = true
	nc := false
	for _, m := range mods {
		switch m {
		case "cloze":
			if c.clozeOrd > 0 {
				v = clozeAnswers(v, c.clozeOrd)
			}
		case "nc":
			nc = true
		}
	}
	if c.question {
		c.typeExpected, c.typeNC = v, nc
		return typeAnsPlaceholder
	}
	return c.typeComparison(v, nc)
}

// typeComparison compares the typed text with expected, or returns expected
// when nothing was typed in (previews). As in Anki, the comparison is a
// block of its own: text after it would otherwise share its last line, and
// that text's taller font pushes the line down from the arrow above.
func (c *context) typeComparison(expected string, nc bool) string {
	if c.typed == nil {
		return expected
	}
	return "<div>" + CompareTyped(expected, *c.typed, nc) + "</div>"
}

// errorMarker is the inline marker shown for a template problem.
func errorMarker(msg string) string {
	return `<span class="cf-error" style="color:#d32f2f;font-weight:bold">⚠ ` + html.EscapeString(msg) + `</span>`
}

// nonEmpty reports whether field content has text after stripping HTML and
// whitespace. Images count as content.
func nonEmpty(v string) bool {
	if strings.Contains(strings.ToLower(v), "<img") {
		return true
	}
	return strings.TrimSpace(strings.ReplaceAll(StripHTML(v), "&nbsp;", " ")) != ""
}

var answerLine = regexp.MustCompile(`(?i)<hr[^>]*\bid\s*=\s*["']?answer\b[^>]*>`)

// AnswerPart returns what a rendered answer shows below its <hr id=answer>
// line, leaving out the question that {{FrontSide}} repeats above it, or the
// whole answer when it has no such line.
func AnswerPart(answer string) string {
	if loc := answerLine.FindStringIndex(answer); loc != nil {
		return answer[loc[1]:]
	}
	return answer
}

// StripHTML removes tags, comments and script/style content, and decodes
// entities that matter for plain text (&nbsp; becomes a space).
func StripHTML(s string) string {
	if !strings.ContainsAny(s, "<&") {
		return s
	}
	var sb strings.Builder
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '<' && strings.HasPrefix(s[i:], "<!--"):
			end := strings.Index(s[i+4:], "-->")
			if end < 0 {
				return sb.String()
			}
			i += 4 + end + 3
		case c == '<' && (hasTagPrefix(s[i:], "<script") || hasTagPrefix(s[i:], "<style")):
			name := "</script"
			if hasTagPrefix(s[i:], "<style") {
				name = "</style"
			}
			end := strings.Index(strings.ToLower(s[i:]), name)
			if end < 0 {
				return sb.String()
			}
			i += end
			gt := strings.IndexByte(s[i:], '>')
			if gt < 0 {
				return sb.String()
			}
			i += gt + 1
		case c == '<' && i+1 < len(s) && (isLetter(s[i+1]) || s[i+1] == '/' || s[i+1] == '!'):
			gt := strings.IndexByte(s[i:], '>')
			if gt < 0 {
				return sb.String()
			}
			tag := strings.ToLower(s[i+1 : i+gt])
			if strings.HasPrefix(tag, "br") || strings.HasPrefix(tag, "div") || strings.HasPrefix(tag, "/div") ||
				strings.HasPrefix(tag, "p") || strings.HasPrefix(tag, "/p") || strings.HasPrefix(tag, "li") {
				sb.WriteByte(' ')
			}
			i += gt + 1
		case c == '&':
			semi := strings.IndexByte(s[i:], ';')
			if semi > 0 && semi < 12 {
				ent := s[i : i+semi+1]
				if ent == "&nbsp;" {
					sb.WriteByte(' ')
				} else {
					sb.WriteString(html.UnescapeString(ent))
				}
				i += semi + 1
			} else {
				sb.WriteByte(c)
				i++
			}
		default:
			sb.WriteByte(c)
			i++
		}
	}
	return sb.String()
}

func hasTagPrefix(s, tag string) bool {
	if len(s) < len(tag) || !strings.EqualFold(s[:len(tag)], tag) {
		return false
	}
	return len(s) == len(tag) || !isLetter(s[len(tag)])
}

func isLetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
