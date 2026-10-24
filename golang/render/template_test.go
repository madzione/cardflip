package render

import (
	"cardflip/golang/internal/usermsg"
	"slices"
	"strings"
	"testing"
)

var basic = Template{QFmt: "{{Front}}", AFmt: "{{FrontSide}}<hr id=answer>{{Back}}"}

func fields(kv ...string) map[string]string {
	m := map[string]string{}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return m
}

// Golden cases for the card-rendering spec: fields, unknown fields,
// FrontSide, sections and the text:, hint: and type: filters.
func TestTemplateGolden(t *testing.T) {
	marker := func(msg string) string { return errorMarker(msg) }
	cases := []struct {
		name     string
		tmpl     Template
		fields   map[string]string
		question string
		answer   string
	}{
		{
			name: "basic substitution (spec)", tmpl: basic,
			fields:   fields("Front", "hola", "Back", "hello"),
			question: "hola", answer: "hola<hr id=answer>hello",
		},
		{
			name: "whitespace inside braces is ignored", tmpl: Template{QFmt: "{{ Front }}", AFmt: "{{  Back}}"},
			fields:   fields("Front", "a", "Back", "b"),
			question: "a", answer: "b",
		},
		{
			name: "field names match exactly", tmpl: Template{QFmt: "{{front}}", AFmt: ""},
			fields:   fields("Front", "a"),
			question: marker("unknown field front"),
		},
		{
			name: "unknown field renders a marker (spec)", tmpl: Template{QFmt: "{{Front}} {{Missing}}", AFmt: "{{Missing}}"},
			fields:   fields("Front", "x"),
			question: "x " + marker("unknown field Missing"), answer: marker("unknown field Missing"),
		},
		{
			name: "answer includes front (spec)", tmpl: basic,
			fields:   fields("Front", "<b>Q</b>", "Back", "A"),
			question: "<b>Q</b>", answer: "<b>Q</b><hr id=answer>A",
		},
		{
			name: "FrontSide is empty on the question", tmpl: Template{QFmt: "[{{FrontSide}}]", AFmt: "{{FrontSide}}"},
			fields:   fields(),
			question: "[]", answer: "[]",
		},
		{
			name: "optional hint omitted (spec)", tmpl: Template{QFmt: "{{Front}}{{#Hint}}Hint: {{Hint}}{{/Hint}}"},
			fields:   fields("Front", "Q", "Hint", ""),
			question: "Q",
		},
		{
			name: "section shown when field has content", tmpl: Template{QFmt: "{{#Hint}}Hint: {{Hint}}{{/Hint}}"},
			fields:   fields("Hint", "starts with p"),
			question: "Hint: starts with p",
		},
		{
			name: "section treats markup and whitespace as empty", tmpl: Template{QFmt: "{{#Hint}}shown{{/Hint}}"},
			fields:   fields("Hint", " <br> <div>&nbsp;</div> "),
			question: "",
		},
		{
			name: "image counts as content", tmpl: Template{QFmt: "{{#Pic}}P{{/Pic}}"},
			fields:   fields("Pic", `<img src="a.png">`),
			question: "P",
		},
		{
			name: "inverted section", tmpl: Template{QFmt: "{{^Hint}}no hint{{/Hint}}|{{^Front}}no front{{/Front}}"},
			fields:   fields("Hint", "", "Front", "x"),
			question: "no hint|",
		},
		{
			name: "nested sections", tmpl: Template{QFmt: "{{#A}}a{{#B}}b{{/B}}{{^B}}!b{{/B}}{{/A}}"},
			fields:   fields("A", "1", "B", ""),
			question: "a!b",
		},
		{
			name: "text filter (spec)", tmpl: Template{QFmt: "{{text:Front}}"},
			fields:   fields("Front", "<b>bold</b>"),
			question: "bold",
		},
		{
			name: "text filter keeps literal angle brackets escaped", tmpl: Template{QFmt: "{{text:Front}}"},
			fields:   fields("Front", "a &lt;b&gt; <i>c</i>"),
			question: "a &lt;b&gt; c",
		},
		{
			name: "hint filter collapses its content", tmpl: Template{QFmt: "{{Front}}{{hint:Extra}}"},
			fields:   fields("Front", "Q", "Extra", "think <i>red</i>"),
			question: `Q<details class="hint"><summary>Extra</summary>think <i>red</i></details>`,
		},
		{
			name: "hint filter on an empty field renders nothing", tmpl: Template{QFmt: "{{hint:Extra}}"},
			fields:   fields("Extra", ""),
			question: "",
		},
		{
			name:     "type filter: input on the question, expected text on the back without typing (spec)",
			tmpl:     Template{QFmt: "{{Front}}{{type:Back}}", AFmt: "{{FrontSide}}<hr id=answer>{{type:Back}}"},
			fields:   fields("Front", "Q", "Back", "dog"),
			question: "Q<cf-typeans></cf-typeans>", answer: "Q<hr id=answer>dog",
		},
		{
			name: "unknown filters are ignored", tmpl: Template{QFmt: "{{unknown:Front}}|{{other:text:Front}}"},
			fields:   fields("Front", "<b>x</b>"),
			question: "<b>x</b>|x",
		},
		{
			name: "furigana filters (spec)", tmpl: Template{QFmt: "{{furigana:Expression}}|{{kana:Expression}}|{{kanji:Expression}}"},
			fields:   fields("Expression", "日本語[にほんご]を 勉強[べんきょう]する"),
			question: "<ruby>日本語<rt>にほんご</rt></ruby>を<ruby>勉強<rt>べんきょう</rt></ruby>する|にほんごをべんきょうする|日本語を勉強する",
		},
		{
			name: "furigana filter chains with text", tmpl: Template{QFmt: "{{furigana:text:Expression}}"},
			fields:   fields("Expression", "<b>猫[ねこ]</b>"),
			question: "<ruby>猫<rt>ねこ</rt></ruby>",
		},
		{
			name: "kanji filter keeps sound tags", tmpl: Template{QFmt: "{{kanji:Expression}}"},
			fields:   fields("Expression", "猫[ねこ] [sound:neko.mp3]"),
			question: "猫 [sound:neko.mp3]",
		},
		{
			name: "filters apply right to left", tmpl: Template{QFmt: "{{text:hint:Front}}"},
			fields:   fields("Front", "x"),
			question: "Frontx",
		},
		{
			name: "text to speech becomes a block", tmpl: Template{QFmt: "{{Front}}{{tts en_US:Front}}"},
			fields:   fields("Front", "hello"),
			question: "hello[anki:tts lang=en_US]hello[/anki:tts]",
		},
		{
			name: "unclosed section is reported, card still renders", tmpl: Template{QFmt: "{{Front}}{{#Hint}}h"},
			fields:   fields("Front", "Q", "Hint", "x"),
			question: "Qh" + marker("{{#Hint}} is not closed"),
		},
		{
			name: "stray closing tag is reported", tmpl: Template{QFmt: "{{Front}}{{/Hint}}"},
			fields:   fields("Front", "Q"),
			question: "Q" + marker("{{/Hint}} has no matching {{#Hint}}"),
		},
		{
			name: "{{ without }} is reported and kept as text", tmpl: Template{QFmt: "{{Front}}<br>{{Back <i>x</i>"},
			fields:   fields("Front", "Q", "Back", "A"),
			question: "Q<br>" + marker("{{Back has no closing }}") + "{{Back <i>x</i>",
		},
		{
			name: "{{ without }} before another tag", tmpl: Template{QFmt: "{{Back} {{Front}}"},
			fields:   fields("Front", "Q", "Back", "A"),
			question: marker("{{Back} has no closing }}") + "{{Back} Q",
		},
		{
			name: "tags without a field name are reported and kept as text", tmpl: Template{QFmt: "{{Front}}{{}}{{cloze:}}{{#}}x{{/}}"},
			fields: fields("Front", "Q"),
			question: "Q" + marker("{{}} has no field name") + "{{}}" + marker("{{cloze:}} has no field name") + "{{cloze:}}" +
				marker("{{#}} has no field name") + "{{#}}x" + marker("{{/}} has no field name") + "{{/}}",
		},
		{
			name: "crossed sections are reported once", tmpl: Template{QFmt: "{{Front}}{{#A}}a{{^B}}b{{/A}}c{{/B}}"},
			fields:   fields("Front", "Q", "A", "x", "B", ""),
			question: "Qab" + marker("{{/A}} closes {{#A}} before {{^B}} is closed") + "c",
		},
		{
			name: "a stray closing tag inside a section leaves it open", tmpl: Template{QFmt: "{{Front}}{{#A}}a{{/B}}b{{/A}}"},
			fields:   fields("Front", "Q", "A", "x"),
			question: "Qa" + marker("{{/B}} has no matching {{#B}}") + "b",
		},
		{
			name: "special fields", tmpl: Template{QFmt: "{{Type}}/{{Deck}}/{{Subdeck}}/{{Card}}/{{Tags}}"},
			fields:   fields(),
			question: "Basic/Lang::Spanish/Spanish/Card 1/verb irregular",
		},
	}
	for _, c := range cases {
		n := Note{Fields: c.fields, Tags: []string{"verb", "irregular"}, NoteType: "Basic", Deck: "Lang::Spanish", CardName: "Card 1"}
		if got := Question(c.tmpl, n, 0); got != c.question {
			t.Errorf("%s: question\n got %q\nwant %q", c.name, got, c.question)
		}
		if c.tmpl.AFmt != "" || c.answer != "" {
			if got := Answer(c.tmpl, n, 0); got != c.answer {
				t.Errorf("%s: answer\n got %q\nwant %q", c.name, got, c.answer)
			}
		}
	}
}

func TestCardFlag(t *testing.T) {
	// Flag as CSS class (spec).
	tmpl := Template{QFmt: `<div class="{{CardFlag}}">{{Front}}</div>`, AFmt: "{{CardFlag}}"}
	n := Note{Fields: fields("Front", "hola"), Flag: 3}
	if got := Question(tmpl, n, 0); got != `<div class="flag3">hola</div>` {
		t.Fatalf("flagged question = %q", got)
	}
	if got := Answer(tmpl, n, 0); got != "flag3" {
		t.Fatalf("flagged answer = %q", got)
	}
	n.Flag = 0
	if got := Question(tmpl, n, 0); got != `<div class="">hola</div>` {
		t.Fatalf("unflagged question = %q", got)
	}
	if got := Answer(tmpl, n, 0); got != "" {
		t.Fatalf("unflagged answer = %q", got)
	}
}

func TestHasContent(t *testing.T) {
	cases := []struct {
		qfmt   string
		fields map[string]string
		want   bool
	}{
		{"{{Front}}", fields("Front", "hola"), true},
		{"{{Front}}", fields("Front", ""), false},
		{"Q: {{Front}}", fields("Front", " <br> "), false},
		{"{{Back}}", fields("Front", "hola", "Back", "hello"), true},
		{"{{#Add Reverse}}{{Back}}{{/Add Reverse}}", fields("Back", "x", "Add Reverse", ""), false},
		{"{{#Add Reverse}}{{Back}}{{/Add Reverse}}", fields("Back", "x", "Add Reverse", "y"), true},
		{"{{Missing}}", fields(), false},
	}
	for _, c := range cases {
		if got := HasContent(c.qfmt, Note{Fields: c.fields}, 0); got != c.want {
			t.Errorf("HasContent(%q, %v) = %v", c.qfmt, c.fields, got)
		}
	}
}

func TestStripHTML(t *testing.T) {
	cases := map[string]string{
		"plain":                          "plain",
		"<b>bold</b> &amp; <i>it</i>":    "bold & it",
		"a<br>b<div>c</div>":             "a b c ",
		"x<script>alert(1)</script>y":    "xy",
		"<style>.a{}</style>z<!-- c -->": "z",
		"1 &lt; 2&nbsp;ok":               "1 < 2 ok",
		"a < b":                          "a < b",
	}
	for in, want := range cases {
		if got := StripHTML(in); strings.TrimSpace(got) != strings.TrimSpace(want) {
			t.Errorf("StripHTML(%q) = %q, want %q", in, got, want)
		}
	}
}

// Cloze-number conditionals (card-rendering spec): {{#cN}} and {{^cN}} test
// the cloze card being rendered, on both sides.
func TestClozeNumberConditionals(t *testing.T) {
	text := "{{c1::Paris}} is in {{c2::France}} {{c3::Europe}}"
	cases := []struct {
		name   string
		tmpl   Template
		fields map[string]string
		want   []string // question per cloze card 1..3
	}{
		{
			name:   "content only on card 1 (spec)",
			tmpl:   Template{QFmt: "{{cloze:Text}}{{#c1}} (capital){{/c1}}"},
			fields: fields("Text", text),
			want: []string{
				renderCloze(text, 1, true) + " (capital)",
				renderCloze(text, 2, true),
				renderCloze(text, 3, true),
			},
		},
		{
			name:   "inverted condition (spec)",
			tmpl:   Template{QFmt: "{{^c2}}Not card 2{{/c2}}"},
			fields: fields("Text", text),
			want:   []string{"Not card 2", "", "Not card 2"},
		},
		{
			name:   "field named c1 wins (spec)",
			tmpl:   Template{QFmt: "{{#c1}}X{{/c1}}"},
			fields: fields("Text", text, "c1", ""),
			want:   []string{"", "", ""},
		},
		{
			name:   "nested in a field section",
			tmpl:   Template{QFmt: "{{#Extra}}{{#c3}}{{Extra}}{{/c3}}{{/Extra}}"},
			fields: fields("Text", text, "Extra", "e"),
			want:   []string{"", "", "e"},
		},
		{
			name:   "not a cloze number",
			tmpl:   Template{QFmt: "{{#c0}}a{{/c0}}{{#c01}}b{{/c01}}{{#C1}}c{{/C1}}"},
			fields: fields("Text", text),
			want: []string{
				errorMarker("unknown field c0") + errorMarker("unknown field c01") + errorMarker("unknown field C1"),
				errorMarker("unknown field c0") + errorMarker("unknown field c01") + errorMarker("unknown field C1"),
				errorMarker("unknown field c0") + errorMarker("unknown field c01") + errorMarker("unknown field C1"),
			},
		},
	}
	for _, c := range cases {
		for i, want := range c.want {
			if got := Question(c.tmpl, Note{Fields: c.fields}, i+1); got != want {
				t.Errorf("%s: card %d question\n got %q\nwant %q", c.name, i+1, got, want)
			}
		}
	}

	// The answer side tests the card too, directly and through FrontSide.
	tmpl := Template{QFmt: "{{#c1}}q1{{/c1}}", AFmt: "{{FrontSide}}|{{#c1}}a1{{/c1}}{{^c1}}a-{{/c1}}"}
	n := Note{Fields: fields("Text", text)}
	if got := Answer(tmpl, n, 1); got != "q1|a1" {
		t.Errorf("card 1 answer = %q", got)
	}
	if got := Answer(tmpl, n, 2); got != "|a-" {
		t.Errorf("card 2 answer = %q", got)
	}

	// Standard note types keep the field lookup: no field c1 is unknown.
	if got := Question(Template{QFmt: "{{#c1}}X{{/c1}}"}, Note{Fields: fields("Front", "f")}, 0); got != errorMarker("unknown field c1") {
		t.Errorf("standard note type: %q", got)
	}
}

// A cloze-number section is not a used field: it neither creates cards nor
// counts as content.
func TestClozeNumberConditionalsCardGeneration(t *testing.T) {
	qfmt := "{{#c1}}x{{/c1}}{{cloze:Text}}"
	f := fields("Text", "{{c2::Paris}} is in {{c3::France}}")
	if got := ClozeCards(qfmt, f); len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Errorf("cards = %v", got)
	}
	if HasContent("{{#c1}}x{{/c1}}", Note{Fields: f}, 1) {
		t.Errorf("cloze-number section counted as content")
	}
}

// Cases for the type-in-the-answer requirements: the input on the question,
// the comparison on the answer, and the cloze and nc: variants.
func TestTypeAnswer(t *testing.T) {
	const input = "<cf-typeans></cf-typeans>"
	good := func(s string) string {
		return `<div><code id="typeans"><span class="typeGood">` + s + `</span></code></div>`
	}
	typeBack := Template{QFmt: "{{Front}}{{type:Back}}", AFmt: "{{FrontSide}}<hr id=answer>{{type:Back}}"}
	typeFront := Template{QFmt: "{{Front}} {{type:Back}}", AFmt: "{{FrontSide}}<hr id=answer>{{Back}}"}
	cases := []struct {
		name     string
		tmpl     Template
		fields   map[string]string
		clozeOrd int
		typed    string
		preview  bool // render with Answer, without typed text
		question string
		answer   string
	}{
		{
			name: "exact match (spec)", tmpl: typeBack, fields: fields("Front", "dog", "Back", "perro"), typed: "perro",
			question: "dog" + input, answer: "dog<hr id=answer>" + good("perro"),
		},
		{
			name: "typo (spec)", tmpl: typeBack, fields: fields("Front", "dog", "Back", "perro"), typed: "pero",
			question: "dog" + input,
			answer: `dog<hr id=answer><div><code id="typeans"><span class="typeGood">pero</span><br><span id="typearrow">&darr;</span><br>` +
				`<span class="typeGood">per</span><span class="typeMissed">r</span><span class="typeGood">o</span></code></div>`,
		},
		{
			name: "nothing typed shows the expected text", tmpl: typeBack, fields: fields("Front", "dog", "Back", "<b>perro</b>"),
			question: "dog" + input, answer: `dog<hr id=answer><div><code id="typeans">perro</code></div>`,
		},
		{
			name:   "only the first type: filter is active",
			tmpl:   Template{QFmt: "{{type:Back}}|{{type:Extra}}", AFmt: "{{type:Back}}|{{type:Extra}}"},
			fields: fields("Back", "perro", "Extra", "x"), typed: "perro",
			question: input + "|", answer: good("perro") + "|",
		},
		{
			name: "input is shown for an empty field", tmpl: typeBack, fields: fields("Front", "dog", "Back", ""),
			question: "dog" + input, answer: `dog<hr id=answer><div><code id="typeans"></code></div>`,
		},
		{
			name:   "cloze answer (spec)",
			tmpl:   Template{QFmt: "{{cloze:Text}}{{type:cloze:Text}}", AFmt: "{{cloze:Text}}<br>{{type:cloze:Text}}"},
			fields: fields("Text", "{{c1::Paris}} is in {{c2::France}}"), clozeOrd: 1, typed: "Paris",
			question: `<span class="cloze">[...]</span> is in France` + input,
			answer:   `<span class="cloze">Paris</span> is in France<br>` + good("Paris"),
		},
		{
			name:   "several deletions of the cloze number are joined",
			tmpl:   Template{QFmt: "{{cloze:Text}}{{type:cloze:Text}}", AFmt: "{{type:cloze:Text}}"},
			fields: fields("Text", "{{c1::rojo::color}}, {{c2::verde}} y {{c1::azul}}"), clozeOrd: 1, typed: "rojo, azul",
			question: `<span class="cloze">[color]</span>, verde y <span class="cloze">[...]</span>` + input,
			answer:   good("rojo, azul"),
		},
		{
			name: "ignore accents (spec)", tmpl: Template{QFmt: "{{type:nc:Back}}", AFmt: "{{type:nc:Back}}"},
			fields: fields("Back", "über"), typed: "uber",
			question: input, answer: good("über"),
		},
		{
			name: "type filter only on the question (spec)", tmpl: typeFront, fields: fields("Front", "dog", "Back", "perro"), typed: "pero",
			question: "dog " + input,
			answer: `dog <hr id=answer><div><code id="typeans"><span class="typeGood">pero</span><br><span id="typearrow">&darr;</span><br>` +
				`<span class="typeGood">per</span><span class="typeMissed">r</span><span class="typeGood">o</span></code></div>perro`,
		},
		{
			name: "question-only type filter keeps its modifiers (spec)",
			tmpl: Template{QFmt: "{{Front}} {{type:nc:Back}}", AFmt: typeFront.AFmt}, fields: fields("Front", "dog", "Back", "über"), typed: "uber",
			question: "dog " + input, answer: "dog <hr id=answer>" + good("über") + "über",
		},
		{
			name:   "question-only type:cloze: filter",
			tmpl:   Template{QFmt: "{{cloze:Text}}{{type:cloze:Text}}", AFmt: "{{FrontSide}}<hr id=answer>{{Extra}}"},
			fields: fields("Text", "{{c1::Paris}} is in {{c2::France}}", "Extra", "capital"), clozeOrd: 1, typed: "Paris",
			question: `<span class="cloze">[...]</span> is in France` + input,
			answer:   `<span class="cloze">[...]</span> is in France<hr id=answer>` + good("Paris") + "capital",
		},
		{
			name: "question-only type filter in a preview (spec)", tmpl: typeFront, fields: fields("Front", "dog", "Back", "<b>perro</b>"), preview: true,
			question: "dog " + input, answer: "dog <hr id=answer><b>perro</b><b>perro</b>",
		},
		{
			name: "question-only type filter without an answer line", tmpl: Template{QFmt: typeFront.QFmt, AFmt: "{{FrontSide}}<br>{{Back}}"},
			fields: fields("Front", "dog", "Back", "perro"), typed: "perro",
			question: "dog " + input, answer: "dog " + good("perro") + "<br>perro",
		},
		{
			name: "answer line before {{FrontSide}} moves too", tmpl: Template{QFmt: typeFront.QFmt, AFmt: "<hr id=answer>{{FrontSide}}|{{Back}}"},
			fields: fields("Front", "dog", "Back", "perro"), typed: "perro",
			question: "dog " + input, answer: "dog <hr id=answer>" + good("perro") + "|perro",
		},
		{
			name: "{{FrontSide}} used twice shows the comparison once", tmpl: Template{QFmt: typeFront.QFmt, AFmt: "{{FrontSide}}<hr id=answer>{{FrontSide}}"},
			fields: fields("Front", "dog", "Back", "perro"), typed: "perro",
			question: "dog " + input, answer: "dog <hr id=answer>" + good("perro") + "dog ",
		},
		{
			name: "answer without {{FrontSide}} ignores the question's filter", tmpl: Template{QFmt: typeFront.QFmt, AFmt: "{{Back}}"},
			fields: fields("Front", "dog", "Back", "perro"), typed: "pero",
			question: "dog " + input, answer: "perro",
		},
	}
	for _, c := range cases {
		n := Note{Fields: c.fields}
		if got := Question(c.tmpl, n, c.clozeOrd); got != c.question {
			t.Errorf("%s: question\n got %q\nwant %q", c.name, got, c.question)
		}
		answer := AnswerTyped(c.tmpl, n, c.clozeOrd, c.typed)
		if c.preview {
			answer = Answer(c.tmpl, n, c.clozeOrd)
		}
		if got := answer; got != c.answer {
			t.Errorf("%s: answer\n got %q\nwant %q", c.name, got, c.answer)
		}
	}
}

// Without typed text (the browser preview) type:cloze: shows the answers.
func TestTypeAnswerClozePreview(t *testing.T) {
	tmpl := Template{QFmt: "{{type:cloze:Text}}", AFmt: "{{type:cloze:Text}}"}
	n := Note{Fields: fields("Text", "{{c1::Paris}} is in {{c2::France}}")}
	if got := Answer(tmpl, n, 2); got != "France" {
		t.Errorf("answer = %q", got)
	}
}

func TestCheckTemplate(t *testing.T) {
	basicFields := []string{"Front", "Back"}
	cases := []struct {
		name       string
		qfmt, afmt string
		fields     []string
		cloze      bool
		want       []Problem
	}{
		{
			name: "unknown field (spec)", qfmt: "{{Frnt}}", afmt: "{{Back}}", fields: basicFields,
			want: []Problem{{Message: `The front template uses {{Frnt}}, but there is no field "Frnt".`}},
		},
		{
			name: "unknown field on the back, with filter and section", qfmt: "{{Front}}",
			afmt: "{{#Extra}}{{hint:Extra}}{{/Extra}}{{text:Bak}}", fields: basicFields,
			want: []Problem{
				{Back: true, Message: `The back template uses {{#Extra}}, but there is no field "Extra".`},
				{Back: true, Message: `The back template uses {{text:Bak}}, but there is no field "Bak".`},
			},
		},
		{
			name: "unclosed section (spec)", qfmt: "{{Front}}{{#Hint}}{{Hint}}", afmt: "", fields: []string{"Front", "Hint"},
			want: []Problem{{Message: "In the front template, {{#Hint}} is not closed."}},
		},
		{
			name: "closing tag without opening", qfmt: "{{Front}}", afmt: "{{Back}}{{/Back}}", fields: basicFields,
			want: []Problem{{Back: true, Message: "In the back template, {{/Back}} has no matching {{#Back}}."}},
		},
		{
			name: "front without fields", qfmt: "Question: {{Deck}}", afmt: "{{Back}}", fields: basicFields,
			want: []Problem{{Message: "The front template uses no field, so it would never produce a card."}},
		},
		{
			name: "missing cloze filter (spec)", qfmt: "{{Text}}", afmt: "{{Text}}", fields: []string{"Text", "Back Extra"}, cloze: true,
			want: []Problem{{Message: "The front template must use {{cloze:Text}} or another cloze field."}},
		},
		{
			name: "special fields are allowed", qfmt: "{{Front}} {{Tags}} {{Type}} {{Deck}} {{Subdeck}} {{Card}} {{CardFlag}}",
			afmt: "{{FrontSide}}{{#Tags}}{{Tags}}{{/Tags}}", fields: basicFields,
		},
		{
			name: "a field named like a special field", qfmt: "{{Type}}", afmt: "{{Type}}", fields: []string{"Type"},
		},
		{
			name: "cloze number sections in cloze note types", qfmt: "{{cloze:Text}}{{#c1}}one{{/c1}}{{^c2}}x{{/c2}}",
			afmt: "{{cloze:Text}}", fields: []string{"Text"}, cloze: true,
		},
		{
			name: "cloze number sections in standard note types", qfmt: "{{Front}}{{#c1}}one{{/c1}}", afmt: "", fields: basicFields,
			want: []Problem{{Message: `The front template uses {{#c1}}, but there is no field "c1".`}},
		},
		{
			name: "type-in cloze counts as cloze filter", qfmt: "{{type:cloze:Text}}", afmt: "", fields: []string{"Text"}, cloze: true,
		},
		{
			name: "{{ without }}", qfmt: "{{Front}}", afmt: "{{FrontSide}}<hr>{{Back", fields: basicFields,
			want: []Problem{{Back: true, Message: "In the back template, {{Back has no closing }}."}},
		},
		{
			name: "{{ without }} alone on the front", qfmt: "{{Fro", afmt: "", fields: basicFields,
			want: []Problem{{Message: "In the front template, {{Fro has no closing }}."}},
		},
		{
			name: "long {{ without }} is shortened", qfmt: "{{Front}}{{" + strings.Repeat("x", 40), afmt: "", fields: basicFields,
			want: []Problem{{Message: "In the front template, {{" + strings.Repeat("x", 30) + "… has no closing }}."}},
		},
		{
			name: "empty tag", qfmt: "{{Front}}{{ }}", afmt: "{{hint:}}", fields: basicFields,
			want: []Problem{
				{Message: "In the front template, {{}} has no field name."},
				{Back: true, Message: "In the back template, {{hint:}} has no field name."},
			},
		},
		{
			name: "crossed sections", qfmt: "{{Front}}{{#Front}}{{#Back}}{{/Front}}{{/Back}}", afmt: "", fields: basicFields,
			want: []Problem{{Message: "In the front template, {{/Front}} closes {{#Front}} before {{#Back}} is closed."}},
		},
		{
			name: "each unknown field is reported once", qfmt: "{{Frnt}}{{text:Frnt}}", afmt: "", fields: basicFields,
			want: []Problem{{Message: `The front template uses {{Frnt}}, but there is no field "Frnt".`}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CheckTemplate(c.qfmt, c.afmt, c.fields, c.cloze)
			for i := range got {
				got[i].Message = usermsg.Plain(got[i].Message)
			}
			if !slices.Equal(got, c.want) {
				t.Errorf("CheckTemplate(%q, %q) =\n%v\nwant\n%v", c.qfmt, c.afmt, got, c.want)
			}
		})
	}
}

// The built-in and Anki's stock templates pass the checks.
// Rewriting a template for a renamed field keeps broken tags as they are.
func TestRenameFieldKeepsBrokenTags(t *testing.T) {
	got := RenameField("{{Front}}{{}} {{Back", "Front", "Word")
	if want := "{{Word}}{{}} {{Back"; got != want {
		t.Errorf("RenameField = %q, want %q", got, want)
	}
}

func TestCheckTemplateBuiltins(t *testing.T) {
	fb := []string{"Front", "Back"}
	cases := []struct {
		qfmt, afmt string
		fields     []string
		cloze      bool
	}{
		{"{{Front}}", "{{FrontSide}}\n\n<hr id=answer>\n\n{{Back}}", fb, false},
		{"{{Back}}", "{{FrontSide}}\n\n<hr id=answer>\n\n{{Front}}", fb, false},
		{"{{#Add Reverse}}{{Back}}{{/Add Reverse}}", "{{FrontSide}}\n\n<hr id=answer>\n\n{{Front}}", []string{"Front", "Back", "Add Reverse"}, false},
		{"{{Front}}\n\n{{type:Back}}", "{{Front}}\n\n<hr id=answer>\n\n{{type:Back}}", fb, false},
		{"{{cloze:Text}}", "{{cloze:Text}}<br>\n{{Back Extra}}", []string{"Text", "Back Extra"}, true},
	}
	for _, c := range cases {
		if got := CheckTemplate(c.qfmt, c.afmt, c.fields, c.cloze); got != nil {
			t.Errorf("CheckTemplate(%q, %q) = %v, want none", c.qfmt, c.afmt, got)
		}
	}
}
