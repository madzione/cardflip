package render

import (
	stdhtml "html"
	"strconv"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/html"
)

// textBlocks are the elements PlainText puts on lines of their own.
var textBlocks = map[string]bool{
	"address": true, "article": true, "aside": true, "blockquote": true, "dd": true, "details": true,
	"div": true, "dl": true, "dt": true, "figcaption": true, "figure": true, "footer": true, "h1": true,
	"h2": true, "h3": true, "h4": true, "h5": true, "h6": true, "header": true, "hr": true, "li": true,
	"main": true, "ol": true, "p": true, "pre": true, "section": true, "table": true, "tr": true, "ul": true,
	"br": true,
}

// PlainText reduces a rendered card side (of Card) to text for reading
// without a browser: block elements and <br> end lines, entities are
// decoded, images show as [image: name], and the side's sounds and videos
// (media) as [sound: name] and [video: name]. Text to speech shows its text
// (tts). Whitespace within lines is collapsed and empty lines are dropped.
func PlainText(side string, media []MediaRef, tts []Utterance) string {
	var b strings.Builder
	skip := "" // a <script> or <style> whose content is left out
	l := html.NewLexer(parse.NewInputString(side))
	var tag string
	var attrs []attr
	get := func(key string) string {
		for _, a := range attrs {
			if a.key == key {
				return a.val
			}
		}
		return ""
	}
	index := func() int {
		i, err := strconv.Atoi(get("idx"))
		if err != nil {
			return -1
		}
		return i
	}
loop:
	for {
		tt, data := l.Next()
		switch tt {
		case html.ErrorToken:
			break loop
		case html.StartTagToken:
			tag = strings.ToLower(string(l.Text()))
			attrs = nil
		case html.AttributeToken:
			attrs = append(attrs, attr{strings.ToLower(string(l.AttrKey())), unquote(string(l.AttrVal()))})
		case html.StartTagCloseToken, html.StartTagVoidToken:
			if skip != "" {
				continue
			}
			switch tag {
			case "script", "style":
				if tt == html.StartTagCloseToken {
					skip = tag
				}
			case "img":
				b.WriteString(" [image: " + get("src") + "] ")
			case "cf-sound", "cf-video":
				if i := index(); i >= 0 && i < len(media) {
					kind := "sound"
					if tag == "cf-video" {
						kind = "video"
					}
					b.WriteString(" [" + kind + ": " + media[i].File + "] ")
				}
			case "cf-tts":
				if i := index(); i >= 0 && i < len(tts) {
					b.WriteString(" " + tts[i].Text + " ")
				}
			default:
				if textBlocks[tag] {
					b.WriteByte('\n')
				}
			}
		case html.EndTagToken:
			name := strings.ToLower(string(l.Text()))
			if skip != "" {
				if name == skip {
					skip = ""
				}
				continue
			}
			if textBlocks[name] {
				b.WriteByte('\n')
			}
		case html.TextToken:
			if skip == "" {
				b.WriteString(stdhtml.UnescapeString(string(data)))
			}
		}
	}
	var lines []string
	for _, line := range strings.Split(b.String(), "\n") {
		if line = strings.Join(strings.Fields(line), " "); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
