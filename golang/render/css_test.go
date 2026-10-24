package render

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var (
	desktop      = Env{Platform: "linux", WidthPx: 800, HeightPx: 600}
	desktopNight = Env{Night: true, Platform: "linux", WidthPx: 800, HeightPx: 600}
	phone        = Env{Platform: "android mobile", WidthPx: 400, HeightPx: 800}
)

// renderQ renders the question of card type ord of a Basic note "<b>x</b>"
// with a cloze span, and returns its root element and contents.
func renderQ(css string, ord int, env Env) Card {
	tmpl := Template{QFmt: `{{Front}} <span class="cloze">c</span>`}
	return RenderCard(tmpl, css, Note{Fields: fields("Front", "<b>x</b>"), CardOrd: ord}, 0, env)
}

// Card root classes (card-rendering spec).
func TestRootClassSelectors(t *testing.T) {
	cases := []struct {
		name string
		css  string
		ord  int
		env  Env
		want string // a substring of the question
		not  string // must not appear
	}{
		{"card type class on the reverse card", ".card2 { font-size: 30px }", 2, desktop,
			`<div class="card card2 linux" style="font-size:30px">`, ""},
		{"card type class not on the forward card", ".card2 { font-size: 30px }", 1, desktop,
			`<div class="card card1 linux">`, "30px"},
		{"card type class as ancestor", ".card2 b { color: red }", 2, desktop, `<b style="color:red">`, ""},
		{"compound with night mode", ".card { font-size: 20px } .card.nightMode { font-size: 22px }", 1, desktopNight,
			`<div class="card card1 nightMode night_mode linux" style="font-size:22px">`, ""},
		{"compound without night mode", ".card { font-size: 20px } .card.nightMode { font-size: 22px }", 1, desktop,
			`style="font-size:20px"`, "22px"},
		{"night mode ancestor beats the plain rule", ".nightMode .cloze { color: lightblue } .cloze { color: blue }", 1, desktopNight,
			`<span class="cloze" style="color:lightblue">`, ""},
		{"night mode ancestor in light mode", ".nightMode .cloze { color: lightblue } .cloze { color: blue }", 1, desktop,
			`<span class="cloze" style="color:blue">`, "lightblue"},
		{"night_mode spelling", ".night_mode b { color: white }", 1, desktopNight, `<b style="color:white">`, ""},
		{"phone rule on a phone", ".mobile .card { font-size: 30px }", 1, phone,
			`<div class="card card1 android mobile" style="font-size:30px">`, ""},
		{"phone rule descendant", ".mobile .card b { color: green }", 1, phone, `<b style="color:green">`, ""},
		{"other platform does not match", ".mobile .card { font-size: 30px } .win b { color: red }", 1, desktop,
			`<div class="card card1 linux">`, "30px"},
		{"unknown platform class is dropped", "", 1, Env{Platform: `linux x"y`},
			`<div class="card card1 linux">`, `x"y`},
	}
	for _, c := range cases {
		q := renderQ(c.css, c.ord, c.env).Question
		if !strings.Contains(q, c.want) || (c.not != "" && strings.Contains(q, c.not)) {
			t.Errorf("%s:\n got %q\nwant it to contain %q", c.name, q, c.want)
		}
	}
}

func TestNightModeCardColors(t *testing.T) {
	css := ".card { color: black; background-color: white } .nightMode.card { color: #eee; background-color: #222 }"
	if c := renderQ(css, 1, desktopNight); c.Background != "#222" || c.Color != "#eee" {
		t.Errorf("night: background %q color %q", c.Background, c.Color)
	}
	if c := renderQ(css, 1, desktop); c.Background != "" || c.Color != "" {
		t.Errorf("light: background %q color %q", c.Background, c.Color)
	}
}

// Environment media queries (card-rendering spec).
func TestMediaQueries(t *testing.T) {
	cases := []struct {
		name  string
		query string
		env   Env
		match bool
	}{
		{"dark scheme in dark mode", "(prefers-color-scheme: dark)", desktopNight, true},
		{"dark scheme in light mode", "(prefers-color-scheme: dark)", desktop, false},
		{"light scheme in light mode", "(prefers-color-scheme: light)", desktop, true},
		{"max-width above the width", "(max-width: 600px)", phone, true},
		{"max-width below the width", "(max-width: 600px)", desktop, false},
		{"max-width equal to the width", "(max-width: 400px)", phone, true},
		{"min-width in em", "(min-width: 40em)", desktop, true},
		{"unknown width", "(max-width: 600px)", Env{}, false},
		{"portrait", "(orientation: portrait)", phone, true},
		{"landscape", "(orientation: landscape)", phone, false},
		{"landscape on desktop", "screen and (orientation: landscape)", desktop, true},
		{"comma list, second matches", "(max-width: 600px), (prefers-color-scheme: dark)", desktopNight, true},
		{"comma list, none matches", "(max-width: 600px), (prefers-color-scheme: dark)", desktop, false},
		{"only screen and", "only screen and (max-width: 480px)", phone, true},
		{"all", "all", desktop, true},
		{"print", "print", desktop, false},
		{"not print", "not print", desktop, true},
		{"not with a feature", "not screen and (max-width: 600px)", desktop, true},
		{"unknown feature", "(hover: hover)", desktop, false},
		{"unknown feature after not", "not all and (pointer: fine)", desktop, false},
		{"unknown feature in an and", "(max-width: 600px) and (min-resolution: 2dppx)", phone, false},
		{"unknown media type", "tv", desktop, false},
		{"or is not supported", "(max-width: 600px) or (min-width: 100px)", phone, false},
	}
	for _, c := range cases {
		css := "@media " + c.query + " { b { color: red } }"
		q := renderQ(css, 1, c.env).Question
		if got := strings.Contains(q, "color:red"); got != c.match {
			t.Errorf("%s: %q in %+v matched %v, want %v", c.name, c.query, c.env, got, c.match)
		}
	}
}

func TestMediaBlocks(t *testing.T) {
	css := `
b { color: black }
@media (prefers-color-scheme: dark) {
  .card { background: black }
  b { color: white }
  @media (max-width: 500px) { b { font-size: 30px } }
}
@supports (display: grid) { b { color: purple } }
@font-face { font-family: x; src: local(x) }
b { font-weight: bold }
`
	dark := renderQ(css, 1, Env{Night: true, WidthPx: 400, HeightPx: 800})
	if dark.Background != "black" || !strings.Contains(dark.Question, `<b style="color:white;font-size:30px;font-weight:bold">`) {
		t.Errorf("dark: %q background %q", dark.Question, dark.Background)
	}
	light := renderQ(css, 1, Env{WidthPx: 400, HeightPx: 800})
	if light.Background != "" || !strings.Contains(light.Question, `<b style="color:black;font-weight:bold">`) {
		t.Errorf("light: %q background %q", light.Question, light.Background)
	}
}

// Fonts from media (card-rendering spec).
func TestFontFaces(t *testing.T) {
	cases := []struct {
		name string
		css  string
		want []FontFace
	}{
		{"TrueType file", `@font-face { font-family: jp; src: url("_jp.ttf") }`,
			[]FontFace{{Family: "jp", File: "_jp.ttf"}}},
		{"OpenType file, unquoted url, quoted family", `@font-face { font-family: 'My Font'; src: url(_my_font.OTF) }`,
			[]FontFace{{Family: "My Font", File: "_my_font.OTF"}}},
		{"WOFF2 ignored", `@font-face { font-family: jp; src: url("_jp.woff2") format("woff2") }`, nil},
		{"remote URL ignored", `@font-face { font-family: jp; src: url(https://example.com/jp.ttf) }`, nil},
		{"local() ignored", `@font-face { font-family: jp; src: local(Arial) }`, nil},
		{"first local TrueType source", `@font-face { font-family: jp; src: local(jp), url(_jp.woff2) format("woff2"), url("//cdn/jp.ttf"), url('_jp.ttf') format("truetype") }`,
			[]FontFace{{Family: "jp", File: "_jp.ttf"}}},
		{"several faces for one family", `
@font-face { font-family: "Noto"; src: url(_noto.ttf) }
@font-face { font-family: "Noto"; src: url(_noto-bold.ttf); font-weight: Bold; font-style: normal }
@font-face { font-family: "Noto"; src: url(_noto.ttf) }`,
			[]FontFace{{Family: "Noto", File: "_noto.ttf"}, {Family: "Noto", File: "_noto-bold.ttf", Weight: "bold", Style: "normal"}}},
		{"in a matching @media block", `@media (prefers-color-scheme: light) { @font-face { font-family: jp; src: url(_jp.ttf) } }`,
			[]FontFace{{Family: "jp", File: "_jp.ttf"}}},
		{"in another @media block", `@media print { @font-face { font-family: jp; src: url(_jp.ttf) } }`, nil},
		{"without a family", `@font-face { src: url(_jp.ttf) }`, nil},
	}
	for _, c := range cases {
		got := renderQ(c.css, 1, desktop).Fonts
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s:\n got %+v\nwant %+v", c.name, got, c.want)
		}
	}
}

func TestFontFamilyStillInlined(t *testing.T) {
	css := `@font-face { font-family: jp; src: url("_jp.ttf") } .card { font-family: jp, serif } b { font-family: "My Font" }`
	c := renderQ(css, 1, desktop)
	if !strings.Contains(c.Question, `<div class="card card1 linux" style="font-family:jp,serif">`) ||
		!strings.Contains(c.Question, `<b style="font-family:&quot;My Font&quot;">`) {
		t.Errorf("got %q", c.Question)
	}
}

// inlineIn inlines css into html, the content of a card with the root
// class "card".
func inlineIn(css, html string) string {
	sheet := parseCSS(css, desktop, []string{"card", "linux"})
	return sheet.inline(html)
}

// CSS selectors (card-rendering spec).
func TestSelectors(t *testing.T) {
	cases := []struct {
		name, css, html, want string
	}{
		{"descendant", ".front b { color: green }",
			`<div class="front"><i><b>x</b></i></div><b>y</b>`,
			`<div class="front"><i><b style="color:green">x</b></i></div><b>y</b>`},
		{"child", "div > b { color: green }",
			`<div><b>a</b><i><b>b</b></i></div><b>c</b>`,
			`<div><b style="color:green">a</b><i><b>b</b></i></div><b>c</b>`},
		{"adjacent sibling", "#answer + div { display: none }",
			`<div>q</div><hr id="answer"><div>a</div><div>b</div>`,
			`<div>q</div><hr id="answer"><div style="display:none">a</div><div>b</div>`},
		{"adjacent sibling skips text", "b + i { color: green }",
			`<b>a</b> and <i>b</i>`,
			`<b>a</b> and <i style="color:green">b</i>`},
		{"general sibling", "hr ~ p { color: green }",
			`<p>a</p><hr><span>s</span><p>b</p>`,
			`<p>a</p><hr><span>s</span><p style="color:green">b</p>`},
		{"id", "#x { color: green }",
			`<b id="x">a</b><b id="xy">b</b>`,
			`<b id="x" style="color:green">a</b><b id="xy">b</b>`},
		{"element, id and class", "b#x.y { color: green }",
			`<b id="x" class="y">a</b><i id="x" class="y">b</i><b id="x">c</b>`,
			`<b id="x" class="y" style="color:green">a</b><i id="x" class="y">b</i><b id="x">c</b>`},
		{"attribute present", "[title] { color: green }",
			`<b title="">a</b><b>b</b>`,
			`<b title style="color:green">a</b><b>b</b>`},
		{"attribute value", `span[lang=ja] { color: green } [data-x="a b"] { color: blue }`,
			`<span lang="ja">a</span><span lang="jap">b</span><i data-x="a b">c</i>`,
			`<span lang="ja" style="color:green">a</span><span lang="jap">b</span><i data-x="a b" style="color:blue">c</i>`},
		{"universal", "* { color: green }",
			`<b>a</b>`,
			`<b style="color:green">a</b>`},
		{"universal with a class and in a chain", "*.y { color: green } div > * { color: blue }",
			`<i class="y">a</i><div><b>b</b></div>`,
			`<i class="y" style="color:green">a</i><div><b style="color:blue">b</b></div>`},
		{"first and last child", "b:first-child { color: green } b:last-child { color: blue }",
			`<div>t<b>1</b>t<b>2</b><b>3</b>t</div>`,
			`<div>t<b style="color:green">1</b>t<b>2</b><b style="color:blue">3</b>t</div>`},
		{"only child", "b:only-child { color: green }",
			`<div><b>1</b></div><div><b>2</b><i>3</i></div>`,
			`<div><b style="color:green">1</b></div><div><b>2</b><i>3</i></div>`},
		{"nth-child number", "li:nth-child(2) { color: green }",
			`<ul><li>1</li><li>2</li><li>3</li></ul>`,
			`<ul><li>1</li><li style="color:green">2</li><li>3</li></ul>`},
		{"nth-child odd", "li:nth-child(odd) { color: green }",
			`<ul><li>1</li><li>2</li><li>3</li></ul>`,
			`<ul><li style="color:green">1</li><li>2</li><li style="color:green">3</li></ul>`},
		{"nth-child an+b", "li:nth-child(2n + 2) { color: green } li:nth-child(-n+1) { color: blue }",
			`<ul><li>1</li><li>2</li><li>3</li><li>4</li></ul>`,
			`<ul><li style="color:blue">1</li><li style="color:green">2</li><li>3</li><li style="color:green">4</li></ul>`},
		{"not", "b:not(.x) { color: green } i:not([title]) { color: blue }",
			`<b class="x">1</b><b>2</b><i title="t">3</i><i>4</i>`,
			`<b class="x">1</b><b style="color:green">2</b><i title="t">3</i><i style="color:blue">4</i>`},
		{"selector list", "i, .y { color: green }",
			`<i>1</i><b class="y">2</b><b>3</b>`,
			`<i style="color:green">1</i><b class="y" style="color:green">2</b><b>3</b>`},
		{"root as body with the root classes", ".card > b, body > i { color: green } div > u { color: blue }",
			`<b>1</b><i>2</i><u>3</u>`,
			`<b style="color:green">1</b><i style="color:green">2</i><u>3</u>`},
		{"implied end of li", "li + li { color: green }",
			`<ul><li>1<li>2</ul>`,
			`<ul><li>1<li style="color:green">2</ul>`},
		{"a block ends an open p", "p > div { color: green } p + div { color: blue }",
			`<p>1<div>2</div>`,
			`<p>1<div style="color:blue">2</div>`},
		{"void elements have no children", "br > b { color: green } br + b { color: blue }",
			`<br><b>1</b>`,
			`<br><b style="color:blue">1</b>`},
		{"stray end tags are ignored", "div > b { color: green }",
			`<div></i></span><b>1</b></div>`,
			`<div></i></span><b style="color:green">1</b></div>`},
	}
	for _, c := range cases {
		if got := inlineIn(c.css, c.html); got != c.want {
			t.Errorf("%s:\n got %s\nwant %s", c.name, got, c.want)
		}
	}
}

// Rules with unsupported selectors are ignored (card-rendering spec).
func TestUnsupportedSelectorsIgnored(t *testing.T) {
	for _, sel := range []string{
		"b:hover", "b::before", "b:before", "b:first-of-type", ":is(b)", "b:has(i)",
		"[title~=x]", "[title^=x]", "[title=x i]", "b:not(.x.y)", "b:not(:not(.x))", "b:not(i b)",
		"b:nth-child(2n of .x)", "b:nth-child(x)", "ns|b", "> b", "b >", "b + + b", "b.", "b[title",
	} {
		html := `<b class="z" title="x">1</b>`
		if got := inlineIn(sel+" { color: red }", html); got != html {
			t.Errorf("%s matched: %s", sel, got)
		}
	}
	// The supported parts of a list still apply.
	if got := inlineIn("b:hover, i, p::after { color: red }", `<b>1</b><i>2</i>`); got != `<b>1</b><i style="color:red">2</i>` {
		t.Errorf("list: %s", got)
	}
}

func TestSelectorScenarios(t *testing.T) {
	// Descendant selector, through a field.
	tmpl := Template{QFmt: "{{Front}}", AFmt: "{{FrontSide}}<hr id=answer><div>Hidden</div>"}
	n := Note{Fields: fields("Front", `<div class="front"><b>x</b></div>`)}
	card := RenderCard(tmpl, ".front b { color: green } #answer + div { display: none }", n, 0, desktop)
	if !strings.Contains(card.Question, `<b style="color:green">x</b>`) {
		t.Errorf("descendant: %s", card.Question)
	}
	// Id and sibling.
	if !strings.Contains(card.Answer, `<hr id="answer"><div style="display:none">Hidden</div>`) {
		t.Errorf("id and sibling: %s", card.Answer)
	}
}

// Conflicts are resolved by importance, specificity and order
// (card-rendering spec).
func TestCascade(t *testing.T) {
	cases := []struct {
		name, css, html, want string
	}{
		{"id beats a later class", "#x { color: red } .y { color: blue }",
			`<b id="x" class="y">a</b>`, `<b id="x" class="y" style="color:red">a</b>`},
		{"class beats a later element", ".y { color: red } b { color: blue }",
			`<b class="y">a</b>`, `<b class="y" style="color:red">a</b>`},
		{"later rule of the same specificity wins", ".y { color: red } .z { color: blue }",
			`<b class="y z">a</b>`, `<b class="y z" style="color:blue">a</b>`},
		{"attribute and pseudo-class count as classes", ":first-child { color: red } .y { color: blue } [title] { color: green }",
			`<b class="y" title="t">a</b>`, `<b class="y" title="t" style="color:green">a</b>`},
		{"not counts its argument", "b:not(#q) { color: red } b.y { color: blue }",
			`<b class="y">a</b>`, `<b class="y" style="color:red">a</b>`},
		{"important beats specificity", "b { color: red !important } #x { color: blue }",
			`<b id="x">a</b>`, `<b id="x" style="color:red">a</b>`},
		{"important with a space", "b { color: red ! important } #x { color: blue }",
			`<b id="x">a</b>`, `<b id="x" style="color:red">a</b>`},
		{"specificity among important rules", "#x { color: red !important } b { color: blue !important }",
			`<b id="x">a</b>`, `<b id="x" style="color:red">a</b>`},
		{"inline style beats a class rule", ".y { color: red; font-size: 20px }",
			`<b class="y" style="color: blue">a</b>`, `<b class="y" style="font-size:20px;color:blue">a</b>`},
		{"inline style beats an id rule", "#x { color: red }",
			`<b id="x" style="color: blue">a</b>`, `<b id="x" style="color:blue">a</b>`},
		{"important rule beats inline style", ".y { color: red !important }",
			`<b class="y" style="color: blue">a</b>`, `<b class="y" style="color:red">a</b>`},
		{"important inline style beats an important rule", ".y { color: red !important }",
			`<b class="y" style="color: blue !important">a</b>`, `<b class="y" style="color:blue">a</b>`},
		{"inline style without matching rules is kept as written", ".z { color: red }",
			`<b style="color: blue !important">a</b>`, `<b style="color: blue !important">a</b>`},
	}
	for _, c := range cases {
		if got := inlineIn(c.css, c.html); got != c.want {
			t.Errorf("%s:\n got %s\nwant %s", c.name, got, c.want)
		}
	}
}

func TestCardRootCascade(t *testing.T) {
	css := "* { margin: 0 } body { padding: 1px } .card { padding: 2px; color: #123 } div { margin: 5px }"
	c := renderQ(css, 1, desktop)
	if !strings.Contains(c.Question, `<div class="card card1 linux" style="margin:0;padding:2px">`) || c.Color != "#123" {
		t.Errorf("got %q, color %q", c.Question, c.Color)
	}
}

// BenchmarkInline2000Rules styles a card side of 500 elements with a
// stylesheet of 2,000 rules; it should stay well under 10 ms.
func BenchmarkInline2000Rules(b *testing.B) {
	var css strings.Builder
	for i := 0; i < 2000; i++ {
		switch i % 8 {
		case 0:
			fmt.Fprintf(&css, ".c%d { color: red }\n", i)
		case 1:
			fmt.Fprintf(&css, ".c%d b, .c%d > i { font-weight: bold }\n", i, i)
		case 2:
			fmt.Fprintf(&css, "div.c%d span.c%d { font-size: %dpx }\n", i, i-1, i%40)
		case 3:
			fmt.Fprintf(&css, "#id%d + div { display: none }\n", i)
		case 4:
			fmt.Fprintf(&css, ".card .c%d li:nth-child(2n+1) { color: blue !important }\n", i)
		case 5:
			fmt.Fprintf(&css, "[data-k=\"%d\"] ~ p:not(.c%d) { margin: 0 }\n", i, i)
		case 6:
			fmt.Fprintf(&css, ".nightMode .c%d:first-child { color: white }\n", i)
		case 7:
			fmt.Fprintf(&css, "div div div span.c%d { text-decoration: underline }\n", i)
		}
	}
	var html strings.Builder
	n := 0
	for n < 500 {
		fmt.Fprintf(&html, `<div class="c%d c%d" id="id%d"><div><div>`, n%2000, (n*7)%2000, n)
		fmt.Fprintf(&html, `<span class="c%d">x</span><b>y</b><p data-k="%d">z</p>`, (n*13)%2000, n)
		fmt.Fprintf(&html, `<ul class="c%d"><li>1</li><li>2</li><li>3</li></ul></div></div></div>`, (n*3)%2000)
		n += 10
	}
	src := html.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sheet := parseCSS(css.String(), desktop, []string{"card", "linux"})
		sheet.inline(src)
	}
}

// Background images (card-rendering spec).
func TestBackgroundImages(t *testing.T) {
	c := renderQ(`.card { background-image: url("_paper.png") }`, 1, desktop)
	if !strings.Contains(c.Question, `<div class="card card1 linux" style="background-image:url(&quot;_paper.png&quot;)">`) {
		t.Errorf("card background: %s", c.Question)
	}
	cases := []struct{ css, want string }{
		{`b { background-image: url(_dot.png) }`, `<b style="background-image:url(_dot.png)">`},
		{`b { background-image: url('_a.png'), url("_b.png") }`, `<b style="background-image:url('_a.png'),url(&quot;_b.png&quot;)">`},
		{`b { background-image: url(https://example.org/x.png) }`, `<b>`},
		{`b { background-image: url(_a.png), url(//example.org/x.png) }`, `<b>`},
		{`b { background-image: url(data:image/png;base64,AAAA) }`, `<b>`},
		{`b { background: url(_dot.png) }`, `<b>`},
		{`b { list-style-image: url(_dot.png) }`, `<b>`},
	}
	for _, c := range cases {
		if q := renderQ(c.css, 1, desktop).Question; !strings.Contains(q, c.want+"x</b>") {
			t.Errorf("%s:\n got %s\nwant %s", c.css, q, c.want)
		}
	}
}

func TestCustomProperties(t *testing.T) {
	render := func(css string) string {
		return RenderCard(Template{QFmt: `<i id="x" style="--Own: 1px">a</i>`, AFmt: ""}, css, Note{}, 0, desktop).Question
	}
	cases := []struct{ name, css, want, not string }{
		{"inlined with its case", `#x { --Shape-Color: #ff8e8e; color: red }`, `style="--Shape-Color:#ff8e8e;color:red;--Own:1px"`, ""},
		{"important beats the style attribute", `#x { --Own: 2px !important }`, `style="--Own:2px"`, "1px"},
		{"url() is dropped", `#x { --bg: url(http://example.com/a.png); color: red }`, `style="color:red;--Own:1px"`, "example.com"},
		{"cascade by specificity", `i#x { --c: blue } i { --c: red }`, `--c:blue`, "red"},
	}
	for _, c := range cases {
		q := render(c.css)
		if !strings.Contains(q, c.want) || (c.not != "" && strings.Contains(q, c.not)) {
			t.Errorf("%s:\n got %q\nwant it to contain %q", c.name, q, c.want)
		}
	}
}
