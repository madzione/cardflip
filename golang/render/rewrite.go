package render

import "strings"

// RenameField rewrites a template so that references to field old (plain,
// filtered and sections) point to new.
func RenameField(tmpl, old, new string) string {
	return serialize(rewrite(parseTemplate(tmpl), func(name string) (string, bool) {
		if name == old {
			return new, true
		}
		return name, true
	}))
}

// RenameFields rewrites a template like RenameField for every old name in
// names at once, so that a new name may be another field's old one.
func RenameFields(tmpl string, names map[string]string) string {
	return serialize(rewrite(parseTemplate(tmpl), func(name string) (string, bool) {
		if n, ok := names[name]; ok {
			return n, true
		}
		return name, true
	}))
}

// RemoveField rewrites a template as if field name no longer existed:
// references to it are dropped, {{#name}} sections are dropped with their
// content, and the content of {{^name}} sections is kept.
func RemoveField(tmpl, name string) string {
	return serialize(rewrite(parseTemplate(tmpl), func(n string) (string, bool) {
		return n, n != name
	}))
}

// rewrite maps field names in nodes; keep=false removes the reference.
func rewrite(nodes []node, mapName func(string) (string, bool)) []node {
	var out []node
	for _, n := range nodes {
		switch n := n.(type) {
		case fieldNode:
			name, keep := mapName(n.name)
			if keep {
				out = append(out, fieldNode{name: name, filters: n.filters})
			}
		case sectionNode:
			body := rewrite(n.body, mapName)
			name, keep := mapName(n.name)
			switch {
			case keep:
				out = append(out, sectionNode{name: name, inverted: n.inverted, body: body})
			case n.inverted:
				out = append(out, body...)
			}
		default:
			out = append(out, n)
		}
	}
	return out
}

func serialize(nodes []node) string {
	var sb strings.Builder
	var walk func([]node)
	walk = func(nodes []node) {
		for _, n := range nodes {
			switch n := n.(type) {
			case textNode:
				sb.WriteString(string(n))
			case fieldNode:
				sb.WriteString("{{")
				for _, f := range n.filters {
					sb.WriteString(f + ":")
				}
				sb.WriteString(n.name + "}}")
			case sectionNode:
				mark := "#"
				if n.inverted {
					mark = "^"
				}
				sb.WriteString("{{" + mark + n.name + "}}")
				walk(n.body)
				sb.WriteString("{{/" + n.name + "}}")
			case errorNode:
				// Template mistakes are dropped when rewriting.
			}
		}
	}
	walk(nodes)
	return sb.String()
}

// FieldsUsed returns the names of the fields a template references.
func FieldsUsed(tmpl string) []string {
	seen := map[string]bool{}
	var out []string
	var walk func([]node)
	walk = func(nodes []node) {
		for _, n := range nodes {
			switch n := n.(type) {
			case fieldNode:
				if !seen[n.name] {
					seen[n.name] = true
					out = append(out, n.name)
				}
			case sectionNode:
				walk(n.body)
			}
		}
	}
	walk(parseTemplate(tmpl))
	return out
}
