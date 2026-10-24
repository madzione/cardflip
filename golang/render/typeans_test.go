package render

import (
	"strings"
	"testing"
)

func TestCompareTyped(t *testing.T) {
	const arrow = `<br><span id="typearrow">&darr;</span><br>`
	good := func(s string) string { return `<span class="typeGood">` + s + `</span>` }
	bad := func(s string) string { return `<span class="typeBad">` + s + `</span>` }
	missed := func(s string) string { return `<span class="typeMissed">` + s + `</span>` }
	code := func(s string) string { return `<code id="typeans">` + s + `</code>` }
	cases := []struct {
		name, expected, typed string
		nc                    bool
		want                  string
	}{
		{name: "exact match (spec)", expected: "perro", typed: "perro", want: code(good("perro"))},
		{name: "surrounding whitespace and HTML are ignored", expected: " <b>perro</b>&nbsp;", typed: " perro\n",
			want: code(good("perro"))},
		{name: "typo marks the missing character (spec)", expected: "perro", typed: "pero",
			want: code(good("pero") + arrow + good("per") + missed("r") + good("o"))},
		{name: "extra characters are wrong", expected: "gato", typed: "gatos",
			want: code(good("gato") + bad("s") + arrow + good("gato"))},
		{name: "wrong character", expected: "casa", typed: "cosa",
			want: code(good("c") + bad("o") + good("sa") + arrow + good("c") + missed("a") + good("sa"))},
		{name: "empty input shows the expected text", expected: "<i>perro</i>", typed: "  ",
			want: code("perro")},
		{name: "text is escaped", expected: "a &lt; b", typed: "a < b", want: code(good("a &lt; b"))},
		{name: "accents count without nc", expected: "über", typed: "uber",
			want: code(bad("u") + good("ber") + arrow + missed("ü") + good("ber"))},
		{name: "nc ignores accents (spec)", expected: "über", typed: "uber", nc: true, want: code(good("über"))},
		{name: "nc compares without marks", expected: "café", typed: "cafes", nc: true,
			want: code(good("cafe") + bad("s") + arrow + good("cafe"))},
		{name: "composed and decomposed input are equal", expected: "café", typed: "café",
			want: code(good("café"))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CompareTyped(tc.expected, tc.typed, tc.nc); got != tc.want {
				t.Errorf("CompareTyped(%q, %q, %v)\n got %s\nwant %s", tc.expected, tc.typed, tc.nc, got, tc.want)
			}
		})
	}
}

func TestCompareTypedCap(t *testing.T) {
	long := strings.Repeat("ab", maxTypeAnsRunes)
	got := CompareTyped(long, long+"x", false)
	want := `<code id="typeans">` + long + "x" + `<br><span id="typearrow">&darr;</span><br>` + long + `</code>`
	if got != want {
		t.Errorf("over the cap the texts are shown without marks, got %.200s…", got)
	}
	if got := CompareTyped(long, long, false); !strings.Contains(got, "typeGood") {
		t.Errorf("an exact match over the cap is still correct, got %.200s…", got)
	}
}
