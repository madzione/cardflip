package render

import (
	"regexp"
	"strconv"
	"strings"
)

// Note is the data a card is rendered from.
type Note struct {
	// Fields maps field names to their HTML content.
	Fields map[string]string
	Tags   []string
	// NoteType, Deck and CardName feed the special fields {{Type}},
	// {{Deck}}, {{Subdeck}} and {{Card}}.
	NoteType string
	Deck     string
	CardName string
	// CardOrd is the card type number starting at 1 (the cloze number for
	// cloze cards), for the root class "cardN"; 0 leaves it out.
	CardOrd int
	// Flag is the card's flag, 0 for none or 1 to 7, for {{CardFlag}}.
	Flag int
	// LatexSVG is set when the note type's LaTeX images are SVG, not PNG.
	LatexSVG bool
	// HasMedia reports whether the collection's media holds a file, for
	// LaTeX images; nil when there is no media.
	HasMedia func(name string) bool
}

// Template is one card template of a note type.
type Template struct {
	QFmt string
	AFmt string
}

func (n Note) context(question bool, clozeOrd int) *context {
	deck := n.Deck
	sub := deck
	if i := strings.LastIndex(sub, "::"); i >= 0 {
		sub = sub[i+2:]
	}
	flag := ""
	if n.Flag != 0 {
		flag = "flag" + strconv.Itoa(n.Flag)
	}
	return &context{
		fields: n.Fields,
		special: map[string]string{
			"Tags":     strings.Join(n.Tags, " "),
			"Type":     n.NoteType,
			"Deck":     deck,
			"Subdeck":  sub,
			"Card":     n.CardName,
			"CardFlag": flag,
		},
		question: question,
		clozeOrd: clozeOrd,
	}
}

// Question renders the question side as raw (unsanitized) HTML. clozeOrd is
// the cloze number for cloze note types and 0 otherwise.
func Question(t Template, n Note, clozeOrd int) string {
	q, _ := question(t, n, clozeOrd)
	return q
}

// question renders the question side and returns the context it was
// rendered in, which holds the question's type: filter.
func question(t Template, n Note, clozeOrd int) (string, *context) {
	c := n.context(true, clozeOrd)
	c.special["FrontSide"] = ""
	var sb strings.Builder
	c.render(parseTemplate(t.QFmt), &sb)
	return sb.String(), c
}

// Answer renders the answer side as raw HTML; {{FrontSide}} is the question
// without its sounds and utterances, which belong to the question side, and
// without its answer input. {{type:…}} shows the expected text. When the
// answer has no type: filter of its own, the question's shows there in place
// of its input, below the <hr id=answer> line, which moves there as in Anki.
func Answer(t Template, n Note, clozeOrd int) string {
	return answer(t, n, clozeOrd, nil)
}

// AnswerTyped is Answer with {{type:…}} comparing typed with the expected
// text.
func AnswerTyped(t Template, n Note, clozeOrd int, typed string) string {
	return answer(t, n, clozeOrd, &typed)
}

// frontTypeMarker stands for the question's answer input inside
// {{FrontSide}} until the answer is rendered. Field content never holds the
// NUL characters.
const frontTypeMarker = "\x00typeans\x00"

func answer(t Template, n Note, clozeOrd int, typed *string) string {
	q, qc := question(t, n, clozeOrd)
	c := n.context(false, clozeOrd)
	c.typed = typed
	c.special["FrontSide"] = strings.ReplaceAll(stripSounds(q), typeAnsPlaceholder, frontTypeMarker)
	var sb strings.Builder
	c.render(parseTemplate(t.AFmt), &sb)
	a := sb.String()
	i := strings.Index(a, frontTypeMarker)
	if c.typeUsed || !qc.typeUsed || i < 0 {
		return strings.ReplaceAll(a, frontTypeMarker, "")
	}
	head, tail := a[:i], strings.ReplaceAll(a[i+len(frontTypeMarker):], frontTypeMarker, "")
	cmp := c.typeComparison(qc.typeExpected, qc.typeNC)
	if loc := answerLine.FindStringIndex(head); loc != nil {
		cmp = head[loc[0]:loc[1]] + cmp
		head = head[:loc[0]] + head[loc[1]:]
	} else if loc := answerLine.FindStringIndex(tail); loc != nil {
		cmp = tail[loc[0]:loc[1]] + cmp
		tail = tail[:loc[0]] + tail[loc[1]:]
	}
	return head + cmp + tail
}

// HasContent reports whether the question of this card substitutes at least
// one non-empty field. Cards whose question would be empty are not created.
func HasContent(qfmt string, n Note, clozeOrd int) bool {
	c := n.context(true, clozeOrd)
	c.special["FrontSide"] = ""
	var sb strings.Builder
	c.render(parseTemplate(qfmt), &sb)
	return c.usedNonEmpty
}

// ClozeCards returns the cloze numbers a cloze note's question template
// generates cards for: the deletions in the fields it uses with cloze:, or
// in all fields when it uses none.
func ClozeCards(qfmt string, fields map[string]string) []int {
	names := clozeFields(qfmt)
	var texts []string
	if len(names) == 0 {
		for _, v := range fields {
			texts = append(texts, v)
		}
	}
	for _, name := range names {
		texts = append(texts, fields[name])
	}
	return ClozeNumbers(texts...)
}

// Card is a card ready for display: sanitized HTML with the note type's CSS
// inlined. Background and Color come from the .card rule and are empty when
// the note type leaves them to the app theme. QuestionMedia and AnswerMedia
// are the sounds and videos of each side, which its <cf-sound> and
// <cf-video> placeholders index, and QuestionTTS and AnswerTTS the
// utterances its <cf-tts> placeholders index. TypeAnswer is set when the
// question has a <cf-typeans> placeholder for typing the answer. A side
// with image occlusion shapes shows them in a <cf-occlusion> element (see
// occlusionElement). Fonts are
// the note type's fonts from the media, for its font-family declarations.
type Card struct {
	Question      string      `json:"question"`
	Answer        string      `json:"answer"`
	Background    string      `json:"background,omitempty"`
	Color         string      `json:"color,omitempty"`
	QuestionMedia []MediaRef  `json:"questionMedia,omitempty"`
	AnswerMedia   []MediaRef  `json:"answerMedia,omitempty"`
	QuestionTTS   []Utterance `json:"questionTts,omitempty"`
	AnswerTTS     []Utterance `json:"answerTts,omitempty"`
	TypeAnswer    bool        `json:"typeAnswer,omitempty"`
	Fonts         []FontFace  `json:"fonts,omitempty"`
}

// RenderCard renders both sides of a card for display in env. The answer
// shows the expected text for {{type:…}}. LaTeX blocks show their
// pre-generated images from the media, or their source when it has none.
func RenderCard(t Template, css string, n Note, clozeOrd int, env Env) Card {
	return renderCard(t, css, n, clozeOrd, nil, env)
}

// RenderCardTyped is RenderCard with the answer comparing typed with the
// expected text of {{type:…}}.
func RenderCardTyped(t Template, css string, n Note, clozeOrd int, typed string, env Env) Card {
	return renderCard(t, css, n, clozeOrd, &typed, env)
}

func renderCard(t Template, css string, n Note, clozeOrd int, typed *string, env Env) Card {
	q, qStyles, qMedia, qTTS := sanitize(wrapMath(latexImages(placeOcclusion(Question(t, n, clozeOrd)), n.LatexSVG, n.HasMedia)))
	a, aStyles, aMedia, aTTS := sanitize(wrapMath(latexImages(placeOcclusion(answer(t, n, clozeOrd, typed)), n.LatexSVG, n.HasMedia)))
	root := env.rootClasses(n.CardOrd)
	sheet := parseCSS(css+"\n"+qStyles+"\n"+aStyles, env, root)
	bg, color, rest := cardColors(sheet.cardDecls())
	open := `<div class="` + strings.Join(root, " ") + `">`
	if len(rest) > 0 {
		open = `<div class="` + strings.Join(root, " ") + `" style="` + strings.ReplaceAll(styleString(rest), `"`, "&quot;") + `">`
	}
	if c := texColor(sheet.matching("span", []string{"cloze"})); c != "" {
		q = strings.ReplaceAll(q, clozeColorMacro+"{", `\color{`+c+"}{")
		a = strings.ReplaceAll(a, clozeColorMacro+"{", `\color{`+c+"}{")
	}
	return Card{
		Question:      open + sheet.inline(q) + "</div>",
		Answer:        open + sheet.inline(a) + "</div>",
		Background:    bg,
		Color:         color,
		QuestionMedia: qMedia,
		AnswerMedia:   aMedia,
		QuestionTTS:   qTTS,
		AnswerTTS:     aTTS,
		TypeAnswer:    strings.Contains(q, "<cf-typeans>"),
		Fonts:         sheet.fonts,
	}
}

var texColorValue = regexp.MustCompile(`^(#[0-9a-fA-F]{3}|#[0-9a-fA-F]{6}|[a-zA-Z]+)$`)

// texColor returns the CSS colour among decls when TeX's \color accepts it.
func texColor(decls []decl) string {
	for _, d := range decls {
		if d.prop == "color" && texColorValue.MatchString(strings.TrimSpace(d.value)) {
			return strings.TrimSpace(d.value)
		}
	}
	return ""
}
