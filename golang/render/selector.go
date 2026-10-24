package render

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/tdewolff/parse/v2/css"
)

// Supported selectors: element, universal, class, id and attribute
// ([attr], [attr=value]) selectors, the descendant, child (>), adjacent
// sibling (+) and general sibling (~) combinators, and the pseudo-classes
// :first-child, :last-child, :only-child, :nth-child() and :not() with one
// simple selector. A selector with anything else never matches.

type combinator byte

const (
	descendant combinator = ' '
	child      combinator = '>'
	adjacent   combinator = '+'
	sibling    combinator = '~'
)

// compound is a sequence of simple selectors without a combinator, such as
// "div.a#b:first-child".
type compound struct {
	tag     string // "" for any element
	id      string
	classes []string
	attrs   []attrSel
	pseudos []pseudoClass
	// comb joins the compound to the one on its left.
	comb combinator
}

type attrSel struct {
	key, val string
	hasVal   bool
}

type pseudoClass struct {
	name string // first-child, last-child, only-child, nth-child or not
	a, b int    // nth-child(an+b)
	not  *compound
}

type selector struct {
	compounds []compound // left to right
	spec      int
	// ancestors holds the Bloom filter bits of the elements, ids and
	// classes that ancestors of a matching element must have.
	ancestors uint64
}

// Specificity counts ids, then classes, attributes and pseudo-classes, then
// elements; each count is assumed to stay below 100.
const (
	specID    = 10000
	specClass = 100
	specElem  = 1
)

// parseSelector compiles one selector of a selector list; ok is false when
// it uses anything unsupported.
func parseSelector(toks []css.Token) (sel selector, ok bool) {
	toks = trimSpace(toks)
	if len(toks) == 0 {
		return sel, false
	}
	p := selectorParser{toks: toks}
	comb := descendant
	for {
		c, ok := p.compound()
		if !ok {
			return sel, false
		}
		c.comb = comb
		sel.compounds = append(sel.compounds, c)
		sel.spec += c.spec()
		if p.done() {
			sel.ancestors = sel.ancestorBits()
			return sel, true
		}
		if comb, ok = p.combinator(); !ok {
			return sel, false
		}
	}
}

func trimSpace(toks []css.Token) []css.Token {
	for len(toks) > 0 && toks[0].TokenType == css.WhitespaceToken {
		toks = toks[1:]
	}
	for len(toks) > 0 && toks[len(toks)-1].TokenType == css.WhitespaceToken {
		toks = toks[:len(toks)-1]
	}
	return toks
}

type selectorParser struct {
	toks []css.Token
	i    int
}

func (p *selectorParser) done() bool { return p.i >= len(p.toks) }

func (p *selectorParser) peek() (css.Token, bool) {
	if p.done() {
		return css.Token{}, false
	}
	return p.toks[p.i], true
}

func isDelim(t css.Token, d string) bool {
	return t.TokenType == css.DelimToken && string(t.Data) == d
}

// combinator reads whitespace and at most one of >, + and ~.
func (p *selectorParser) combinator() (combinator, bool) {
	comb := combinator(0)
	for t, ok := p.peek(); ok; t, ok = p.peek() {
		switch {
		case t.TokenType == css.WhitespaceToken:
			if comb == 0 {
				comb = descendant
			}
		case comb != 0 && comb != descendant:
			return comb, true
		case isDelim(t, ">"), isDelim(t, "+"), isDelim(t, "~"):
			comb = combinator(t.Data[0])
		default:
			return comb, comb != 0
		}
		p.i++
	}
	return 0, false // a selector can't end in a combinator
}

// compound reads the simple selectors up to the next combinator.
func (p *selectorParser) compound() (c compound, ok bool) {
	if t, ok := p.peek(); ok {
		switch {
		case t.TokenType == css.IdentToken:
			c.tag = strings.ToLower(string(t.Data))
			p.i++
		case isDelim(t, "*"):
			p.i++
			if p.done() || p.toks[p.i].TokenType == css.WhitespaceToken || isCombinator(p.toks[p.i]) {
				return c, true
			}
		}
	}
	start := p.i
	for t, more := p.peek(); more; t, more = p.peek() {
		switch {
		case t.TokenType == css.HashToken:
			c.id = string(t.Data[1:])
			p.i++
		case isDelim(t, "."):
			p.i++
			name, ok := p.peek()
			if !ok || name.TokenType != css.IdentToken {
				return c, false
			}
			c.classes = append(c.classes, string(name.Data))
			p.i++
		case t.TokenType == css.LeftBracketToken:
			a, ok := p.attribute()
			if !ok {
				return c, false
			}
			c.attrs = append(c.attrs, a)
		case t.TokenType == css.ColonToken:
			pc, ok := p.pseudoClass()
			if !ok {
				return c, false
			}
			c.pseudos = append(c.pseudos, pc)
		case t.TokenType == css.WhitespaceToken || isCombinator(t):
			return c, c.tag != "" || p.i > start
		default:
			return c, false
		}
	}
	return c, c.tag != "" || p.i > start
}

func isCombinator(t css.Token) bool {
	return isDelim(t, ">") || isDelim(t, "+") || isDelim(t, "~")
}

// attribute reads [key] or [key=value].
func (p *selectorParser) attribute() (a attrSel, ok bool) {
	p.i++ // [
	var parts []css.Token
	for t, more := p.peek(); more; t, more = p.peek() {
		p.i++
		if t.TokenType == css.RightBracketToken {
			break
		}
		if t.TokenType != css.WhitespaceToken {
			parts = append(parts, t)
		}
		if p.done() {
			return a, false
		}
	}
	if len(parts) == 0 || parts[0].TokenType != css.IdentToken {
		return a, false
	}
	a.key = strings.ToLower(string(parts[0].Data))
	switch {
	case len(parts) == 1:
		return a, true
	case len(parts) == 3 && isDelim(parts[1], "="):
		switch parts[2].TokenType {
		case css.IdentToken:
			a.val = string(parts[2].Data)
		case css.StringToken:
			a.val = unquote(string(parts[2].Data))
		default:
			return a, false
		}
		a.hasVal = true
		return a, true
	}
	return a, false
}

// pseudoClass reads one of the supported pseudo-classes after its colon.
func (p *selectorParser) pseudoClass() (pc pseudoClass, ok bool) {
	p.i++ // :
	t, more := p.peek()
	if !more {
		return pc, false
	}
	p.i++
	switch t.TokenType {
	case css.IdentToken:
		pc.name = strings.ToLower(string(t.Data))
		switch pc.name {
		case "first-child", "last-child", "only-child":
			return pc, true
		}
		return pc, false
	case css.FunctionToken:
		pc.name = strings.ToLower(strings.TrimSuffix(string(t.Data), "("))
	default:
		return pc, false // also pseudo-elements (::before)
	}
	var args []css.Token
	for depth := 1; ; {
		t, more := p.peek()
		if !more {
			return pc, false
		}
		p.i++
		switch t.TokenType {
		case css.FunctionToken, css.LeftParenthesisToken:
			depth++
		case css.RightParenthesisToken:
			depth--
		}
		if depth == 0 {
			break
		}
		args = append(args, t)
	}
	switch pc.name {
	case "nth-child":
		pc.a, pc.b, ok = parseNth(tokens(args))
		return pc, ok
	case "not":
		inner := selectorParser{toks: trimSpace(args)}
		c, ok := inner.compound()
		if !ok || !inner.done() || c.simpleCount() != 1 {
			return pc, false
		}
		for _, q := range c.pseudos {
			if q.name == "not" {
				return pc, false
			}
		}
		pc.not = &c
		return pc, true
	}
	return pc, false
}

var nthPattern = regexp.MustCompile(`^(?:([+-]?\d*)n([+-]\d+)?|([+-]?\d+))$`)

// parseNth parses the an+b argument of :nth-child(), also "odd" and
// "even".
func parseNth(arg string) (a, b int, ok bool) {
	arg = strings.ToLower(strings.Join(strings.Fields(arg), ""))
	switch arg {
	case "odd":
		return 2, 1, true
	case "even":
		return 2, 0, true
	}
	m := nthPattern.FindStringSubmatch(arg)
	if m == nil {
		return 0, 0, false
	}
	if m[3] != "" {
		b, _ = strconv.Atoi(m[3])
		return 0, b, true
	}
	switch m[1] {
	case "", "+":
		a = 1
	case "-":
		a = -1
	default:
		a, _ = strconv.Atoi(m[1])
	}
	if m[2] != "" {
		b, _ = strconv.Atoi(m[2])
	}
	return a, b, true
}

// simpleCount is the number of simple selectors in c.
func (c *compound) simpleCount() int {
	n := len(c.classes) + len(c.attrs) + len(c.pseudos)
	if c.tag != "" {
		n++
	}
	if c.id != "" {
		n++
	}
	if n == 0 {
		n = 1 // *
	}
	return n
}

func (c *compound) spec() int {
	s := specClass * (len(c.classes) + len(c.attrs))
	if c.id != "" {
		s += specID
	}
	if c.tag != "" {
		s += specElem
	}
	for _, pc := range c.pseudos {
		if pc.not != nil {
			s += pc.not.spec()
		} else {
			s += specClass
		}
	}
	return s
}

// ancestorBits returns the Bloom filter bits of the compounds left of a
// descendant or child combinator: they match ancestors of the element, as
// siblings share their ancestors.
func (s *selector) ancestorBits() uint64 {
	var m uint64
	ancestor := false
	for i := len(s.compounds) - 1; i >= 0; i-- {
		c := s.compounds[i]
		if ancestor {
			if c.tag != "" {
				m |= bloomBit('e', c.tag)
			}
			if c.id != "" {
				m |= bloomBit('#', c.id)
			}
			for _, cl := range c.classes {
				m |= bloomBit('.', cl)
			}
		}
		if c.comb == descendant || c.comb == child {
			ancestor = true
		}
	}
	return m
}

// key returns the bucket of the rule index a selector goes in: its last
// compound's id, else its first class, else its element, else "*".
func (s *selector) key() (kind byte, name string) {
	c := s.compounds[len(s.compounds)-1]
	switch {
	case c.id != "":
		return '#', c.id
	case len(c.classes) > 0:
		return '.', c.classes[0]
	case c.tag != "":
		return 'e', c.tag
	}
	return '*', ""
}

// matches reports whether e matches s, matching from the right.
func (s *selector) matches(e *element) bool {
	if s.ancestors&^e.ancestors != 0 {
		return false
	}
	return matchFrom(s.compounds, len(s.compounds)-1, e)
}

func matchFrom(cs []compound, i int, e *element) bool {
	if !cs[i].matches(e) {
		return false
	}
	if i == 0 {
		return true
	}
	switch cs[i].comb {
	case child:
		return e.parent != nil && matchFrom(cs, i-1, e.parent)
	case descendant:
		for a := e.parent; a != nil; a = a.parent {
			if matchFrom(cs, i-1, a) {
				return true
			}
		}
	case adjacent:
		p := e.prev()
		return p != nil && matchFrom(cs, i-1, p)
	case sibling:
		for p := e.prev(); p != nil; p = p.prev() {
			if matchFrom(cs, i-1, p) {
				return true
			}
		}
	}
	return false
}

func (c *compound) matches(e *element) bool {
	if c.tag != "" && c.tag != e.tag {
		return false
	}
	if c.id != "" && c.id != e.id {
		return false
	}
	for _, cl := range c.classes {
		if !contains(e.classes, cl) {
			return false
		}
	}
	for _, a := range c.attrs {
		v, ok := e.attr(a.key)
		if !ok || (a.hasVal && v != a.val) {
			return false
		}
	}
	for _, pc := range c.pseudos {
		if !pc.matches(e) {
			return false
		}
	}
	return true
}

func (pc *pseudoClass) matches(e *element) bool {
	// The position among the parent's children, from 1; the topmost
	// element is the only child of the document.
	pos, count := 1, 1
	if e.parent != nil {
		pos, count = e.index+1, len(e.parent.children)
	}
	switch pc.name {
	case "first-child":
		return pos == 1
	case "last-child":
		return pos == count
	case "only-child":
		return count == 1
	case "nth-child":
		if pc.a == 0 {
			return pos == pc.b
		}
		n := pos - pc.b
		return n%pc.a == 0 && n/pc.a >= 0
	case "not":
		return !pc.not.matches(e)
	}
	return false
}
