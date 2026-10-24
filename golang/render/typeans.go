package render

import (
	"html"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// maxTypeAnsRunes caps the texts compared character by character; longer
// ones are shown side by side without marks, so the diff stays cheap.
const maxTypeAnsRunes = 2000

// CompareTyped renders the comparison of what the user typed with the
// expected field content, as Anki does: a <code id=typeans> block that holds
// the expected text marked typeGood when everything matches, and otherwise
// the typed text (typeGood/typeBad), an arrow and the expected text
// (typeGood/typeMissed). The expected HTML is stripped to text; both sides
// are NFC-normalised and trimmed. ignoreCombining drops accents and other
// combining marks from both sides first (the nc: variant). When nothing was
// typed, the block holds the expected text unmarked.
func CompareTyped(expected, typed string, ignoreCombining bool) string {
	exp := norm.NFC.String(strings.TrimSpace(StripHTML(expected)))
	got := norm.NFC.String(strings.TrimSpace(typed))
	if got == "" {
		return `<code id="typeans">` + html.EscapeString(exp) + `</code>`
	}
	expCmp, gotCmp := exp, got
	if ignoreCombining {
		expCmp, gotCmp = stripCombining(exp), stripCombining(got)
	}
	if expCmp == gotCmp {
		return `<code id="typeans"><span class="typeGood">` + html.EscapeString(exp) + `</span></code>`
	}
	const arrow = `<br><span id="typearrow">&darr;</span><br>`
	e, g := []rune(expCmp), []rune(gotCmp)
	if len(e) > maxTypeAnsRunes || len(g) > maxTypeAnsRunes {
		return `<code id="typeans">` + html.EscapeString(gotCmp) + arrow + html.EscapeString(expCmp) + `</code>`
	}
	inLCSg, inLCSe := lcsMarks(g, e)
	var sb strings.Builder
	sb.WriteString(`<code id="typeans">`)
	writeMarked(&sb, g, inLCSg, "typeGood", "typeBad")
	sb.WriteString(arrow)
	writeMarked(&sb, e, inLCSe, "typeGood", "typeMissed")
	sb.WriteString(`</code>`)
	return sb.String()
}

// lcsMarks reports for each rune of a and of b whether it belongs to a
// longest common subsequence of the two.
func lcsMarks(a, b []rune) ([]bool, []bool) {
	n, m := len(a), len(b)
	// l[i][j] is the LCS length of a[i:] and b[j:].
	l := make([][]uint16, n+1)
	for i := range l {
		l[i] = make([]uint16, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case a[i] == b[j]:
				l[i][j] = l[i+1][j+1] + 1
			case l[i+1][j] >= l[i][j+1]:
				l[i][j] = l[i+1][j]
			default:
				l[i][j] = l[i][j+1]
			}
		}
	}
	ma, mb := make([]bool, n), make([]bool, m)
	for i, j := 0, 0; i < n && j < m; {
		switch {
		case a[i] == b[j]:
			ma[i], mb[j] = true, true
			i++
			j++
		case l[i+1][j] >= l[i][j+1]:
			i++
		default:
			j++
		}
	}
	return ma, mb
}

// writeMarked writes rs as runs of <span class=good> for marked runes and
// <span class=bad> for the others.
func writeMarked(sb *strings.Builder, rs []rune, marked []bool, good, bad string) {
	for i := 0; i < len(rs); {
		j := i
		for j < len(rs) && marked[j] == marked[i] {
			j++
		}
		class := bad
		if marked[i] {
			class = good
		}
		sb.WriteString(`<span class="` + class + `">` + html.EscapeString(string(rs[i:j])) + `</span>`)
		i = j
	}
}

// stripCombining removes combining marks (accents, diacritics) after
// decomposing s, so "über" becomes "uber".
func stripCombining(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, norm.NFD.String(s))
	return norm.NFC.String(s)
}
