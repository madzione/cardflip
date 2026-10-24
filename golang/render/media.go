package render

import (
	"html"
	"path"
	"strconv"
	"strings"

	"github.com/tdewolff/parse/v2"
	htmllex "github.com/tdewolff/parse/v2/html"
)

// MediaRef is an audio or video file a card side plays, in order of
// appearance. The side's HTML refers to it as <cf-sound idx=N> or
// <cf-video idx=N>, where N is its index in the side's list.
type MediaRef struct {
	File string `json:"file"`
	// Kind is "audio" or "video".
	Kind string `json:"kind"`
}

// Media kinds.
const (
	Audio = "audio"
	Video = "video"
)

// Elements that refer to media and become placeholders.
var mediaElements = map[string]bool{"audio": true, "video": true}

var videoExts = map[string]bool{
	".mp4": true, ".webm": true, ".m4v": true, ".mov": true, ".mkv": true, ".avi": true,
	".mpg": true, ".mpeg": true, ".ogv": true, ".3gp": true,
}

// soundKind is the kind of a [sound:] reference, by file extension.
func soundKind(name string) string {
	if videoExts[strings.ToLower(path.Ext(name))] {
		return Video
	}
	return Audio
}

// mediaName returns the media file name of a raw src value or [sound:]
// argument, or "" when it does not name a file in the collection's media.
func mediaName(raw string) string {
	name := strings.TrimSpace(html.UnescapeString(raw))
	lower := strings.ToLower(name)
	if name == "" || strings.Contains(name, "://") || strings.HasPrefix(name, "//") || strings.HasPrefix(lower, "data:") {
		return ""
	}
	return name
}

// mediaList collects a side's media while sanitizing it.
type mediaList []MediaRef

// placeholder records a file and returns the element that stands for it.
func (m *mediaList) placeholder(file, kind string) string {
	if file == "" {
		return ""
	}
	*m = append(*m, MediaRef{File: file, Kind: kind})
	tag := "cf-sound"
	if kind == Video {
		tag = "cf-video"
	}
	return "<" + tag + ` idx="` + strconv.Itoa(len(*m)-1) + `"></` + tag + ">"
}

// sounds replaces the [sound:] tags in a text with placeholders.
func (m *mediaList) sounds(text string) string {
	return soundTag.ReplaceAllStringFunc(text, func(tag string) string {
		name := mediaName(tag[len("[sound:") : len(tag)-1])
		return m.placeholder(name, soundKind(name))
	})
}

// stripSounds removes [sound:] tags, [anki:tts] blocks, <audio> and <video>
// from raw HTML and keeps everything else, so that {{FrontSide}} on the
// answer shows the question without its sounds and utterances.
func stripSounds(src string) string {
	src = ttsTag.ReplaceAllString(src, "")
	var out strings.Builder
	l := htmllex.NewLexer(parse.NewInputString(src))
	skip := ""
	depth := 0
	var tag string
	for {
		tt, data := l.Next()
		switch tt {
		case htmllex.ErrorToken:
			return out.String()
		case htmllex.StartTagToken:
			tag = strings.ToLower(string(l.Text()))
			if skip == "" && mediaElements[tag] {
				skip, depth = tag, 0
			}
			if tag == skip {
				depth++
			}
		case htmllex.StartTagVoidToken:
			if skip != "" && tag == skip {
				depth--
				if depth == 0 {
					skip = ""
				}
				continue
			}
		case htmllex.EndTagToken:
			if skip != "" && strings.ToLower(string(l.Text())) == skip {
				depth--
				if depth == 0 {
					skip = ""
				}
				continue
			}
		case htmllex.TextToken:
			if skip == "" {
				out.WriteString(soundTag.ReplaceAllString(string(data), ""))
			}
			continue
		}
		if skip == "" {
			out.Write(data)
		}
	}
}
