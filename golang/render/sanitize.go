package render

import (
	"regexp"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/html"
)

// Elements removed together with their content: scripts, styles (collected
// for inlining), audio and video (replaced by placeholders) and embedded
// content the app cannot show.
var droppedElements = map[string]bool{
	"script": true, "style": true, "audio": true, "video": true, "iframe": true,
	"object": true, "embed": true, "noscript": true, "template": true, "frame": true,
	"frameset": true, "applet": true, "canvas": true,
}

var urlAttrs = map[string]bool{"href": true, "src": true, "action": true, "formaction": true, "background": true, "poster": true, "xlink:href": true}

var soundTag = regexp.MustCompile(`\[sound:[^\]]*\]`)

type attr struct{ key, val string }

// sanitize removes everything that must not reach the display: scripts,
// event handler attributes, javascript: URLs and embedded content. [sound:]
// tags, <audio> and <video> become <cf-sound>/<cf-video> placeholders, and
// [anki:tts] blocks <cf-tts> placeholders. It returns the clean HTML, the
// CSS of any <style> elements, which the caller inlines, and the media and
// utterances the placeholders refer to. The <cf-math> elements
// of wrapMath pass through: they hold only escaped text. So does the empty
// <cf-typeans> placeholder of {{type:…}}, which the app turns into an input,
// and the <cf-tts-voices> placeholder of {{tts-voices:}}.
func sanitize(src string) (string, string, []MediaRef, []Utterance) {
	var tts ttsList
	src = tts.mark(src)
	var out, styles strings.Builder
	var media mediaList
	l := html.NewLexer(parse.NewInputString(src))
	skip := "" // element being dropped
	depth := 0
	mediaSrc := "" // file of the <audio> or <video> being dropped
	var tag string
	var attrs []attr
	endMedia := func() {
		if mediaElements[skip] {
			out.WriteString(media.placeholder(mediaName(mediaSrc), skip))
			mediaSrc = ""
		}
	}
	for {
		tt, data := l.Next()
		switch tt {
		case html.ErrorToken:
			endMedia() // an unclosed <audio>
			return unmark(out.String()), unmark(styles.String()), media, tts.out
		case html.StartTagToken:
			tag = strings.ToLower(string(l.Text()))
			attrs = attrs[:0]
			if skip != "" && tag == skip {
				depth++
			}
		case html.AttributeToken:
			attrs = append(attrs, attr{strings.ToLower(string(l.AttrKey())), unquote(string(l.AttrVal()))})
		case html.StartTagCloseToken, html.StartTagVoidToken:
			if skip != "" {
				if mediaElements[skip] && tag == "source" && mediaSrc == "" {
					mediaSrc = attrValue(attrs, "src")
				}
				if tt == html.StartTagVoidToken && tag == skip {
					depth--
					if depth == 0 {
						endMedia()
						skip = ""
					}
				}
				continue
			}
			if droppedElements[tag] {
				if mediaElements[tag] {
					mediaSrc = attrValue(attrs, "src")
				}
				if tt == html.StartTagCloseToken {
					skip, depth = tag, 1
				} else if mediaElements[tag] {
					out.WriteString(media.placeholder(mediaName(mediaSrc), tag))
					mediaSrc = ""
				}
				continue
			}
			writeTag(&out, tag, attrs, tt == html.StartTagVoidToken)
		case html.EndTagToken:
			name := strings.ToLower(string(l.Text()))
			if skip != "" {
				if name == skip {
					depth--
					if depth == 0 {
						endMedia()
						skip = ""
					}
				}
				continue
			}
			if droppedElements[name] {
				continue
			}
			out.WriteString("</" + name + ">")
		case html.TextToken:
			if skip == "style" {
				styles.Write(data)
				styles.WriteByte('\n')
				continue
			}
			if skip != "" {
				continue
			}
			out.WriteString(tts.placeholders(media.sounds(string(data))))
		case html.CommentToken, html.DoctypeToken, html.SVGToken, html.MathToken, html.XMLToken:
			// Dropped: comments carry nothing visible; inline SVG and MathML
			// can embed scripts and are not rendered by the app.
		default:
			if skip == "" {
				out.Write(data)
			}
		}
	}
}

func attrValue(attrs []attr, key string) string {
	for _, a := range attrs {
		if a.key == key {
			return a.val
		}
	}
	return ""
}

func writeTag(out *strings.Builder, tag string, attrs []attr, void bool) {
	out.WriteString("<" + tag)
	for _, a := range attrs {
		if strings.HasPrefix(a.key, "on") || a.key == "srcdoc" || a.key == "formaction" {
			continue
		}
		if urlAttrs[a.key] && unsafeURL(a.val) {
			continue
		}
		out.WriteString(" " + a.key)
		if a.val != "" {
			out.WriteString(`="` + strings.ReplaceAll(a.val, `"`, "&quot;") + `"`)
		}
	}
	if void {
		out.WriteString(" /")
	}
	out.WriteString(">")
}

func unsafeURL(v string) bool {
	v = strings.ToLower(strings.Map(func(r rune) rune {
		if r <= ' ' {
			return -1
		}
		return r
	}, v))
	for _, scheme := range []string{"javascript:", "vbscript:", "data:text/html", "file:"} {
		if strings.HasPrefix(v, scheme) || strings.HasPrefix(v, strings.ReplaceAll(scheme, ":", "&#58;")) {
			return true
		}
	}
	return false
}

func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}

var mediaPlaceholder = regexp.MustCompile(`<cf-(?:sound|video) idx="\d+"></cf-(?:sound|video)>`)

// SanitizeHTML cleans HTML that is not a card, such as a deck description,
// for display: as for cards, without scripts, event handlers, unsafe URLs
// and embedded content, and also without styles, sounds, videos and text
// to speech.
func SanitizeHTML(src string) string {
	out, _, _, _ := sanitize(ttsTag.ReplaceAllString(soundTag.ReplaceAllString(src, ""), ""))
	return mediaPlaceholder.ReplaceAllString(out, "")
}
