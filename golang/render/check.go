package render

import (
	"cardflip/golang/internal/usermsg"
	"slices"
	"strings"
)

// Problem is a template mistake that makes a card type unsafe to save.
type Problem struct {
	// Back is set when the problem is in the answer template.
	Back bool `json:"back"`
	// Message is a sentence naming the problem and the template side.
	Message string `json:"message"`
}

// specialFields are the names templates can use besides the note's fields.
var specialFields = map[string]bool{
	"FrontSide": true, "Tags": true, "Type": true, "Deck": true,
	"Subdeck": true, "Card": true, "CardFlag": true,
}

// CheckTemplate reports what makes a card type with these question and
// answer templates unsafe to save for a note type with fields: the mistakes
// the parser marks ({{ without }}, tags without a field name, unclosed,
// unmatched or crossed sections), references to fields that do not exist, a
// front that uses no field, and, for cloze note types, a front without a
// cloze: filter.
// It returns nil when there is nothing to report.
func CheckTemplate(qfmt, afmt string, fields []string, cloze bool) []Problem {
	known := map[string]bool{}
	for _, f := range fields {
		known[f] = true
	}
	front := checkSide(parseTemplate(qfmt), false, known, cloze)
	back := checkSide(parseTemplate(afmt), true, known, cloze)
	out := front

	usesField, usesCloze := false, false
	walkNodes(parseTemplate(qfmt), func(n node) {
		if f, ok := n.(fieldNode); ok {
			usesField = usesField || known[f.name]
			usesCloze = usesCloze || slices.Contains(f.filters, "cloze")
		}
	})
	switch {
	case cloze && !usesCloze:
		name := "Text"
		if len(fields) > 0 {
			name = fields[0]
		}
		out = append(out, Problem{Message: usermsg.Text("templateNeedsCloze", "The front template must use %s or another cloze field.", "{{cloze:"+name+"}}")})
	case !usesField && len(front) == 0:
		// With a misspelt field or a broken tag, fixing that may fix this.
		out = append(out, Problem{Message: usermsg.Text("templateNoField", "The front template uses no field, so it would never produce a card.")})
	}
	return append(out, back...)
}

// checkSide reports the syntax errors and unknown fields of one side.
func checkSide(nodes []node, back bool, known map[string]bool, cloze bool) []Problem {
	side, code := "front", "templateFront"
	if back {
		side, code = "back", "templateBack"
	}
	var out []Problem
	reported := map[string]bool{}
	unknown := func(name, tag string) {
		if known[name] || specialFields[name] || reported[name] {
			return
		}
		reported[name] = true
		out = append(out, Problem{Back: back, Message: usermsg.Text(code+"UnknownField", "The "+side+" template uses %s, but there is no field %q.", tag, name)})
	}
	walkNodes(nodes, func(n node) {
		switch n := n.(type) {
		case errorNode:
			out = append(out, Problem{Back: back, Message: usermsg.Text(code+"Mistake", "In the "+side+" template, %s.", string(n))})
		case fieldNode:
			if n.name == "" {
				return // {{tts-voices:}}
			}
			unknown(n.name, "{{"+strings.Join(append(slices.Clone(n.filters), n.name), ":")+"}}")
		case sectionNode:
			if _, ok := clozeNumber(n.name); ok && cloze && !known[n.name] {
				return
			}
			mark := "#"
			if n.inverted {
				mark = "^"
			}
			unknown(n.name, "{{"+mark+n.name+"}}")
		}
	})
	return out
}

// walkNodes calls fn for every node, sections before their content.
func walkNodes(nodes []node, fn func(node)) {
	for _, n := range nodes {
		fn(n)
		if s, ok := n.(sectionNode); ok {
			walkNodes(s.body, fn)
		}
	}
}
