package render

import (
	"path"
	"sort"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// The app's HTML widget ignores <style>, so the note type's CSS is applied
// by inlining it: each element gets the declarations that win for it in
// the cascade, as its style attribute. The selectors of selector.go are
// matched against the element tree of tree.go, in which the card's root
// classes (".card", ".nightMode", ".mobile"…) are on the root and on an
// element above it, so rules for them style the card itself. @media blocks
// apply when they match the environment. url() values are dropped, except
// in background-image when they name media files. @font-face rules for
// TrueType and OpenType files in the media are reported, for the app to
// load; font-family declarations are inlined as written. Custom properties
// (--name) are inlined too, for the app to read, such as the colours of
// image occlusion shapes.

type decl struct {
	prop, value string
	important   bool
}

type cssRule struct {
	sel   selector
	decls []decl
	order int
}

type stylesheet struct {
	// root holds the card's root classes.
	root []string
	// The rules by the id, first class or element of their selector's last
	// compound, as browsers index them; universal holds the others.
	byID, byClass, byTag map[string][]*cssRule
	universal            []*cssRule
	fonts                []FontFace
}

// FontFace is an @font-face rule whose font is a TrueType or OpenType file
// in the collection's media. A family can have several faces. Weight and
// Style are the rule's descriptors as written, empty when absent.
type FontFace struct {
	Family string `json:"family"`
	File   string `json:"file"`
	Weight string `json:"weight,omitempty"`
	Style  string `json:"style,omitempty"`
}

var fontExts = map[string]bool{".ttf": true, ".otf": true}

// fontFile returns the first local TrueType or OpenType file among the
// values of an @font-face src, or "".
func fontFile(values []css.Token) string {
	for _, raw := range urls(values) {
		name := mediaName(raw)
		if name != "" && fontExts[strings.ToLower(path.Ext(name))] {
			return name
		}
	}
	return ""
}

// urls returns the addresses of the url() values among values, unquoted.
func urls(values []css.Token) []string {
	var out []string
	for i, v := range values {
		switch {
		case v.TokenType == css.URLToken:
			raw := string(v.Data)
			raw = strings.TrimSpace(raw[strings.IndexByte(raw, '(')+1 : len(raw)-1])
			out = append(out, unquote(raw))
		case v.TokenType == css.FunctionToken && strings.EqualFold(string(v.Data), "url("):
			raw := ""
			if i+1 < len(values) && values[i+1].TokenType == css.StringToken {
				raw = unquote(string(values[i+1].Data))
			}
			out = append(out, raw)
		}
	}
	return out
}

// keepsURLs reports whether a declaration of prop may keep the url()
// values among values: background-image may when they all name media
// files, which the app loads itself. Other url() values could make the
// app fetch remote content.
func keepsURLs(prop string, values []css.Token) bool {
	if prop != "background-image" {
		return false
	}
	for _, u := range urls(values) {
		if mediaName(u) == "" {
			return false
		}
	}
	return true
}

// addFont adds face unless it is incomplete or already there.
func (s *stylesheet) addFont(face FontFace) {
	if face.Family == "" || face.File == "" {
		return
	}
	for _, f := range s.fonts {
		if f == face {
			return
		}
	}
	s.fonts = append(s.fonts, face)
}

// parseCSS keeps the rules that can apply to a card with the root classes
// root, shown in env.
func parseCSS(src string, env Env, root []string) stylesheet {
	sheet := stylesheet{root: root, byID: map[string][]*cssRule{}, byClass: map[string][]*cssRule{}, byTag: map[string][]*cssRule{}}
	p := css.NewParser(parse.NewInputString(src), false)
	// One entry per enclosing at-rule: whether its contents apply.
	var atRules []bool
	skip := func() bool {
		for _, ok := range atRules {
			if !ok {
				return true
			}
		}
		return false
	}
	var selectors []selector
	inRule := false
	var decls []decl
	// The @font-face rule being read, if it applies.
	var face *FontFace
	order := 0
	for {
		gt, _, data := p.Next()
		switch gt {
		case css.ErrorGrammar:
			return sheet
		case css.BeginAtRuleGrammar:
			name := strings.ToLower(string(data))
			if name == "@font-face" && !skip() {
				face = &FontFace{}
			}
			atRules = append(atRules, name == "@media" && env.matchMedia(tokens(p.Values())))
		case css.EndAtRuleGrammar:
			if face != nil {
				sheet.addFont(*face)
				face = nil
			}
			if len(atRules) > 0 {
				atRules = atRules[:len(atRules)-1]
			}
		case css.BeginRulesetGrammar:
			selectors, decls = nil, nil
			inRule = !skip()
			if !inRule {
				continue
			}
			for _, part := range splitList(p.Values()) {
				if sel, ok := parseSelector(part); ok {
					selectors = append(selectors, sel)
				}
			}
		case css.DeclarationGrammar, css.CustomPropertyGrammar:
			if face != nil {
				value := strings.TrimSpace(tokens(p.Values()))
				switch strings.ToLower(string(data)) {
				case "font-family":
					face.Family = strings.TrimSpace(unquote(value))
				case "src":
					face.File = fontFile(p.Values())
				case "font-weight":
					face.Weight = strings.ToLower(value)
				case "font-style":
					face.Style = strings.ToLower(value)
				}
				continue
			}
			if !inRule || selectors == nil {
				continue
			}
			if gt == css.CustomPropertyGrammar {
				if d, ok := customProperty(data, p.Values()); ok {
					decls = append(decls, d)
				}
				continue
			}
			prop := strings.ToLower(string(data))
			if len(urls(p.Values())) > 0 && !keepsURLs(prop, p.Values()) {
				continue
			}
			if d, ok := declaration(prop, p.Values()); ok {
				decls = append(decls, d)
			}
		case css.EndRulesetGrammar:
			if !inRule {
				continue
			}
			for _, sel := range selectors {
				sheet.add(sel, decls, order)
				order++
			}
			selectors, inRule = nil, false
		}
	}
}

func tokens(values []css.Token) string {
	var sb strings.Builder
	for _, v := range values {
		sb.Write(v.Data)
	}
	return sb.String()
}

// declaration returns the declaration of prop with the value tokens
// values, without a trailing !important, which it records.
func declaration(prop string, values []css.Token) (decl, bool) {
	values = trimSpace(values)
	d := decl{prop: prop}
	if n := len(values); n >= 2 && values[n-1].TokenType == css.IdentToken &&
		strings.EqualFold(string(values[n-1].Data), "important") {
		// "!" and "important", maybe with whitespace between them.
		rest := trimSpace(values[:n-1])
		if len(rest) > 0 && isDelim(rest[len(rest)-1], "!") {
			values, d.important = trimSpace(rest[:len(rest)-1]), true
		}
	}
	d.value = tokens(values)
	return d, d.value != ""
}

// customProperty returns the declaration of a custom property such as
// --active-shape-color, whose name keeps its case. One with a url() is
// dropped, like other url() values the app does not load itself.
func customProperty(name []byte, values []css.Token) (decl, bool) {
	value := strings.TrimSpace(tokens(values))
	if value == "" || strings.Contains(strings.ToLower(value), "url(") {
		return decl{}, false
	}
	d := decl{prop: string(name), value: value}
	if v, ok := strings.CutSuffix(value, "important"); ok {
		if v, ok := strings.CutSuffix(strings.TrimSpace(v), "!"); ok {
			d.value, d.important = strings.TrimSpace(v), true
		}
	}
	return d, d.value != ""
}

// splitList splits a selector list at its top-level commas.
func splitList(toks []css.Token) [][]css.Token {
	var parts [][]css.Token
	depth, start := 0, 0
	for i, t := range toks {
		switch t.TokenType {
		case css.FunctionToken, css.LeftParenthesisToken, css.LeftBracketToken:
			depth++
		case css.RightParenthesisToken, css.RightBracketToken:
			depth--
		case css.CommaToken:
			if depth == 0 {
				parts = append(parts, toks[start:i])
				start = i + 1
			}
		}
	}
	return append(parts, toks[start:])
}

// add adds a rule for sel to the index.
func (s *stylesheet) add(sel selector, decls []decl, order int) {
	r := &cssRule{sel: sel, decls: decls, order: order}
	switch kind, name := sel.key(); kind {
	case '#':
		s.byID[name] = append(s.byID[name], r)
	case '.':
		s.byClass[name] = append(s.byClass[name], r)
	case 'e':
		s.byTag[name] = append(s.byTag[name], r)
	default:
		s.universal = append(s.universal, r)
	}
}

func (s *stylesheet) empty() bool {
	return len(s.byID) == 0 && len(s.byClass) == 0 && len(s.byTag) == 0 && len(s.universal) == 0
}

// Cascade levels, lowest first, as in browsers: the element's own style
// attribute beats normal rules, !important rules beat it unless it is
// !important too.
const (
	levelNormal = iota
	levelOwn
	levelImportant
	levelOwnImportant
)

// cascade returns the declarations that win for e, given its own style
// attribute own, or nil when no rule matches e.
func (s *stylesheet) cascade(e *element, own string) []decl {
	type weighted struct {
		decl
		level, spec, order int
	}
	var all []weighted
	add := func(rules []*cssRule) {
		for _, r := range rules {
			if !r.sel.matches(e) {
				continue
			}
			for _, d := range r.decls {
				level := levelNormal
				if d.important {
					level = levelImportant
				}
				all = append(all, weighted{d, level, r.sel.spec, r.order})
			}
		}
	}
	add(s.universal)
	add(s.byTag[e.tag])
	if e.id != "" {
		add(s.byID[e.id])
	}
	for _, c := range e.classes {
		add(s.byClass[c])
	}
	if all == nil {
		return nil
	}
	for _, d := range parseStyle(own) {
		level := levelOwn
		if d.important {
			level = levelOwnImportant
		}
		all = append(all, weighted{decl: d, level: level})
	}
	// Stable, so a rule's declarations keep their order.
	sort.SliceStable(all, func(i, j int) bool {
		a, b := all[i], all[j]
		if a.level != b.level {
			return a.level < b.level
		}
		if a.spec != b.spec {
			return a.spec < b.spec
		}
		return a.order < b.order
	})
	decls := make([]decl, len(all))
	for i, w := range all {
		decls[i] = decl{prop: w.prop, value: w.value}
	}
	return lastWins(decls)
}

// parseStyle returns the declarations of a style attribute.
func parseStyle(style string) []decl {
	if strings.TrimSpace(style) == "" {
		return nil
	}
	var decls []decl
	p := css.NewParser(parse.NewInputString(style), true)
	for {
		gt, _, data := p.Next()
		switch gt {
		case css.ErrorGrammar:
			return decls
		case css.DeclarationGrammar:
			if d, ok := declaration(strings.ToLower(string(data)), p.Values()); ok {
				decls = append(decls, d)
			}
		case css.CustomPropertyGrammar:
			if d, ok := customProperty(data, p.Values()); ok {
				decls = append(decls, d)
			}
		}
	}
}

// cardDecls returns the declarations for the card's root element.
func (s *stylesheet) cardDecls() []decl {
	return s.cascade(cardRoot(s.root), "")
}

// matching returns the declarations for an element with the tag and
// classes right inside the card's root.
func (s *stylesheet) matching(tag string, classes []string) []decl {
	e := newElement(tag, []attr{{"class", strings.Join(classes, " ")}}, cardRoot(s.root))
	return s.cascade(e, "")
}

// lastWins keeps only the declaration that wins for each property.
func lastWins(decls []decl) []decl {
	last := map[string]int{}
	for i, d := range decls {
		last[d.prop] = i
	}
	var out []decl
	for i, d := range decls {
		if last[d.prop] == i {
			out = append(out, d)
		}
	}
	return out
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func styleString(decls []decl) string {
	parts := make([]string, len(decls))
	for i, d := range decls {
		parts[i] = d.prop + ":" + d.value
	}
	return strings.Join(parts, ";")
}

// inline sets the style attribute of each element of src, the content
// of a card side, to the declarations that win for it, when rules match
// it.
func (s *stylesheet) inline(src string) string {
	if s.empty() {
		return src
	}
	f := parseFragment(src, cardRoot(s.root))
	style := map[*element]string{}
	for _, e := range f.elements {
		own, _ := e.attr("style")
		if decls := s.cascade(e, own); decls != nil {
			style[e] = styleString(decls)
		}
	}
	return f.write(style)
}

// cardColors splits .card declarations into the colours the app applies to
// the card surface and the remaining style for the card's content. Anki's
// stock colours (black on white) count as unset, so cards follow the app
// theme unless the note type chose its own colours.
func cardColors(decls []decl) (background, color string, rest []decl) {
	for _, d := range decls {
		v := strings.ToLower(strings.TrimSpace(d.value))
		switch d.prop {
		case "background", "background-color":
			if v != "white" && v != "#fff" && v != "#ffffff" && !strings.ContainsAny(v, " ") {
				background = d.value
			}
		case "color":
			if v != "black" && v != "#000" && v != "#000000" {
				color = d.value
			}
		default:
			rest = append(rest, d)
		}
	}
	return background, color, lastWins(rest)
}
