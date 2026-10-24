package render

import (
	"regexp"
	"strings"
)

// furiganaReading matches a reading as Anki writes it: base[reading], where
// the base runs back to the previous space or tag end. One space before the
// base belongs to the match, so it disappears from the output.
var furiganaReading = regexp.MustCompile(` ?([^ >]+?)\[(.+?)\]`)

// replaceReadings replaces every base[reading] in s with fn(base, reading).
// [sound:…] tags are left alone. Like Anki, &nbsp; counts as a space.
func replaceReadings(s string, fn func(base, reading string) string) string {
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return furiganaReading.ReplaceAllStringFunc(s, func(m string) string {
		sub := furiganaReading.FindStringSubmatch(m)
		if strings.HasPrefix(sub[2], "sound:") {
			return m
		}
		return fn(sub[1], sub[2])
	})
}

// furigana renders {{furigana:…}}: each base with its reading as ruby text.
func furigana(s string) string {
	return replaceReadings(s, func(base, reading string) string {
		return "<ruby>" + base + "<rt>" + reading + "</rt></ruby>"
	})
}

// kana renders {{kana:…}}: the readings in place of their bases.
func kana(s string) string {
	return replaceReadings(s, func(_, reading string) string { return reading })
}

// kanji renders {{kanji:…}}: the bases without their readings.
func kanji(s string) string {
	return replaceReadings(s, func(base, _ string) string { return base })
}
