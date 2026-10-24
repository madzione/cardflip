package render

import (
	"crypto/sha1"
	"encoding/hex"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Anki shows [latex]…[/latex], [$]…[/$] and [$$]…[/$$] as images it
// generates with TeX and stores in the media, named after the source:
// "latex-" + the hex SHA-1 of the normalised source + ".png", or ".svg"
// when the note type is set to SVG. Shared decks ship those images, so
// reproducing the names exactly (rslib latex.rs) finds them without TeX.

// latexBlock matches the three kinds of block, case-insensitively and
// across lines, as Anki does: the first closer ends a block, and an empty
// block is no block.
var latexBlock = regexp.MustCompile(`(?si)\[latex\](.+?)\[/latex\]|\[\$\](.+?)\[/\$\]|\[\$\$\](.+?)\[/\$\$\]`)

// latexNewline is Anki's `<br( /)?>|<div>` in extended mode, where the
// space is ignored: <br> and <br/> are newlines, <br /> is removed with the
// other tags.
var latexNewline = regexp.MustCompile(`(?i)<br/?>|<div>`)

// latexTags is Anki's strip_html: comments, styles, scripts and then
// anything from < to the next >, even inside a quoted attribute.
var latexTags = regexp.MustCompile(`(?si)<!--.*?-->|<style.*?>.*?</style>|<script.*?>.*?</script>|<.*?>`)

// latexSource is a block's LaTeX as Anki normalises it from the field's
// HTML: <br> and <div> become newlines, other tags go and entities are
// decoded.
func latexSource(raw string) string {
	s := latexNewline.ReplaceAllString(raw, "\n")
	s = latexTags.ReplaceAllString(s, "")
	return decodeLatexEntities(s)
}

// decodeLatexEntities decodes character references as Anki does: every &
// must start an HTML 4.01 named reference or a decimal or &#x hex one, all
// ending in ";". Otherwise the text stays as it is, references and all.
// After decoding, no-break spaces become spaces.
func decodeLatexEntities(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}
	var sb strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '&' {
			sb.WriteByte(s[i])
			i++
			continue
		}
		end := strings.IndexByte(s[i+1:], ';')
		if end < 0 {
			return s
		}
		r, ok := latexEntity(s[i+1 : i+1+end])
		if !ok {
			return s
		}
		sb.WriteRune(r)
		i += end + 2
	}
	return strings.ReplaceAll(sb.String(), " ", " ")
}

// latexEntity decodes the name of a character reference, the text between
// & and ;.
func latexEntity(name string) (rune, bool) {
	if digits, ok := strings.CutPrefix(name, "#"); ok {
		base := 10
		if hexDigits, ok := strings.CutPrefix(digits, "x"); ok {
			digits, base = hexDigits, 16
		}
		n, err := strconv.ParseUint(digits, base, 32)
		if err != nil || !utf8.ValidRune(rune(n)) {
			return 0, false
		}
		return rune(n), true
	}
	r, ok := htmlEntities[name]
	return r, ok
}

// latexFile is the name of the image Anki generates for LaTeX source.
func latexFile(src string, svg bool) string {
	sum := sha1.Sum([]byte(src))
	ext := ".png"
	if svg {
		ext = ".svg"
	}
	return "latex-" + hex.EncodeToString(sum[:]) + ext
}

// latexBlocks calls fn for each LaTeX block in rendered HTML with its
// normalised source, [$] wrapped in $…$ and [$$] in a displaymath
// environment, and replaces the block with what fn returns.
func latexBlocks(src string, fn func(latex string) string) string {
	if !strings.Contains(src, "[") {
		return src
	}
	return latexBlock.ReplaceAllStringFunc(src, func(block string) string {
		m := latexBlock.FindStringSubmatch(block)
		switch {
		case m[1] != "":
			return fn(latexSource(m[1]))
		case m[2] != "":
			return fn("$" + latexSource(m[2]) + "$")
		default:
			return fn(`\begin{displaymath}` + latexSource(m[3]) + `\end{displaymath}`)
		}
	})
}

// latexImages replaces the LaTeX blocks in rendered HTML with the images
// Anki would have generated for them, PNG or (svg) SVG, when hasMedia finds
// them, and with their source as preformatted text otherwise. A nil
// hasMedia finds nothing. Backslashes in the source are written as
// references, so that wrapMath leaves \( and \[ in it alone.
func latexImages(src string, svg bool, hasMedia func(name string) bool) string {
	return latexBlocks(src, func(latex string) string {
		file := latexFile(latex, svg)
		if hasMedia != nil && hasMedia(file) {
			return `<img class="latex" alt="` + html.EscapeString(latex) + `" src="` + file + `">`
		}
		return `<pre class="latex">` + strings.ReplaceAll(html.EscapeString(latex), `\`, "&#92;") + `</pre>`
	})
}
