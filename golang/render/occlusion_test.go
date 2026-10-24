package render

import (
	"encoding/json"
	"html"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// Anki's stock Image Occlusion template and CSS (26.09).
const (
	ioFront = `{{#Header}}<div>{{Header}}</div>{{/Header}}
<div style="display: none">{{cloze:Occlusion}}</div>
<div id="err"></div>
<div id="image-occlusion-container">
    {{Image}}
    <canvas id="image-occlusion-canvas"></canvas>
</div>
<script>
try {
    anki.imageOcclusion.setup();
} catch (exc) {
    document.getElementById("err").innerHTML = ` + "`Error loading image occlusion. Is your Anki version up to date?<br><br>${exc}`" + `;
}
</script>
`
	ioBack = ioFront + `
<div><button id="toggle">Toggle Masks</button></div>
{{#Back Extra}}<div>{{Back Extra}}</div>{{/Back Extra}}
`
	ioCSS = `#image-occlusion-canvas {
    --inactive-shape-color: #ffeba2;
    --active-shape-color: #ff8e8e;
    --inactive-shape-border: 1px #212121;
    --active-shape-border: 1px #212121;
    --highlight-shape-color: #ff8e8e00;
    --highlight-shape-border: 1px #ff8e8e;
}
.nightMode #image-occlusion-canvas { --inactive-shape-color: #444444; }

.card {
    font-family: arial;
    font-size: 20px;
    text-align: center;
    color: black;
    background-color: white;
}
`
)

// ioNote is an occlusion note with three cards: a rectangle (1), an ellipse
// (2) and a group of a polygon and a rectangle (3), all "Hide all, guess
// one", and a text label.
func ioNote(occlusion string) Note {
	return Note{Fields: map[string]string{
		"Occlusion":  occlusion,
		"Image":      `<img src="heart diagram.png">`,
		"Header":     "Heart",
		"Back Extra": "Anatomy",
		"Comments":   "",
	}, NoteType: "Image Occlusion"}
}

const ioShapes = `{{c1::image-occlusion:rect:left=.1:top=.2:width=.3:height=.1:oi=1}}<br>` +
	`{{c2::image-occlusion:ellipse:left=.5:top=.5:rx=.1:ry=.05:angle=10:oi=1}}<br>` +
	`{{c3::image-occlusion:polygon:points=.1,.6 .3,.6 .2,.8:oi=1}}<br>` +
	`{{c3::image-occlusion:rect:left=.7:top=.7:width=.1:height=.1:oi=1}}<br>` +
	`{{c0::image-occlusion:text:left=.6:top=.1:text=Label\: x &amp; y:scale=1:fs=.05:oi=1}}`

var occlusionAttr = regexp.MustCompile(`<cf-occlusion ([^>]*)>`)

// occlusionOf returns the attributes of the single <cf-occlusion> of a
// rendered side and its decoded shapes.
func occlusionOf(t *testing.T, side string) (map[string]string, []OcclusionShape) {
	t.Helper()
	all := occlusionAttr.FindAllStringSubmatch(side, -1)
	if len(all) != 1 {
		t.Fatalf("%d <cf-occlusion> elements in %s", len(all), side)
	}
	attrs := map[string]string{}
	for _, m := range regexp.MustCompile(`([a-z-]+)="([^"]*)"`).FindAllStringSubmatch(all[0][1], -1) {
		attrs[m[1]] = html.UnescapeString(m[2])
	}
	var shapes []OcclusionShape
	if err := json.Unmarshal([]byte(attrs["shapes"]), &shapes); err != nil {
		t.Fatalf("shapes %q: %v", attrs["shapes"], err)
	}
	return attrs, shapes
}

func TestParseOcclusions(t *testing.T) {
	got := parseOcclusions(ioShapes)
	want := []OcclusionShape{
		{Kind: "rect", Ord: 1, Left: .1, Top: .2, Width: .3, Height: .1, OccludeInactive: true},
		{Kind: "ellipse", Ord: 2, Left: .5, Top: .5, RX: .1, RY: .05, Angle: 10, OccludeInactive: true},
		{Kind: "polygon", Ord: 3, Points: [][2]float64{{.1, .6}, {.3, .6}, {.2, .8}}, OccludeInactive: true},
		{Kind: "rect", Ord: 3, Left: .7, Top: .7, Width: .1, Height: .1, OccludeInactive: true},
		{Kind: "text", Ord: 0, Left: .6, Top: .1, Text: "Label: x & y", Scale: 1, FontSize: .05, OccludeInactive: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseOcclusions =\n%+v\nwant\n%+v", got, want)
	}
}

func TestParseOcclusionsHideOne(t *testing.T) {
	// "Hide one, guess one" has no oi property.
	got := parseOcclusions(`{{c1::image-occlusion:rect:left=.05:top=.05:width=.2:height=.2}}`)
	want := []OcclusionShape{{Kind: "rect", Ord: 1, Left: .05, Top: .05, Width: .2, Height: .2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseOcclusions = %+v, want %+v", got, want)
	}
}

func TestParseOcclusionsIgnoresUnknown(t *testing.T) {
	got := parseOcclusions(
		// Unknown properties are dropped, the shape stays.
		`{{c1::image-occlusion:rect:left=.1:top=.1:width=.2:height=.2:ordinal=1:colour=red:noequals:oi=1}}` +
			// Unknown kinds, shapes without a size and bad numbers are dropped.
			`{{c2::image-occlusion:star:left=.1:top=.1:points=5}}` +
			`{{c3::image-occlusion:rect:left=.1:top=.1:width=0:height=.2}}` +
			`{{c4::image-occlusion:polygon:points=.1,.1 .2,x .3,.3}}` +
			`{{c5::image-occlusion:text:left=.1:top=.1:text= }}` +
			// A bad number reads as 0.
			`{{c6::image-occlusion:ellipse:left=abc:top=.1:rx=.1:ry=.1}}` +
			// Plain clozes in the same field are not shapes.
			`{{c7::Paris}}`)
	want := []OcclusionShape{
		{Kind: "rect", Ord: 1, Left: .1, Top: .1, Width: .2, Height: .2, OccludeInactive: true},
		{Kind: "ellipse", Ord: 6, Top: .1, RX: .1, RY: .1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseOcclusions =\n%+v\nwant\n%+v", got, want)
	}
}

func TestOcclusionCardNumbers(t *testing.T) {
	// One card per shape number; shapes sharing a number are one card, and
	// the text label (c0) makes none.
	got := ClozeCards(ioFront, ioNote(ioShapes).Fields)
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("ClozeCards = %v, want [1 2 3]", got)
	}
}

func TestRenderOcclusion(t *testing.T) {
	n := ioNote(ioShapes)
	tmpl := Template{QFmt: ioFront, AFmt: ioBack}
	for _, ord := range []int{1, 2, 3} {
		card := RenderCard(tmpl, ioCSS, n, ord, desktop)
		for side, html := range map[string]string{"q": card.Question, "a": card.Answer} {
			attrs, shapes := occlusionOf(t, html)
			if attrs["side"] != side || attrs["active"] != string(rune('0'+ord)) {
				t.Errorf("card %d %s: side=%q active=%q", ord, side, attrs["side"], attrs["active"])
			}
			if attrs["image"] != "heart diagram.png" {
				t.Errorf("card %d %s: image = %q", ord, side, attrs["image"])
			}
			// Every shape is listed on every card, the app picks by number.
			if len(shapes) != 5 {
				t.Errorf("card %d %s: %d shapes", ord, side, len(shapes))
			}
			// The note type's colours come with it, through the canvas id.
			if !strings.Contains(attrs["style"], "--inactive-shape-color:#ffeba2") ||
				!strings.Contains(attrs["style"], "--highlight-shape-border:1px #ff8e8e") {
				t.Errorf("card %d %s: style = %q", ord, side, attrs["style"])
			}
			// The template's markup for its script is gone, the image shows
			// only through the occlusion view, the rest stays.
			for _, gone := range []string{"image-occlusion-container", "<canvas", "<script", "anki.imageOcclusion", `id="err"`,
				"Toggle Masks", "<img", "display: none", "display:none", "{{c"} {
				if strings.Contains(html, gone) {
					t.Errorf("card %d %s keeps %q: %s", ord, side, gone, html)
				}
			}
			if !strings.Contains(html, "Heart") {
				t.Errorf("card %d %s: header missing: %s", ord, side, html)
			}
			if side == "a" && !strings.Contains(html, "Anatomy") {
				t.Errorf("card %d answer: back extra missing: %s", ord, html)
			}
		}
	}
}

func TestRenderOcclusionNightColours(t *testing.T) {
	card := RenderCard(Template{QFmt: ioFront, AFmt: ioBack}, ioCSS, ioNote(ioShapes), 1, desktopNight)
	attrs, _ := occlusionOf(t, card.Question)
	if !strings.Contains(attrs["style"], "--inactive-shape-color:#444444") {
		t.Errorf("style = %q", attrs["style"])
	}
}

func TestRenderOcclusionModes(t *testing.T) {
	tmpl := Template{QFmt: ioFront, AFmt: ioBack}
	hideOne := `{{c1::image-occlusion:rect:left=.05:top=.05:width=.2:height=.2}}<br>` +
		`{{c2::image-occlusion:rect:left=.6:top=.6:width=.3:height=.25}}`
	for _, tc := range []struct {
		field string
		oi    bool
	}{{ioShapes, true}, {hideOne, false}} {
		_, shapes := occlusionOf(t, RenderCard(tmpl, ioCSS, ioNote(tc.field), 1, desktop).Question)
		for _, s := range shapes {
			if s.OccludeInactive != tc.oi {
				t.Errorf("%s: shape %+v, want oi %v", tc.field, s, tc.oi)
			}
		}
	}
}

func TestRenderOcclusionWithoutContainer(t *testing.T) {
	// A cloned note type with its own template, detected by content: the
	// first image of the side is the one occluded.
	tmpl := Template{QFmt: `<p>{{Header}}</p>{{Image}}{{cloze:Occlusion}}<p>after</p>`, AFmt: `{{cloze:Occlusion}}{{Image}}`}
	n := ioNote(ioShapes)
	card := RenderCard(tmpl, "", n, 2, desktop)
	for _, side := range []string{card.Question, card.Answer} {
		attrs, _ := occlusionOf(t, side)
		if attrs["image"] != "heart diagram.png" || strings.Contains(side, "<img") {
			t.Errorf("side = %s", side)
		}
	}
	if !strings.Contains(card.Question, "after") {
		t.Errorf("question = %s", card.Question)
	}
}

func TestPlainClozeUnaffected(t *testing.T) {
	q := RenderCard(Template{QFmt: "{{cloze:Text}}", AFmt: "{{cloze:Text}}"}, "",
		Note{Fields: map[string]string{"Text": "{{c1::Paris}} <img src=a.png>"}}, 1, desktop).Question
	if strings.Contains(q, "cf-occlusion") || !strings.Contains(q, `src="a.png"`) {
		t.Errorf("question = %s", q)
	}
}
