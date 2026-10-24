package render

import (
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"script ignored (spec)", "<script>alert(1)</script>Text", "Text"},
		{"sound tag (spec)", "[sound:word.mp3] palabra", `<cf-sound idx="0"></cf-sound> palabra`},
		{"several sound tags", "a[sound:1.mp3]b[sound:x y.ogg]c", `a<cf-sound idx="0"></cf-sound>b<cf-sound idx="1"></cf-sound>c`},
		{"audio and video elements", `x<audio src="a.mp3" controls>fallback</audio>y<video><source src="v.mp4"></video>z`,
			`x<cf-sound idx="0"></cf-sound>y<cf-video idx="1"></cf-video>z`},
		{"nested dropped element", "<video><video>a</video>b</video>c", "c"},
		{"remote and empty sources", `<audio src="https://x/a.mp3"></audio><video></video>[sound:]ok`, "ok"},
		{"event attributes", `<div onclick="evil()" class="a" OnMouseOver='x'>t</div>`, `<div class="a">t</div>`},
		{"javascript URLs", `<a href=" javascript:alert(1)">l</a><img src="cat.jpg">`, `<a>l</a><img src="cat.jpg">`},
		{"image size from the editor kept", `<img src="cat.jpg" width="120"><img src="d.png" style="width: 50%">`,
			`<img src="cat.jpg" width="120"><img src="d.png" style="width: 50%">`},
		{"iframe and object", `<iframe src="https://x"></iframe><object data="x">o</object>ok`, "ok"},
		{"comments and svg", `<!-- c -->a<svg><script>x</script></svg>b`, "ab"},
		{"text to speech", "[anki:tts lang=en_US]hello[/anki:tts]world", `<cf-tts idx="0"></cf-tts>world`},
		{"formatting kept", `<b>b</b><i>i</i><ul><li>l</li></ul><table><tr><td>c</td></tr></table><br/><a href="https://example.org">x</a><span style="color: red">r</span>`,
			`<b>b</b><i>i</i><ul><li>l</li></ul><table><tr><td>c</td></tr></table><br /><a href="https://example.org">x</a><span style="color: red">r</span>`},
		{"math element kept", `<cf-math display="block">a &lt; b</cf-math>`, `<cf-math display="block">a &lt; b</cf-math>`},
		{"answer input placeholder kept", `a<cf-typeans></cf-typeans>`, `a<cf-typeans></cf-typeans>`},
		{"ruby text kept", `<ruby>日本<rp>(</rp><rt>にほん</rt><rp>)</rp></ruby>`, `<ruby>日本<rp>(</rp><rt>にほん</rt><rp>)</rp></ruby>`},
		{"quotes in attribute values", `<span title='say "hi"'>x</span>`, `<span title="say &quot;hi&quot;">x</span>`},
	}
	for _, c := range cases {
		got, _, _, _ := sanitize(c.in)
		if got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}

func TestSanitizeCollectsStyles(t *testing.T) {
	html, styles, _, _ := sanitize(`<style>.x { color: red }</style><span class="x">a</span>`)
	if html != `<span class="x">a</span>` || !strings.Contains(styles, ".x { color: red }") {
		t.Fatalf("html %q styles %q", html, styles)
	}
}

func TestSanitizeMedia(t *testing.T) {
	a := func(f string) MediaRef { return MediaRef{File: f, Kind: Audio} }
	v := func(f string) MediaRef { return MediaRef{File: f, Kind: Video} }
	cases := []struct {
		name, in string
		want     []MediaRef
	}{
		{"sound tag", "[sound:word.mp3] palabra", []MediaRef{a("word.mp3")}},
		{"video by extension (spec)", "[sound:clip.mp4]", []MediaRef{v("clip.mp4")}},
		{"entities in names", "[sound:a&amp;b.mp3]<audio src='c&amp;d.ogg'></audio>", []MediaRef{a("a&b.mp3"), a("c&d.ogg")}},
		{"audio element", `<audio controls src="a.mp3">x</audio>`, []MediaRef{a("a.mp3")}},
		{"self-closing audio", `<audio src="a.mp3"/>b`, []MediaRef{a("a.mp3")}},
		{"source child", `<video controls><source src="v.webm"><source src="v.mp4"></video>`, []MediaRef{v("v.webm")}},
		{"order of appearance", `[sound:1.mp3]<video src="2.mp4"></video><b>[sound:3.ogg]</b>`, []MediaRef{a("1.mp3"), v("2.mp4"), a("3.ogg")}},
		{"unclosed audio", `<audio src="a.mp3">rest`, []MediaRef{a("a.mp3")}},
		{"sound inside script ignored", `<script>[sound:x.mp3]</script>`, nil},
	}
	for _, c := range cases {
		_, _, got, _ := sanitize(c.in)
		if len(got) != len(c.want) {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: got %v want %v", c.name, got, c.want)
				break
			}
		}
	}
}

func TestStripSounds(t *testing.T) {
	in := `<div class=q>hola [sound:hola.mp3]<audio src="a.mp3">f</audio><video><video>x</video></video><img src="cat.jpg"></div>`
	want := `<div class=q>hola <img src="cat.jpg"></div>`
	if got := stripSounds(in); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestCSSInlining(t *testing.T) {
	css := `
.card { font-family: arial; font-size: 20px; text-align: center; color: black; background-color: white; }
b { color: #aa0000 }
.big { font-size: 30px }
span.big { font-weight: bold }
.card .note { font-style: italic }
.cloze { font-weight: bold; color: blue !important; }
.nightMode .cloze, .night_mode b { color: lightblue }
@media (max-width: 600px) { b { color: green } }
a:hover, p::before, p:first-of-type, :is(p), p ~ { color: purple }
i { background: url(evil.png) }
`
	sheet := parseCSS(css, Env{}, []string{"card"})
	cases := []struct{ in, want string }{
		{"<b>x</b>", `<b style="color:#aa0000">x</b>`},
		{`<span class="big">x</span>`, `<span class="big" style="font-size:30px;font-weight:bold">x</span>`},
		{`<div class="big note">x</div>`, `<div class="big note" style="font-size:30px;font-style:italic">x</div>`},
		{`<b style="color: red">x</b>`, `<b style="color:red">x</b>`},
		{`<span class="cloze">[...]</span>`, `<span class="cloze" style="font-weight:bold;color:blue">[...]</span>`},
		{`<i>x</i><p>y</p>`, `<i>x</i><p>y</p>`},
	}
	for _, c := range cases {
		if got := sheet.inline(c.in); got != c.want {
			t.Errorf("inline(%q)\n got %q\nwant %q", c.in, got, c.want)
		}
	}
}

func TestRenderCardColors(t *testing.T) {
	n := Note{Fields: fields("Front", "Q", "Back", "A")}
	cases := []struct {
		name, css, bg, color string
	}{
		// Dark mode default (spec): no colours set, the app theme decides.
		{"no colours", ".card { font-size: 20px; }", "", ""},
		{"stock Anki colours count as unset", ".card { color: black; background-color: white; }", "", ""},
		{"explicit colours", ".card { color: #222; background: #fdf6e3; }", "#fdf6e3", "#222"},
	}
	for _, c := range cases {
		card := RenderCard(basic, c.css, n, 0, Env{})
		if card.Background != c.bg || card.Color != c.color {
			t.Errorf("%s: background %q color %q, want %q %q", c.name, card.Background, card.Color, c.bg, c.color)
		}
	}
}

func TestRenderCardEndToEnd(t *testing.T) {
	tmpl := Template{
		QFmt: "<style>.word{font-size:40px}</style><div class=word>{{Front}}</div>",
		AFmt: "{{FrontSide}}<hr id=answer>{{Back}}",
	}
	n := Note{Fields: fields("Front", "<script>alert(1)</script>hola [sound:hola.mp3]", "Back", `<img src="cat.jpg" onerror="x()">`)}
	card := RenderCard(tmpl, ".card { text-align: center; color: black }", n, 0, Env{})
	wantQ := `<div class="card" style="text-align:center"><div class="word" style="font-size:40px">hola <cf-sound idx="0"></cf-sound></div></div>`
	if card.Question != wantQ {
		t.Errorf("question\n got %q\nwant %q", card.Question, wantQ)
	}
	wantA := `<div class="card" style="text-align:center"><div class="word" style="font-size:40px">hola </div><hr id="answer"><img src="cat.jpg"></div>`
	if card.Answer != wantA {
		t.Errorf("answer\n got %q\nwant %q", card.Answer, wantA)
	}
	// {{FrontSide}} keeps the question's sounds on the question side (spec).
	if len(card.QuestionMedia) != 1 || card.QuestionMedia[0] != (MediaRef{File: "hola.mp3", Kind: Audio}) || len(card.AnswerMedia) != 0 {
		t.Errorf("media q %v a %v", card.QuestionMedia, card.AnswerMedia)
	}

	both := RenderCard(tmpl, "", Note{Fields: fields("Front", "[sound:q.mp3]", "Back", "[sound:clip.mp4] a")}, 0, Env{})
	if want := []MediaRef{{File: "clip.mp4", Kind: Video}}; len(both.AnswerMedia) != 1 || both.AnswerMedia[0] != want[0] ||
		!strings.Contains(both.Answer, `<cf-video idx="0"></cf-video> a`) || strings.Contains(both.Answer, "cf-sound") {
		t.Errorf("answer %q media %v", both.Answer, both.AnswerMedia)
	}

	cloze := RenderCard(clozeTmpl, ".cloze { color: blue }", Note{Fields: fields("Text", "{{c1::Paris}} is in {{c2::France}}", "Back Extra", "")}, 1, Env{})
	if !strings.Contains(cloze.Answer, `<span class="cloze" style="color:blue">Paris</span>`) {
		t.Errorf("cloze answer %q", cloze.Answer)
	}
}

// furigana: output reaches the display as ruby text and the sound tag next
// to it still becomes a sound (spec).
func TestRenderCardFurigana(t *testing.T) {
	tmpl := Template{QFmt: "{{furigana:Expression}}"}
	card := RenderCard(tmpl, "", Note{Fields: fields("Expression", "猫[ねこ] [sound:neko.mp3]")}, 0, Env{})
	want := `<div class="card"><ruby>猫<rt>ねこ</rt></ruby> <cf-sound idx="0"></cf-sound></div>`
	if card.Question != want {
		t.Errorf("question\n got %q\nwant %q", card.Question, want)
	}
	if len(card.QuestionMedia) != 1 || card.QuestionMedia[0] != (MediaRef{File: "neko.mp3", Kind: Audio}) {
		t.Errorf("media %v", card.QuestionMedia)
	}
}

// RTL template (spec): dir="rtl" from the template and from a field
// reaches the display, which lays the text out right to left.
func TestRenderCardKeepsDir(t *testing.T) {
	tmpl := Template{QFmt: `<div dir="rtl">{{Arabic}}</div>`, AFmt: "{{FrontSide}}{{Notes}}"}
	card := RenderCard(tmpl, "", Note{Fields: fields("Arabic", "كتاب", "Notes", `<p dir="RTL">جمع</p>`)}, 0, Env{})
	if want := `<div class="card"><div dir="rtl">كتاب</div></div>`; card.Question != want {
		t.Errorf("question\n got %q\nwant %q", card.Question, want)
	}
	if !strings.Contains(card.Answer, `<p dir="RTL">جمع</p>`) {
		t.Errorf("answer %q", card.Answer)
	}
}

func TestRenderCardTyped(t *testing.T) {
	tmpl := Template{QFmt: "{{Front}} {{type:Back}}", AFmt: "{{FrontSide}}<hr id=answer>{{type:Back}}"}
	n := Note{Fields: fields("Front", "dog", "Back", "perro")}
	card := RenderCardTyped(tmpl, ".typeGood { color: green }", n, 0, "perro", Env{})
	if !card.TypeAnswer || !strings.Contains(card.Question, "<cf-typeans></cf-typeans>") {
		t.Errorf("question %q type answer %v", card.Question, card.TypeAnswer)
	}
	// The question's input is not repeated by {{FrontSide}}; deck CSS for
	// Anki's classes applies to the comparison.
	want := `dog <hr id="answer"><div><code id="typeans"><span class="typeGood" style="color:green">perro</span></code></div>`
	if strings.Contains(card.Answer, "cf-typeans") || !strings.Contains(card.Answer, want) {
		t.Errorf("answer\n got %q\nwant it to contain %q", card.Answer, want)
	}
	if plain := RenderCard(basic, "", n, 0, Env{}); plain.TypeAnswer {
		t.Errorf("a card without type: has TypeAnswer set")
	}
}

// A type: filter only on the question shows its expected text in the
// preview, below the answer line, so the browser's answer column has it.
func TestRenderCardTypedViaFrontSide(t *testing.T) {
	tmpl := Template{QFmt: "{{Front}} {{type:Back}}", AFmt: "{{FrontSide}}<hr id=answer>{{Back}}"}
	n := Note{Fields: fields("Front", "dog", "Back", "perro")}
	preview := RenderCard(tmpl, "", n, 0, Env{})
	if part := strings.TrimSpace(StripHTML(AnswerPart(preview.Answer))); part != "perroperro" {
		t.Errorf("answer part = %q of answer %q", part, preview.Answer)
	}
	typed := RenderCardTyped(tmpl, "", n, 0, "pero", Env{})
	for _, a := range []string{preview.Answer, typed.Answer} {
		if strings.Contains(a, frontTypeMarker) || strings.Contains(a, "cf-typeans") {
			t.Errorf("answer %q contains the marker or the input", a)
		}
	}
	if !strings.Contains(typed.Answer, `<span class="typeMissed">r</span>`) {
		t.Errorf("typed answer %q has no comparison", typed.Answer)
	}
}

func TestSanitizeHTML(t *testing.T) {
	for in, want := range map[string]string{
		"Vocabulary from <b>Book 1</b>":                                          "Vocabulary from <b>Book 1</b>",
		`<a href="https://example.com" onclick="x()">site</a>`:                   `<a href="https://example.com">site</a>`,
		`<a href="javascript:alert(1)">x</a><script>alert(1)</script>`:           `<a>x</a>`,
		`<style>b{color:red}</style>Hi [sound:a.mp3]<audio src="b.mp3"></audio>`: `Hi `,
	} {
		if got := SanitizeHTML(in); got != want {
			t.Errorf("SanitizeHTML(%q) = %q, want %q", in, got, want)
		}
	}
}
