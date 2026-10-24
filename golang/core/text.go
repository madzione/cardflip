package core

import (
	"html"
	"strings"

	"cardflip/golang/render"
)

// searchFields is a note's fields for search: HTML stripped, entities
// decoded, whitespace collapsed and lower-cased.
func searchFields(fields []string) []string {
	out := make([]string, len(fields))
	for i, f := range fields {
		out[i] = strings.ToLower(strings.Join(strings.Fields(render.StripHTML(f)), " "))
	}
	return out
}

func htmlUnescape(s string) string { return html.UnescapeString(s) }

func collapseSpace(s string) string { return strings.Join(strings.Fields(s), " ") }

func stripHTML(s string) string { return render.StripHTML(s) }
