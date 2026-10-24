package render

import "testing"

func TestPlainText(t *testing.T) {
	media := []MediaRef{{File: "hola.mp3", Kind: "audio"}, {File: "clip.mp4", Kind: "video"}}
	tts := []Utterance{{Text: "buenos días", Lang: "es_ES"}}
	for _, c := range []struct{ in, want string }{
		{`<div class="card">hola</div>`, "hola"},
		{`<b>hello</b> &amp; <i>bye</i>&nbsp;now`, "hello & bye now"},
		{`line 1<br>line 2<div>line 3</div><ul><li>a</li><li>b</li></ul>`, "line 1\nline 2\nline 3\na\nb"},
		{`see <img src="map.png"> here`, "see [image: map.png] here"},
		{`<cf-sound idx="0"></cf-sound><cf-video idx="1"></cf-video><cf-sound idx="7"></cf-sound>`, "[sound: hola.mp3] [video: clip.mp4]"},
		{`<cf-tts idx="0"></cf-tts>`, "buenos días"},
		{`<style>.x{color:red}</style>a<script>var b = "<b>";</script>c`, "ac"},
		{`Paris is the capital of <span class="cloze">[...]</span>`, "Paris is the capital of [...]"},
	} {
		if got := PlainText(c.in, media, tts); got != c.want {
			t.Errorf("PlainText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
