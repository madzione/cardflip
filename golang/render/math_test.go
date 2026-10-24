package render

import (
	"strings"
	"testing"
)

func TestWrapMath(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"inline formula (spec)", `Area is \(\pi r^2\)`, `Area is <cf-math display="inline">\pi r^2</cf-math>`},
		{"display formula (spec)", `\[\frac{a}{b}\]`, `<cf-math display="block">\frac{a}{b}</cf-math>`},
		{"several formulas", `\(a\) and \[b\] and \(c\)`,
			`<cf-math display="inline">a</cf-math> and <cf-math display="block">b</cf-math> and <cf-math display="inline">c</cf-math>`},
		{"line break inside", `\(a +<br>b\)`, `<cf-math display="inline">a + b</cf-math>`},
		{"entities inside", `\(a&nbsp;&lt;&nbsp;b\)`, `<cf-math display="inline">a &lt; b</cf-math>`},
		{"formatting tags inside", `\(<b>x</b>^2\)`, `<cf-math display="inline">x^2</cf-math>`},
		{"removed tags follow the formula", `<b>\(a</b> b\)`, `<b><cf-math display="inline">a b</cf-math></b>`},
		{"spans across elements", `<div>\(a</div><div>b\)</div>`, `<div><cf-math display="inline">ab</cf-math></div><div></div>`},
		{"unbalanced opener", `costs \( 5`, `costs \( 5`},
		{"unbalanced closer", `a \) b`, `a \) b`},
		{"mismatched pair", `\(a\]`, `\(a\]`},
		{"delimiters in attributes are not math", `<span title="\(x\)">t</span>`, `<span title="\(x\)">t</span>`},
		{"script is not crossed", `\(a<script>b\)</script>`, `\(a<script>b\)</script>`},
		{"no math", `plain <b>text</b>`, `plain <b>text</b>`},
	}
	for _, c := range cases {
		if got := wrapMath(c.in); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}

func TestMathSurvivesSanitize(t *testing.T) {
	got, _, _, _ := sanitize(wrapMath(`x \(a < b\) y`))
	if want := `x <cf-math display="inline">a &lt; b</cf-math> y`; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestClozeInMath(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		ord      int
		question string
		answer   string
	}{
		{
			name: "cloze in a formula (spec)", text: `\(E = {{c1::mc^2}}\)`, ord: 1,
			question: `<cf-math display="inline">E = \text{[...]}</cf-math>`,
			answer:   `<cf-math display="inline">E = \cfcloze{mc^2}</cf-math>`,
		},
		{
			name: "hint in a formula", text: `\(E = {{c1::mc^2::energy}}\)`, ord: 1,
			question: `<cf-math display="inline">E = \text{[energy]}</cf-math>`,
			answer:   `<cf-math display="inline">E = \cfcloze{mc^2}</cf-math>`,
		},
		{
			name: "inactive deletion in a formula", text: `\({{c1::a}} + {{c2::b}}\)`, ord: 2,
			question: `<cf-math display="inline">a + \text{[...]}</cf-math>`,
			answer:   `<cf-math display="inline">a + \cfcloze{b}</cf-math>`,
		},
		{
			name: "formula inside a deletion", text: `{{c1::\(x^2\)}} grows`, ord: 1,
			question: `<span class="cloze">[...]</span> grows`,
			answer:   `<span class="cloze"><cf-math display="inline">x^2</cf-math></span> grows`,
		},
		{
			name: "deletion outside a formula is unchanged", text: `{{c1::Einstein}}: \(E = mc^2\)`, ord: 1,
			question: `<span class="cloze">[...]</span>: <cf-math display="inline">E = mc^2</cf-math>`,
			answer:   `<span class="cloze">Einstein</span>: <cf-math display="inline">E = mc^2</cf-math>`,
		},
	}
	for _, c := range cases {
		n := Note{Fields: map[string]string{"Text": c.text}}
		tmpl := Template{QFmt: "{{cloze:Text}}", AFmt: "{{cloze:Text}}"}
		if got := wrapMath(Question(tmpl, n, c.ord)); got != c.question {
			t.Errorf("%s question:\n got %q\nwant %q", c.name, got, c.question)
		}
		if got := wrapMath(Answer(tmpl, n, c.ord)); got != c.answer {
			t.Errorf("%s answer:\n got %q\nwant %q", c.name, got, c.answer)
		}
	}
}

func TestClozeInMathColour(t *testing.T) {
	n := Note{Fields: map[string]string{"Text": `\(E = {{c1::mc^2}}\)`}}
	tmpl := Template{QFmt: "{{cloze:Text}}", AFmt: "{{cloze:Text}}"}
	// Without a CSS colour the app picks the highlight.
	if card := RenderCard(tmpl, "", n, 1, Env{}); !strings.Contains(card.Answer, `\cfcloze{mc^2}`) {
		t.Errorf("default: %q", card.Answer)
	}
	card := RenderCard(tmpl, ".cloze { font-weight: bold; color: blue; }", n, 1, Env{})
	if !strings.Contains(card.Answer, `\color{blue}{mc^2}`) {
		t.Errorf("css colour: %q", card.Answer)
	}
	// A colour TeX does not understand is left to the app.
	card = RenderCard(tmpl, ".cloze { color: rgb(0, 0, 255); }", n, 1, Env{})
	if !strings.Contains(card.Answer, `\cfcloze{mc^2}`) {
		t.Errorf("rgb colour: %q", card.Answer)
	}
}
