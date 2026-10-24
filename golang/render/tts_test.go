package render

import (
	"reflect"
	"strings"
	"testing"
)

func TestTTSFieldSpoken(t *testing.T) {
	// Spec: Field spoken.
	tmpl := Template{QFmt: "{{Front}}{{tts es_ES:Front}}", AFmt: "{{FrontSide}}<hr id=answer>{{Back}}"}
	card := RenderCard(tmpl, "", Note{Fields: fields("Front", "perro", "Back", "dog")}, 0, Env{})
	if !strings.Contains(card.Question, `perro<cf-tts idx="0"></cf-tts>`) {
		t.Errorf("question %q", card.Question)
	}
	want := []Utterance{{Text: "perro", Lang: "es_ES", Speed: 1}}
	if !reflect.DeepEqual(card.QuestionTTS, want) {
		t.Errorf("question utterances %+v, want %+v", card.QuestionTTS, want)
	}
	// {{FrontSide}} does not repeat the question's utterances.
	if strings.Contains(card.Answer, "cf-tts") || len(card.AnswerTTS) != 0 {
		t.Errorf("answer %q utterances %+v", card.Answer, card.AnswerTTS)
	}
}

func TestTTSClozeOnly(t *testing.T) {
	// Spec: Cloze only.
	tmpl := Template{QFmt: "{{cloze:Text}}", AFmt: "{{cloze:Text}}{{tts en_US:cloze-only:Text}}"}
	card := RenderCard(tmpl, "", Note{Fields: fields("Text", "{{c1::Paris}} is in {{c2::France}}")}, 1, Env{})
	want := []Utterance{{Text: "Paris", Lang: "en_US", Speed: 1}}
	if !reflect.DeepEqual(card.AnswerTTS, want) {
		t.Errorf("answer utterances %+v, want %+v", card.AnswerTTS, want)
	}
}

func TestTTSCloze(t *testing.T) {
	tmpl := Template{QFmt: "{{cloze:Text}}{{tts en_US:cloze:Text}}", AFmt: "{{cloze:Text}}{{tts en_US:cloze:Text}}"}
	n := Note{Fields: fields("Text", "{{c1::Paris::city}} is in <b>{{c2::France}}</b>")}
	card := RenderCard(tmpl, "", n, 1, Env{})
	if len(card.QuestionTTS) != 1 || card.QuestionTTS[0].Text != "blank is in France" {
		t.Errorf("question utterances %+v", card.QuestionTTS)
	}
	if len(card.AnswerTTS) != 1 || card.AnswerTTS[0].Text != "Paris is in France" {
		t.Errorf("answer utterances %+v", card.AnswerTTS)
	}
	// The visible cloze is unchanged.
	if !strings.Contains(card.Question, "[city]") {
		t.Errorf("question %q", card.Question)
	}
}

func TestTTSOptions(t *testing.T) {
	cases := []struct {
		name, tmpl string
		want       Utterance
	}{
		{"voices and speed", "{{tts en_US voices=Apple_Samantha,Microsoft_Zira speed=1.4:Front}}",
			Utterance{Text: "hi there", Lang: "en_US", Voices: []string{"Apple_Samantha", "Microsoft_Zira"}, Speed: 1.4}},
		{"bad speed is normal", "{{tts en_US speed=fast:Front}}", Utterance{Text: "hi there", Lang: "en_US", Speed: 1}},
		{"html stripped, entities decoded", "{{tts en_US:Rich}}", Utterance{Text: "a < b & c", Lang: "en_US", Speed: 1}},
		{"sounds left out", "{{tts en_US:Sound}}", Utterance{Text: "word", Lang: "en_US", Speed: 1}},
		{"template block", "[anki:tts lang=ja_JP voices=Kyoko speed=0.8]猫[/anki:tts]",
			Utterance{Text: "猫", Lang: "ja_JP", Voices: []string{"Kyoko"}, Speed: 0.8}},
		{"fields inside a template block", "[anki:tts lang=de_DE]{{Front}}[/anki:tts]", Utterance{Text: "hi there", Lang: "de_DE", Speed: 1}},
	}
	n := Note{Fields: fields("Front", "hi <i>there</i>", "Rich", "<div>a &lt; b</div>&amp;&nbsp;c", "Sound", "word [sound:w.mp3]")}
	for _, c := range cases {
		card := RenderCard(Template{QFmt: c.tmpl}, "", n, 0, Env{})
		if len(card.QuestionTTS) != 1 || !reflect.DeepEqual(card.QuestionTTS[0], c.want) {
			t.Errorf("%s: utterances %+v, want %+v", c.name, card.QuestionTTS, c.want)
		}
		if !strings.Contains(card.Question, `<cf-tts idx="0"></cf-tts>`) {
			t.Errorf("%s: question %q", c.name, card.Question)
		}
	}
}

func TestTTSBlockInField(t *testing.T) {
	tmpl := Template{QFmt: "{{Front}}", AFmt: "{{FrontSide}}<hr id=answer>{{Back}}"}
	n := Note{Fields: fields(
		"Front", "<b>Hund</b> [anki:tts lang=de_DE]<b>der</b> Hund[/anki:tts]",
		"Back", "[sound:dog.mp3] dog [anki:tts lang=en_US]dog[/anki:tts]",
	)}
	card := RenderCard(tmpl, "", n, 0, Env{})
	if !strings.Contains(card.Question, `<b>Hund</b> <cf-tts idx="0"></cf-tts>`) {
		t.Errorf("question %q", card.Question)
	}
	if want := []Utterance{{Text: "der Hund", Lang: "de_DE", Speed: 1}}; !reflect.DeepEqual(card.QuestionTTS, want) {
		t.Errorf("question utterances %+v", card.QuestionTTS)
	}
	// Sounds and utterances are numbered separately, in order of appearance.
	if !strings.Contains(card.Answer, `<cf-sound idx="0"></cf-sound> dog <cf-tts idx="0"></cf-tts>`) {
		t.Errorf("answer %q", card.Answer)
	}
	if want := []Utterance{{Text: "dog", Lang: "en_US", Speed: 1}}; !reflect.DeepEqual(card.AnswerTTS, want) {
		t.Errorf("answer utterances %+v", card.AnswerTTS)
	}
}

func TestTTSSanitize(t *testing.T) {
	cases := []struct {
		name, in, want string
		texts          []string
	}{
		{"several in order", "[anki:tts lang=a]1[/anki:tts]x[anki:tts lang=b]2[/anki:tts]",
			`<cf-tts idx="0"></cf-tts>x<cf-tts idx="1"></cf-tts>`, []string{"1", "2"}},
		{"inside a script, dropped", "<script>[anki:tts lang=a]1[/anki:tts]</script>[anki:tts lang=b]2[/anki:tts]",
			`<cf-tts idx="0"></cf-tts>`, []string{"2"}},
		{"inside an attribute, dropped", `<img alt="[anki:tts lang=a]1[/anki:tts]" src="x.png">`,
			`<img alt="" src="x.png">`, nil},
		{"empty text, no button", "[anki:tts lang=a] <br> [/anki:tts]ok", "ok", nil},
		{"unclosed block is text", "[anki:tts lang=a]hi", "[anki:tts lang=a]hi", nil},
	}
	for _, c := range cases {
		got, _, _, tts := sanitize(c.in)
		if got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
		var texts []string
		for _, u := range tts {
			texts = append(texts, u.Text)
		}
		if !reflect.DeepEqual(texts, c.texts) {
			t.Errorf("%s: texts %q, want %q", c.name, texts, c.texts)
		}
	}
	if got := SanitizeHTML("a[anki:tts lang=en]b[/anki:tts]c"); got != "ac" {
		t.Errorf("SanitizeHTML %q", got)
	}
}

func TestTTSVoices(t *testing.T) {
	card := RenderCard(Template{QFmt: "{{Front}}{{tts-voices:}}"}, "", Note{Fields: fields("Front", "x")}, 0, Env{})
	if !strings.Contains(card.Question, "x<cf-tts-voices></cf-tts-voices>") {
		t.Errorf("question %q", card.Question)
	}
	if p := CheckTemplate("{{Front}}{{tts-voices:}}", "{{ tts-voices: }}", []string{"Front"}, false); len(p) != 0 {
		t.Errorf("problems %+v", p)
	}
}
