package render

import "testing"

func TestTranslateChem(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"water", `\ce{H2O}`, `{\mathrm{H}_{2}\mathrm{O}}`},
		{"sulfate ion", `\ce{SO4^2-}`, `{\mathrm{SO}_{4}^{2-}}`},
		{"sulfate ion with braces", `\ce{SO4^{2-}}`, `{\mathrm{SO}_{4}^{2-}}`},
		{"charge without caret", `\ce{Ca2+ + 2Cl-}`, `{\mathrm{Ca}^{2+} + 2\mathrm{Cl}^{-}}`},
		{"reaction (spec)", `\ce{2H2 + O2 -> 2H2O}`,
			`{2\mathrm{H}_{2} + \mathrm{O}_{2} \rightarrow 2\mathrm{H}_{2}\mathrm{O}}`},
		{"reaction with states", `\ce{2H2(g) + O2(g) -> 2H2O(l)}`,
			`{2\mathrm{H}_{2}(\mathrm{g}) + \mathrm{O}_{2}(\mathrm{g}) \rightarrow 2\mathrm{H}_{2}\mathrm{O}(\mathrm{l})}`},
		{"state as its own token", `\ce{NaCl (aq)}`, `{\mathrm{NaCl} (\mathrm{aq})}`},
		{"equilibrium arrow with text", `\ce{N2 + 3H2 <=>[Fe][450 C] 2NH3}`,
			`{\mathrm{N}_{2} + 3\mathrm{H}_{2} \xrightleftharpoons[450\ \mathrm{C}]{\mathrm{Fe}} 2\mathrm{NH}_{3}}`},
		{"arrow without spaces", `\ce{A->B}`, `{\mathrm{A} \rightarrow \mathrm{B}}`},
		{"other arrows", `\ce{A <- B <-> C}`, `{\mathrm{A} \leftarrow \mathrm{B} \leftrightarrow \mathrm{C}}`},
		{"brackets and parentheses", `\ce{[Cu(NH3)4]^2+}`, `{\mathrm{[Cu(NH}_{3}\mathrm{)}_{4}\mathrm{]}^{2+}}`},
		{"hydrate", `\ce{CuSO4*5H2O}`, `{\mathrm{CuSO}_{4}\cdot 5\mathrm{H}_{2}\mathrm{O}}`},
		{"isotope", `\ce{^{227}_{90}Th}`, `{{}^{227}_{90}\mathrm{Th}}`},
		{"electron", `\ce{Fe^3+ + e- -> Fe^2+}`, `{\mathrm{Fe}^{3+} + e^{-} \rightarrow \mathrm{Fe}^{2+}}`},
		{"gas and precipitate", `\ce{CO2 ^ + BaSO4 v}`, `{\mathrm{CO}_{2} \uparrow + \mathrm{BaSO}_{4} \downarrow}`},
		{"unknown text", `\ce{hello_x}`, `{\text{hello\_x}}`},
		{"inside other TeX", `x = \ce{H2O} + \cent`, `x = {\mathrm{H}_{2}\mathrm{O}} + \cent`},
		{"units", `\pu{123 kJ/mol}`, `{123\,\mathrm{kJ/mol}}`},
		{"unclosed", `\ce{H2O`, `\ce{H2O`},
	}
	for _, c := range cases {
		if got := translateChem(c.in); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}

func TestChemInCard(t *testing.T) {
	got := wrapMath(`\(\ce{2H2 + O2 -> 2H2O}\)`)
	want := `<cf-math display="inline">{2\mathrm{H}_{2} + \mathrm{O}_{2} \rightarrow 2\mathrm{H}_{2}\mathrm{O}}</cf-math>`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
