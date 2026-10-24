package render

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

// Utterance is text a card side speaks with the device's speech engine.
// The side's HTML refers to it as <cf-tts idx=N>, where N is its index in
// the side's list, which is in order of appearance.
type Utterance struct {
	Text string `json:"text"`
	// Lang is the language as Anki writes it, e.g. "en_US".
	Lang string `json:"lang"`
	// Voices are the names of the voices to use, the first one the device
	// has wins; empty for any voice of the language.
	Voices []string `json:"voices"`
	// Speed is the speaking rate, 1 for normal.
	Speed float64 `json:"speed"`
}

// ttsVoicesPlaceholder is {{tts-voices:}}, which the app fills with the
// device's voices.
const ttsVoicesPlaceholder = "<cf-tts-voices></cf-tts-voices>"

// ttsTag is a [anki:tts lang=… voices=… speed=…]text[/anki:tts] block, in
// a template, a field or made by the tts filter.
var ttsTag = regexp.MustCompile(`(?s)\[anki:tts([^\]]*)\](.*?)\[/anki:tts\]`)

// isTTSFilter reports whether a filter is {{tts <lang> …:Field}}.
func isTTSFilter(name string) bool {
	return name == "tts" || strings.HasPrefix(name, "tts ")
}

// ttsFilter renders {{tts en_US voices=A,B speed=1.2:Field}} as Anki does:
// a [anki:tts] block with the field's text, which the sanitizer turns into
// an utterance. opts is the filter after "tts".
func ttsFilter(opts, v string) string {
	var head strings.Builder
	for _, o := range strings.Fields(opts) {
		if !strings.Contains(o, "=") {
			o = "lang=" + o
		}
		head.WriteString(" " + o)
	}
	text := html.EscapeString(speechText(soundTag.ReplaceAllString(v, "")))
	return "[anki:tts" + head.String() + "]" + text + "[/anki:tts]"
}

// speechText is HTML as plain text to speak, its whitespace collapsed.
func speechText(s string) string {
	return strings.Join(strings.Fields(StripHTML(s)), " ")
}

// parseUtterance reads a [anki:tts] block's options and text.
func parseUtterance(opts, text string) Utterance {
	u := Utterance{Text: speechText(text), Speed: 1}
	for _, o := range strings.Fields(html.UnescapeString(opts)) {
		key, val, _ := strings.Cut(o, "=")
		switch key {
		case "lang":
			u.Lang = val
		case "voices":
			for _, name := range strings.Split(val, ",") {
				if name = strings.TrimSpace(name); name != "" {
					u.Voices = append(u.Voices, name)
				}
			}
		case "speed":
			if s, err := strconv.ParseFloat(val, 64); err == nil && s > 0 {
				u.Speed = s
			}
		}
	}
	return u
}

// clozeForSpeech renders a cloze field for the tts filter: on the question
// the active deletion is read as "blank", on the answer every deletion is
// read out.
func clozeForSpeech(text string, ord int, question bool) string {
	var sb strings.Builder
	var write func([]clozePiece)
	write = func(ps []clozePiece) {
		for _, p := range ps {
			switch {
			case p.isText():
				sb.WriteString(p.text)
			case p.active(ord) && question:
				sb.WriteString(" blank ")
			default:
				write(p.children)
			}
		}
	}
	write(parseCloze(text))
	return sb.String()
}

// Utterances stand for their blocks while the HTML is sanitized as these
// markers: a private-use character, the index, another one. Only markers
// that end up in text become placeholders, not ones inside dropped
// elements or attributes.
const (
	ttsMarkOpen  = ""
	ttsMarkClose = ""
)

var ttsMark = regexp.MustCompile(ttsMarkOpen + `(\d+)` + ttsMarkClose)

// ttsList collects a side's utterances while sanitizing it.
type ttsList struct {
	found []Utterance // every block, by marker index
	out   []Utterance // those shown, by placeholder index
}

// mark replaces the [anki:tts] blocks in src with markers.
func (l *ttsList) mark(src string) string {
	return ttsTag.ReplaceAllStringFunc(src, func(block string) string {
		m := ttsTag.FindStringSubmatch(block)
		u := parseUtterance(m[1], m[2])
		if u.Text == "" {
			return ""
		}
		l.found = append(l.found, u)
		return ttsMarkOpen + strconv.Itoa(len(l.found)-1) + ttsMarkClose
	})
}

// placeholders replaces the markers in a text with <cf-tts> placeholders.
func (l *ttsList) placeholders(text string) string {
	if !strings.Contains(text, ttsMarkOpen) {
		return text
	}
	return ttsMark.ReplaceAllStringFunc(text, func(mark string) string {
		i, _ := strconv.Atoi(mark[len(ttsMarkOpen) : len(mark)-len(ttsMarkClose)])
		l.out = append(l.out, l.found[i])
		return `<cf-tts idx="` + strconv.Itoa(len(l.out)-1) + `"></cf-tts>`
	})
}

// unmark removes markers left where no placeholder can go.
func unmark(s string) string {
	if !strings.Contains(s, ttsMarkOpen) {
		return s
	}
	return ttsMark.ReplaceAllString(s, "")
}
