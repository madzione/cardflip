package render

import (
	"fmt"
	"testing"
)

var clozeTmpl = Template{QFmt: "{{cloze:Text}}", AFmt: "{{cloze:Text}}<br>{{Back Extra}}"}

// Golden cases for cloze deletions.
func TestClozeGolden(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		ord      int
		question string
		answer   string
	}{
		{
			name: "cloze question and answer (spec)", text: "{{c1::Paris}} is in {{c2::France}}", ord: 1,
			question: `<span class="cloze">[...]</span> is in France`,
			answer:   `<span class="cloze">Paris</span> is in France<br>`,
		},
		{
			name: "second card of the same note", text: "{{c1::Paris}} is in {{c2::France}}", ord: 2,
			question: `Paris is in <span class="cloze">[...]</span>`,
			answer:   `Paris is in <span class="cloze">France</span><br>`,
		},
		{
			name: "cloze hint (spec)", text: "{{c1::Paris::city}} is in France", ord: 1,
			question: `<span class="cloze">[city]</span> is in France`,
			answer:   `<span class="cloze">Paris</span> is in France<br>`,
		},
		{
			name: "hint of an inactive cloze is not shown", text: "{{c1::Paris::city}} is in {{c2::France::country}}", ord: 2,
			question: `Paris is in <span class="cloze">[country]</span>`,
			answer:   `Paris is in <span class="cloze">France</span><br>`,
		},
		{
			name: "same number twice", text: "{{c1::a}} and {{c1::b}}", ord: 1,
			question: `<span class="cloze">[...]</span> and <span class="cloze">[...]</span>`,
			answer:   `<span class="cloze">a</span> and <span class="cloze">b</span><br>`,
		},
		{
			name: "markup inside a deletion", text: "{{c1::<b>Paris</b>}}", ord: 1,
			question: `<span class="cloze">[...]</span>`,
			answer:   `<span class="cloze"><b>Paris</b></span><br>`,
		},
		{
			name: "nested deletion, outer active", text: "{{c1::The {{c2::capital}} Paris}}", ord: 1,
			question: `<span class="cloze">[...]</span>`,
			answer:   `<span class="cloze">The capital Paris</span><br>`,
		},
		{
			name: "nested deletion, inner active", text: "{{c1::The {{c2::capital}} Paris}}", ord: 2,
			question: `The <span class="cloze">[...]</span> Paris`,
			answer:   `The <span class="cloze">capital</span> Paris<br>`,
		},
		{
			name: "multi-number deletion, first listed card (spec)", text: "{{c1,2::Paris}} is in {{c3::France}}", ord: 1,
			question: `<span class="cloze">[...]</span> is in France`,
			answer:   `<span class="cloze">Paris</span> is in France<br>`,
		},
		{
			name: "multi-number deletion, second listed card (spec)", text: "{{c1,2::Paris}} is in {{c3::France}}", ord: 2,
			question: `<span class="cloze">[...]</span> is in France`,
			answer:   `<span class="cloze">Paris</span> is in France<br>`,
		},
		{
			name: "multi-number deletion, unlisted card (spec)", text: "{{c1,2::Paris}} is in {{c3::France}}", ord: 3,
			question: `Paris is in <span class="cloze">[...]</span>`,
			answer:   `Paris is in <span class="cloze">France</span><br>`,
		},
		{
			name: "hint on a multi-number deletion (spec)", text: "{{c1,2::Paris::city}}", ord: 2,
			question: `<span class="cloze">[city]</span>`,
			answer:   `<span class="cloze">Paris</span><br>`,
		},
		{
			name: "multi-number deletion nested, inner active", text: "{{c1::The {{c2, 3::capital}} Paris}}", ord: 3,
			question: `The <span class="cloze">[...]</span> Paris`,
			answer:   `The <span class="cloze">capital</span> Paris<br>`,
		},
		{
			name: "multi-number deletion nested, outer active", text: "{{c1::The {{c2, 3::capital}} Paris}}", ord: 1,
			question: `<span class="cloze">[...]</span>`,
			answer:   `<span class="cloze">The capital Paris</span><br>`,
		},
		{
			name: "malformed number list is literal (spec)", text: "{{c1,,2::Paris}}", ord: 1,
			question: `{{c1,,2::Paris}}`, answer: `{{c1,,2::Paris}}<br>`,
		},
		{
			name: "unclosed deletion stays literal", text: "{{c1::Paris is", ord: 1,
			question: `{{c1::Paris is`, answer: `{{c1::Paris is<br>`,
		},
	}
	for _, c := range cases {
		n := Note{Fields: fields("Text", c.text, "Back Extra", "")}
		if got := Question(clozeTmpl, n, c.ord); got != c.question {
			t.Errorf("%s: question\n got %q\nwant %q", c.name, got, c.question)
		}
		if got := Answer(clozeTmpl, n, c.ord); got != c.answer {
			t.Errorf("%s: answer\n got %q\nwant %q", c.name, got, c.answer)
		}
	}
}

func TestClozeNumbers(t *testing.T) {
	cases := map[string]string{
		"{{c1::Paris}} is the capital of {{c2::France}}": "[1 2]",
		"{{c3::a}} {{c1::b}} {{c3::c}}":                  "[1 3]",
		"{{c1::outer {{c4::inner}}}}":                    "[1 4]",
		"no clozes":                                      "[]",
		"{{c0::zero}} {{cx::bad}} {{c2:one colon}}":      "[]",
		"{{C2::upper}}":                                  "[2]",
		"{{c1,2::x}} {{c4::y}}":                          "[1 2 4]",
		"{{c1,,2::bad}} {{c0,1::bad}}":                   "[]",
	}
	for in, want := range cases {
		if got := fmt.Sprint(ClozeNumbers(in)); got != want {
			t.Errorf("ClozeNumbers(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestClozeCards(t *testing.T) {
	f := fields("Text", "{{c1::Paris}} is the capital of {{c2::France}}", "Back Extra", "{{c5::not used}}")
	if got := fmt.Sprint(ClozeCards("{{cloze:Text}}", f)); got != "[1 2]" {
		t.Errorf("cards = %s", got)
	}
	// Editing removes a deletion: only cloze 1 remains.
	f["Text"] = "{{c1::Paris}} is the capital of France"
	if got := fmt.Sprint(ClozeCards("{{cloze:Text}}", f)); got != "[1]" {
		t.Errorf("cards after edit = %s", got)
	}
}

func TestClozeOpen(t *testing.T) {
	cases := map[string]string{
		"{{c1::x}}":       "[1] 6",
		"{{c1,2::x}}":     "[1 2] 8",
		"{{c1, 2, 3::x}}": "[1 2 3] 12",
		"{{C12,3::x}}":    "[12 3] 9",
		"{{c1,,2::x}}":    "[] 0",
		"{{c0,1::x}}":     "[] 0",
		"{{c1,0::x}}":     "[] 0",
		"{{c1,::x}}":      "[] 0",
		"{{c,1::x}}":      "[] 0",
		"{{c1 ,2::x}}":    "[] 0",
		"{{c1,2:x}}":      "[] 0",
		"{{c1,2":          "[] 0",
	}
	for in, want := range cases {
		ords, n := clozeOpen(in)
		if got := fmt.Sprint(ords, " ", n); got != want {
			t.Errorf("clozeOpen(%q) = %s, want %s", in, got, want)
		}
	}
}
