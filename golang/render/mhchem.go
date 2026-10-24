package render

import (
	"regexp"
	"strings"
)

// A subset of mhchem, translated to plain TeX before typesetting: \ce{} with
// element symbols (upright), counts (subscripts), charges, stoichiometric
// numbers, states such as (aq), "+" between species, hydrate dots, bonds and
// reaction arrows with optional [above][below] text; \pu{} with numbers and
// upright units. Anything else inside \ce is shown as \text{…}.

// translateChem replaces each \ce{…} and \pu{…} in tex with plain TeX.
func translateChem(tex string) string {
	if !strings.Contains(tex, `\ce`) && !strings.Contains(tex, `\pu`) {
		return tex
	}
	var sb strings.Builder
	for i := 0; i < len(tex); {
		name, arg, end, ok := chemCommand(tex, i)
		if !ok {
			sb.WriteByte(tex[i])
			i++
			continue
		}
		if name == "ce" {
			sb.WriteString("{" + translateCe(arg) + "}")
		} else {
			sb.WriteString("{" + translatePu(arg) + "}")
		}
		i = end
	}
	return sb.String()
}

// chemCommand recognises \ce{…} or \pu{…} at tex[i] and returns its name,
// its argument and the position after the closing brace.
func chemCommand(tex string, i int) (name, arg string, end int, ok bool) {
	for _, n := range []string{"ce", "pu"} {
		cmd := `\` + n
		if !strings.HasPrefix(tex[i:], cmd) {
			continue
		}
		j := i + len(cmd)
		if j < len(tex) && isLetter(tex[j]) {
			return "", "", 0, false // a longer command such as \cent
		}
		for j < len(tex) && tex[j] == ' ' {
			j++
		}
		if j >= len(tex) || tex[j] != '{' {
			return "", "", 0, false
		}
		depth := 0
		for k := j; k < len(tex); k++ {
			switch tex[k] {
			case '\\':
				k++ // an escaped character, e.g. \{
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return n, tex[j+1 : k], k + 1, true
				}
			}
		}
	}
	return "", "", 0, false
}

// arrows maps mhchem arrows to TeX, longest first so "<->" wins over "<-".
var arrows = []struct{ ce, tex, ext string }{
	{"<=>", `\rightleftharpoons`, `\xrightleftharpoons`},
	{"<->", `\leftrightarrow`, `\xleftrightarrow`},
	{"->", `\rightarrow`, `\xrightarrow`},
	{"<-", `\leftarrow`, `\xleftarrow`},
}

// arrowPattern finds arrows with their [above][below] text so they can be
// split from neighbouring species written without spaces.
var arrowPattern = regexp.MustCompile(`(<=>|<->|->|<-)((?:\[[^\]]*\]){0,2})`)

func translateCe(s string) string {
	s = arrowPattern.ReplaceAllString(s, " $1$2 ")
	var parts []string
	for _, tok := range ceTokens(s) {
		parts = append(parts, ceToken(tok))
	}
	return strings.Join(parts, " ")
}

// ceTokens splits on spaces outside the brackets of arrow text.
func ceTokens(s string) []string {
	var out []string
	var cur strings.Builder
	depth := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '[':
			depth++
		case c == ']' && depth > 0:
			depth--
		case (c == ' ' || c == '\t' || c == '\n') && depth == 0:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteByte(c)
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

var states = map[string]bool{"s": true, "l": true, "g": true, "aq": true, "cr": true}

func ceToken(tok string) string {
	switch tok {
	case "+":
		return "+"
	case "^":
		return `\uparrow`
	case "v":
		return `\downarrow`
	}
	if st := stateSuffix(tok); st != "" {
		return `(\mathrm{` + st + `})`
	}
	for _, a := range arrows {
		if !strings.HasPrefix(tok, a.ce) {
			continue
		}
		above, below, ok := arrowText(tok[len(a.ce):])
		if !ok {
			break
		}
		if above == "" && below == "" {
			return a.tex
		}
		out := a.ext
		if below != "" {
			out += "[" + ceText(below) + "]"
		}
		return out + "{" + ceText(above) + "}"
	}
	if tex, ok := species(tok); ok {
		return tex
	}
	return `\text{` + escapeText(tok) + `}`
}

// arrowText parses the optional "[above][below]" after an arrow.
func arrowText(s string) (above, below string, ok bool) {
	var texts []string
	for s != "" {
		if s[0] != '[' {
			return "", "", false
		}
		end := strings.IndexByte(s, ']')
		if end < 0 {
			return "", "", false
		}
		texts = append(texts, s[1:end])
		s = s[end+1:]
	}
	if len(texts) > 0 {
		above = texts[0]
	}
	if len(texts) > 1 {
		below = texts[1]
	}
	return above, below, true
}

// ceText typesets arrow text: formulas where they parse, numbers, and other
// words as text.
func ceText(s string) string {
	var parts []string
	for _, w := range strings.Fields(s) {
		if tex, ok := species(w); ok {
			parts = append(parts, tex)
		} else if isNumber(w) {
			parts = append(parts, w)
		} else {
			parts = append(parts, `\text{`+escapeText(w)+`}`)
		}
	}
	return strings.Join(parts, `\ `)
}

// species translates one chemical species such as 2H2O, SO4^2-, Ca2+,
// [Cu(NH3)4]^2+, CuSO4*5H2O or H2O(l). ok is false when tok is not one.
func species(tok string) (string, bool) {
	var out, roman strings.Builder
	flush := func() {
		if roman.Len() > 0 {
			out.WriteString(`\mathrm{` + roman.String() + `}`)
			roman.Reset()
		}
	}
	i := 0
	coefficient := func() {
		j := i
		for j < len(tok) && (isDigit(tok[j]) || tok[j] == '.' || tok[j] == '/') {
			j++
		}
		if j > i && j < len(tok) && (isUpper(tok[j]) || tok[j] == '(' || tok[j] == '[') {
			out.WriteString(tok[i:j])
			i = j
		}
	}
	coefficient()
	// Isotope prefix: ^{227}_{90}Th or ^227_90Th.
	if i < len(tok) && tok[i] == '^' {
		sup, n := scriptArg(tok[i+1:])
		if n == 0 {
			return "", false
		}
		out.WriteString("{}^{" + sup + "}")
		i += 1 + n
		if i < len(tok) && tok[i] == '_' {
			sub, n := scriptArg(tok[i+1:])
			if n == 0 {
				return "", false
			}
			out.WriteString("_{" + sub + "}")
			i += 1 + n
		}
	}
	atoms := false
	for i < len(tok) {
		c := tok[i]
		rest := tok[i:]
		switch {
		case isUpper(c):
			j := i + 1
			for j < len(tok) && isLower(tok[j]) {
				j++
			}
			roman.WriteString(tok[i:j])
			i, atoms = j, true
		case c == 'e' && !atoms && (i+1 == len(tok) || tok[i+1] == '-' || tok[i+1] == '^'):
			// An electron.
			out.WriteString("e")
			i, atoms = i+1, true
		case c == '(' && stateSuffix(rest) != "":
			flush()
			out.WriteString(`(\mathrm{` + stateSuffix(rest) + `})`)
			i = len(tok)
		case c == '(' || c == ')' || c == '[' || c == ']':
			roman.WriteByte(c)
			i++
		case isDigit(c):
			j := i
			for j < len(tok) && isDigit(tok[j]) {
				j++
			}
			flush()
			if j < len(tok) && (tok[j] == '+' || tok[j] == '-') && chargeEnd(tok, j+1) {
				// Ca2+: a charge, not a count.
				out.WriteString("^{" + tok[i:j+1] + "}")
				i = j + 1
				continue
			}
			out.WriteString("_{" + tok[i:j] + "}")
			i = j
		case c == '^':
			sup, n := scriptArg(rest[1:])
			if n == 0 {
				return "", false
			}
			flush()
			out.WriteString("^{" + sup + "}")
			i += 1 + n
		case (c == '+' || c == '-') && chargeEnd(tok, i+1):
			flush()
			out.WriteString("^{" + string(c) + "}")
			i++
		case c == '-' || c == '=' || c == '#':
			// Bonds between atoms.
			flush()
			out.WriteString(map[byte]string{'-': "{-}", '=': "{=}", '#': `{\equiv}`}[c])
			i++
		case c == '*' || c == '.':
			// Hydrates: CuSO4*5H2O.
			flush()
			out.WriteString(`\cdot `)
			i++
			coefficient()
		default:
			return "", false
		}
	}
	flush()
	if !atoms {
		return "", false
	}
	return out.String(), true
}

// stateSuffix returns "aq" when s is "(aq)" (or another state) at the end of
// a species.
func stateSuffix(s string) string {
	if len(s) < 3 || s[0] != '(' || s[len(s)-1] != ')' {
		return ""
	}
	if inner := s[1 : len(s)-1]; states[inner] {
		return inner
	}
	return ""
}

// chargeEnd reports whether a charge sign ending before tok[i] ends the
// species, possibly followed by a state.
func chargeEnd(tok string, i int) bool {
	return i == len(tok) || stateSuffix(tok[i:]) != ""
}

// scriptArg reads a superscript or subscript argument: {…} or a run of
// digits and charge signs. It returns the argument and the bytes consumed.
func scriptArg(s string) (string, int) {
	if strings.HasPrefix(s, "{") {
		if end := strings.IndexByte(s, '}'); end > 0 {
			return s[1:end], end + 1
		}
		return "", 0
	}
	j := 0
	for j < len(s) && (isDigit(s[j]) || s[j] == '+' || s[j] == '-') {
		j++
	}
	return s[:j], j
}

func translatePu(s string) string {
	var parts []string
	for _, w := range strings.Fields(s) {
		if isNumber(w) {
			parts = append(parts, w)
		} else {
			parts = append(parts, `\mathrm{`+strings.ReplaceAll(w, "*", `\cdot `)+`}`)
		}
	}
	return strings.Join(parts, `\,`)
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isDigit(s[i]) && !strings.ContainsRune(".,-+", rune(s[i])) {
			return false
		}
	}
	return true
}

// escapeText makes s safe inside \text{}.
func escapeText(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '{', '}', '_', '#', '%', '&', '$':
			sb.WriteString(`\` + string(c))
		case '^':
			sb.WriteString(`\textasciicircum{}`)
		case '~':
			sb.WriteString(`\textasciitilde{}`)
		case '\\':
			sb.WriteString(`\textbackslash{}`)
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isUpper(c byte) bool { return c >= 'A' && c <= 'Z' }
func isLower(c byte) bool { return c >= 'a' && c <= 'z' }
