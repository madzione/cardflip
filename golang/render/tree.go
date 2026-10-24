package render

import (
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/html"
)

// The note type's CSS is matched against an element tree of a card side, so
// that selectors can look at ancestors and siblings. The tree is only used
// for matching: the side is written back token by token, as it came, with
// the styles that apply added to its start tags.

// element is an element of a card side, or one of the elements above it.
type element struct {
	tag      string
	id       string
	classes  []string
	attrs    []attr
	parent   *element
	children []*element
	index    int // among the parent's children
	// ancestors is a Bloom filter of the ancestors' elements, ids and
	// classes, to rule out most selectors without walking up the tree.
	ancestors uint64
}

func newElement(tag string, attrs []attr, parent *element) *element {
	e := &element{tag: tag, attrs: attrs, parent: parent}
	for _, a := range attrs {
		switch a.key {
		case "id":
			e.id = a.val
		case "class":
			for _, c := range strings.Fields(a.val) {
				if !contains(e.classes, c) {
					e.classes = append(e.classes, c)
				}
			}
		}
	}
	if parent != nil {
		e.index = len(parent.children)
		parent.children = append(parent.children, e)
		e.ancestors = parent.ancestors | parent.names()
	}
	return e
}

// names returns the Bloom filter bits of e's element, id and classes.
func (e *element) names() uint64 {
	m := bloomBit('e', e.tag)
	if e.id != "" {
		m |= bloomBit('#', e.id)
	}
	for _, c := range e.classes {
		m |= bloomBit('.', c)
	}
	return m
}

// bloomBit returns the Bloom filter bit of an element (kind 'e'), id ('#')
// or class ('.') name.
func bloomBit(kind byte, name string) uint64 {
	h := uint32(2166136261) ^ uint32(kind) // FNV-1a
	for i := 0; i < len(name); i++ {
		h = (h ^ uint32(name[i])) * 16777619
	}
	return 1 << (h % 64)
}

// prev returns the element right before e among its siblings, or nil.
func (e *element) prev() *element {
	if e.parent == nil || e.index == 0 {
		return nil
	}
	return e.parent.children[e.index-1]
}

func (e *element) attr(key string) (string, bool) {
	for _, a := range e.attrs {
		if a.key == key {
			return a.val, true
		}
	}
	return "", false
}

// cardRoot returns the element a card side's content sits in: the card's
// root, matched as Anki's <body> with the root classes, below an <html>
// that has them too, so that ".mobile .card" and ".nightMode .card" match
// it like ".card.mobile".
func cardRoot(classes []string) *element {
	a := []attr{{"class", strings.Join(classes, " ")}}
	return newElement("body", a, newElement("html", a, nil))
}

// fragment is a card side split into the tokens it is written back from.
type fragment struct {
	items []item
	// elements in document order.
	elements []*element
}

// item is a start tag, or text written as it is when el is nil.
type item struct {
	text string
	el   *element
	void bool // written as <tag />
}

// closesP holds the start tags that end an open <p>, as in browsers.
var closesP = map[string]bool{
	"address": true, "article": true, "aside": true, "blockquote": true, "details": true, "dialog": true,
	"div": true, "dl": true, "dd": true, "dt": true, "fieldset": true, "figcaption": true, "figure": true,
	"footer": true, "form": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"header": true, "hgroup": true, "hr": true, "li": true, "main": true, "menu": true, "nav": true,
	"ol": true, "p": true, "pre": true, "section": true, "table": true, "ul": true,
}

// impliedEnds lists, for start tags that end open elements of their own
// kind, which elements they end and which ones they do not look past.
var impliedEnds = map[string][2][]string{
	"li":     {{"li"}, {"ul", "ol"}},
	"dd":     {{"dd", "dt"}, {"dl"}},
	"dt":     {{"dd", "dt"}, {"dl"}},
	"tr":     {{"tr"}, {"table", "thead", "tbody", "tfoot"}},
	"td":     {{"td", "th"}, {"tr", "table"}},
	"th":     {{"td", "th"}, {"tr", "table"}},
	"option": {{"option"}, {"select", "datalist"}},
}

var pScope = []string{"button", "table", "td", "th", "caption", "object", "marquee", "applet"}

// parseFragment builds the element tree of src, sanitized HTML, below root.
// Like browsers it knows void elements, ends <p>, <li> and table cells
// that the next one implies, and ignores end tags of elements that are not
// open.
func parseFragment(src string, root *element) fragment {
	var f fragment
	stack := []*element{root}
	// closeUntil pops the innermost open element named in names, and those
	// inside it, unless an element named in stop comes first.
	closeUntil := func(names, stop []string) {
		for i := len(stack) - 1; i > 0; i-- {
			switch {
			case contains(names, stack[i].tag):
				stack = stack[:i]
				return
			case contains(stop, stack[i].tag):
				return
			}
		}
	}
	l := html.NewLexer(parse.NewInputString(src))
	var tag string
	var attrs []attr
	for {
		tt, data := l.Next()
		switch tt {
		case html.ErrorToken:
			return f
		case html.StartTagToken:
			tag = strings.ToLower(string(l.Text()))
			attrs = nil
		case html.AttributeToken:
			attrs = append(attrs, attr{strings.ToLower(string(l.AttrKey())), unquote(string(l.AttrVal()))})
		case html.StartTagCloseToken, html.StartTagVoidToken:
			if ends, ok := impliedEnds[tag]; ok {
				closeUntil(ends[0], ends[1])
			}
			if closesP[tag] {
				closeUntil([]string{"p"}, pScope)
			}
			e := newElement(tag, attrs, stack[len(stack)-1])
			f.elements = append(f.elements, e)
			f.items = append(f.items, item{el: e, void: tt == html.StartTagVoidToken})
			if tt == html.StartTagCloseToken && !voidElements[tag] {
				stack = append(stack, e)
			}
		case html.EndTagToken:
			closeUntil([]string{strings.ToLower(string(l.Text()))}, nil)
			f.items = append(f.items, item{text: string(data)})
		default:
			f.items = append(f.items, item{text: string(data)})
		}
	}
}

// write writes the fragment back, with style as the style attribute of the
// elements it has one for.
func (f fragment) write(style map[*element]string) string {
	var out strings.Builder
	for _, it := range f.items {
		if it.el == nil {
			out.WriteString(it.text)
			continue
		}
		attrs := it.el.attrs
		if s, ok := style[it.el]; ok {
			attrs = withAttr(attrs, "style", s)
		}
		writeTag(&out, it.el.tag, attrs, it.void)
	}
	return out.String()
}

// withAttr returns a copy of attrs with key set to val.
func withAttr(attrs []attr, key, val string) []attr {
	out := make([]attr, 0, len(attrs)+1)
	set := false
	for _, a := range attrs {
		if a.key == key {
			if set {
				continue
			}
			a.val, set = val, true
		}
		out = append(out, a)
	}
	if !set {
		out = append(out, attr{key, val})
	}
	return out
}
