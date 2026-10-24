package render

import (
	"strings"
	"testing"
)

// latexNames lists field HTML and the image names Anki 26.09.3 gives its
// LaTeX blocks (extract_latex), PNG unless svg.
var latexNames = []struct {
	field string
	svg   bool
	files []string
}{
	// The fields of the fixture packages latex.apkg and latex-legacy.apkg
	// (importer/apkg/testdata), whose media holds these names.
	{"[latex]a<br>b &amp; <b>c</b>[/latex]", false, []string{"latex-403627a63287ae1f9285454676975d2a1cc0bb84.png"}},
	{`[$]x^2[/$] and [$$]\sum_i i[/$$]`, false, []string{
		"latex-76b2878564ab19b80abb21ba964abe90fe120846.png",
		"latex-2fda732308020dc39757cf45846ca8ec1319ae47.png",
	}},
	{`[latex]\LaTeX[/latex]`, false, []string{"latex-a0188aa9bb752581e1f8018162e402bcc76b4f66.png"}},
	{"[$]E = mc^2[/$]", true, []string{"latex-3e4efedf3b1b80c4b04fd17126b5f3df859565c1.svg"}},
	{`[$$]\frac{1}{2}&nbsp;<i>x</i>[/$$]`, true, []string{"latex-63c17e1596cd6249836c819fc5364dc51e179f0a.svg"}},
	// The same source in SVG.
	{"[latex]a<br>b &amp; <b>c</b>[/latex]", true, []string{"latex-403627a63287ae1f9285454676975d2a1cc0bb84.svg"}},
	// Newlines: <br>, <br/> and <div> are, <br /> is not; a newline in the
	// field counts the same.
	{"[latex]a<br/>b[/latex]", false, []string{"latex-fcd127ffa1016069006ad91f3f361248f9bdf272.png"}},
	{"[latex]a<div>b</div>[/latex]", false, []string{"latex-fcd127ffa1016069006ad91f3f361248f9bdf272.png"}},
	{"[latex]a\nb[/latex]", false, []string{"latex-fcd127ffa1016069006ad91f3f361248f9bdf272.png"}},
	{"[latex]a<BR />b[/latex]", false, []string{"latex-da23614e02469a0d7c7bd1bdab5c9c474b1904dc.png"}},
	{"[latex]a<br>b<br />c[/latex]", false, []string{"latex-0a61d1bcbf41df80fa691814613423e9dc7394be.png"}},
	// Tags, comments and styles go; a > in an attribute ends the tag.
	{"[latex]<!-- c -->x<style>y</style>z[/latex]", false, []string{"latex-128ffb50647585c6708c29ea00975f46e0cc879b.png"}},
	{`[latex]<span style="a>b">q</span>[/latex]`, false, []string{"latex-6486f4e48077230075cf56a493d87fc46cd2b2b3.png"}},
	// Entities: decoded once; &nbsp; is a space, but a literal no-break
	// space only after decoding.
	{"[latex]x &lt; y[/latex]", false, []string{"latex-11217dd78c73a5295bd500ced4406886a40eb1bc.png"}},
	{"[latex]x < y[/latex]", false, []string{"latex-11217dd78c73a5295bd500ced4406886a40eb1bc.png"}},
	{"[latex]&amp;lt;[/latex]", false, []string{"latex-5d2c1a80d8c17e5a143aee09c29a86838f18ab02.png"}},
	{"[latex]&#x41;&#66;[/latex]", false, []string{"latex-06d945942aa26a61be18c3e22bf19bbca8dd2b5d.png"}},
	{"[latex]a&nbsp;b[/latex]", false, []string{"latex-7dbde93504122a707f849f2c12bdd9de71b41929.png"}},
	{"[latex] [/latex]", false, []string{"latex-ab90d23f7402359d51e25399fe46dac3401a3352.png"}},
	// One reference Anki does not know leaves all of them undecoded: &apos;
	// is not HTML 4, &amp needs its ;, a bare & starts no reference.
	{"[latex]&quot;&apos;&#39;[/latex]", false, []string{"latex-d670a5c75af2881293a4d846a0a2669b71a942d1.png"}},
	{"[latex]&amp[/latex]", false, []string{"latex-02a128531473fc36ecd014b8aa500b221a595e64.png"}},
	{"[latex]&AMP;[/latex]", false, []string{"latex-0c7a877a5fde6967a384f2de65e808c50a6068d9.png"}},
	{"[latex]a &foo; b[/latex]", false, []string{"latex-ca17a4d62f7617c7379f63bafbc46bfe7c0bcd1b.png"}},
	{"[latex]a & b[/latex]", false, []string{"latex-92d2cb006b06fe6cfacf5d445c4eefc95f825bf8.png"}},
	// Case-insensitive markers; the first closer ends a block, so [$] in
	// [latex] is part of its source.
	{"[LATEX]x[/LaTeX]", false, []string{"latex-11f6ad8ec52a2984abaafd7c3b516503785c2072.png"}},
	{"[latex]a [$]b[/$] c[/latex]", false, []string{"latex-04926a772507fa6cbe6039dd4b240b8317c603fc.png"}},
	{"[$]a[/$][$]b[/$]", false, []string{
		"latex-c7d457e388298246adb06c587bccd419ea67f7e8.png",
		"latex-8136a7ef6a03334a7246df9097e5bcc31ba33fd2.png",
	}},
	// No blocks: empty, or mismatched markers.
	{"[latex][/latex]", false, nil},
	{"[$$]x[/$]", false, nil},
}

func TestLatexFileNames(t *testing.T) {
	for _, tc := range latexNames {
		var got []string
		latexBlocks(tc.field, func(latex string) string {
			got = append(got, latexFile(latex, tc.svg))
			return ""
		})
		if strings.Join(got, " ") != strings.Join(tc.files, " ") {
			t.Errorf("%q (svg %v): got %q, want %q", tc.field, tc.svg, got, tc.files)
		}
	}
}

func TestLatexSource(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"a<br>b &amp; <b>c</b>", "a\nb & c"},
		{"&eacute;&lang;&rang;", "é〈〉"},
		{"&#0; &lt;", "\x00 <"},
		{"&#xD800; &lt;", "&#xD800; &lt;"},
		{"&#x110000; &lt;", "&#x110000; &lt;"},
		{"&#X41; &lt;", "&#X41; &lt;"},
		{"&#x4a;&#x4A;&lt;", "JJ<"},
		{"&#+65;&lt;", "&#+65;&lt;"},
		{"&#; &lt;", "&#; &lt;"},
		{"&lt; &nbsp;", "<  "},
	} {
		if got := latexSource(tc.raw); got != tc.want {
			t.Errorf("latexSource(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

func TestLatexImageShipped(t *testing.T) {
	stored := map[string]bool{
		"latex-76b2878564ab19b80abb21ba964abe90fe120846.png": true,
		"latex-76b2878564ab19b80abb21ba964abe90fe120846.svg": true,
	}
	tmpl := Template{QFmt: "{{Front}}", AFmt: "{{FrontSide}}<hr id=answer>{{Back}}"}
	for _, svg := range []bool{false, true} {
		n := Note{Fields: map[string]string{"Front": "[$]x^2[/$]", "Back": "b"}, LatexSVG: svg,
			HasMedia: func(name string) bool { return stored[name] }}
		card := RenderCard(tmpl, "", n, 0, Env{})
		ext := map[bool]string{false: ".png", true: ".svg"}[svg]
		want := `<img class="latex" alt="$x^2$" src="latex-76b2878564ab19b80abb21ba964abe90fe120846` + ext + `">`
		for side, h := range map[string]string{"question": card.Question, "answer": card.Answer} {
			if !strings.Contains(h, want) || strings.Contains(h, "[$]") || strings.Contains(h, "<pre") {
				t.Errorf("svg %v, %s = %s, want it to contain %s", svg, side, h, want)
			}
		}
	}
}

func TestLatexImageMissing(t *testing.T) {
	tmpl := Template{QFmt: "{{Front}}", AFmt: "{{Back}}"}
	for _, has := range []func(string) bool{nil, func(string) bool { return false }} {
		n := Note{Fields: map[string]string{"Front": `[latex]\LaTeX[/latex]`, "Back": `[latex]\[ a<b \][/latex]`}, HasMedia: has}
		card := RenderCard(tmpl, "", n, 0, Env{})
		if want := `<pre class="latex">&#92;LaTeX</pre>`; !strings.Contains(card.Question, want) || strings.Contains(card.Question, "<img") {
			t.Errorf("question = %s, want %s and no image", card.Question, want)
		}
		// The source is text: no formula, and the < stays escaped.
		if want := `<pre class="latex">&#92;[ a&lt;b &#92;]</pre>`; !strings.Contains(card.Answer, want) || strings.Contains(card.Answer, "cf-math") {
			t.Errorf("answer = %s, want %s and no formula", card.Answer, want)
		}
	}
}
