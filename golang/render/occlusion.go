package render

import (
	"encoding/json"
	"html"
	"regexp"
	"strconv"
	"strings"
)

// Image occlusion notes, as Anki makes them, keep their shapes as cloze
// deletions in one field, coordinates relative to the image (0 to 1):
//
//	{{c1::image-occlusion:rect:left=.1:top=.2:width=.3:height=.1:oi=1}}
//	{{c2::image-occlusion:ellipse:left=.5:top=.5:rx=.1:ry=.05:angle=10:oi=1}}
//	{{c3::image-occlusion:polygon:points=.1,.6 .3,.6 .2,.8:oi=1}}
//	{{c0::image-occlusion:text:left=.6:top=.1:text=Label:scale=1:fs=.05}}
//
// Shapes sharing a number are asked together; text (c0) is a label shown on
// every card. oi=1 is "Hide all, guess one": the shapes not being asked are
// masked too. A colon inside a value is escaped as "\:". Anki draws the
// shapes with a script in its template; the cloze filter of such a field
// instead emits a <cf-occlusion> element the app draws natively, and
// placeOcclusion removes the template's markup for the script.

// OcclusionShape is one shape of an image occlusion note. Which values are
// set depends on Kind: rect has Left, Top, Width and Height; ellipse Left
// and Top (of its bounding box), RX and RY; polygon Points; text Left, Top,
// Text, Scale and FontSize (relative to the image height). Angle rotates
// a shape in degrees around its Left and Top. Ord is the cloze number, 0 for
// a label without a card.
type OcclusionShape struct {
	Kind            string       `json:"kind"`
	Ord             int          `json:"ord"`
	Left            float64      `json:"left,omitempty"`
	Top             float64      `json:"top,omitempty"`
	Width           float64      `json:"width,omitempty"`
	Height          float64      `json:"height,omitempty"`
	RX              float64      `json:"rx,omitempty"`
	RY              float64      `json:"ry,omitempty"`
	Angle           float64      `json:"angle,omitempty"`
	Points          [][2]float64 `json:"points,omitempty"`
	Text            string       `json:"text,omitempty"`
	Scale           float64      `json:"scale,omitempty"`
	FontSize        float64      `json:"fontSize,omitempty"`
	OccludeInactive bool         `json:"oi,omitempty"`
}

var occlusionDeletion = regexp.MustCompile(`(?s)\{\{c(\d+)::image-occlusion:(.*?)\}\}`)

// hasOcclusions reports whether field text holds image occlusion shapes.
func hasOcclusions(text string) bool {
	return strings.Contains(text, "image-occlusion:") && occlusionDeletion.MatchString(text)
}

// parseOcclusions returns the shapes of a field in their order. Unknown
// shape kinds and properties are ignored, and so are shapes without a
// usable size or position.
func parseOcclusions(text string) []OcclusionShape {
	var out []OcclusionShape
	for _, m := range occlusionDeletion.FindAllStringSubmatch(text, -1) {
		ord, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		parts := splitUnescaped(m[2])
		s := OcclusionShape{Kind: strings.TrimSpace(parts[0]), Ord: ord}
		for _, p := range parts[1:] {
			key, val, ok := strings.Cut(p, "=")
			if !ok {
				continue
			}
			s.set(strings.TrimSpace(key), val)
		}
		if s.valid() {
			out = append(out, s)
		}
	}
	return out
}

// splitUnescaped splits s at the colons not escaped as "\:", and unescapes
// those.
func splitUnescaped(s string) []string {
	var parts []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '\\' && i+1 < len(s) && s[i+1] == ':':
			cur.WriteByte(':')
			i++
		case s[i] == ':':
			parts = append(parts, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(s[i])
		}
	}
	return append(parts, cur.String())
}

func (s *OcclusionShape) set(key, val string) {
	num := func() float64 {
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err != nil {
			return 0
		}
		return f
	}
	switch key {
	case "left":
		s.Left = num()
	case "top":
		s.Top = num()
	case "width":
		s.Width = num()
	case "height":
		s.Height = num()
	case "rx":
		s.RX = num()
	case "ry":
		s.RY = num()
	case "angle":
		s.Angle = num()
	case "points":
		s.Points = parsePoints(val)
	case "text":
		s.Text = html.UnescapeString(strings.ReplaceAll(val, "&nbsp;", " "))
	case "scale":
		s.Scale = num()
	case "fs":
		s.FontSize = num()
	case "oi":
		s.OccludeInactive = strings.TrimSpace(val) == "1"
	}
}

// parsePoints parses a polygon's "x,y x,y …"; a malformed pair drops the
// polygon's points.
func parsePoints(val string) [][2]float64 {
	var pts [][2]float64
	for _, pair := range strings.Fields(val) {
		xs, ys, ok := strings.Cut(pair, ",")
		x, errX := strconv.ParseFloat(xs, 64)
		y, errY := strconv.ParseFloat(ys, 64)
		if !ok || errX != nil || errY != nil {
			return nil
		}
		pts = append(pts, [2]float64{x, y})
	}
	return pts
}

func (s OcclusionShape) valid() bool {
	switch s.Kind {
	case "rect":
		return s.Width > 0 && s.Height > 0
	case "ellipse":
		return s.RX > 0 && s.RY > 0
	case "polygon":
		return len(s.Points) >= 3
	case "text":
		return strings.TrimSpace(s.Text) != ""
	}
	return false
}

// occlusionElement renders an occlusion field for cloze number ord on one
// side: a <cf-occlusion> element with the shapes as JSON, which takes the
// place of the whole field. Its id is the one of the canvas in Anki's
// template, so the note type's CSS variables for the shape colours
// (--inactive-shape-color …) are inlined on it. placeOcclusion adds the
// image.
func occlusionElement(text string, ord int, question bool) string {
	shapes := parseOcclusions(text)
	if shapes == nil {
		shapes = []OcclusionShape{}
	}
	data, _ := json.Marshal(shapes)
	side := "a"
	if question {
		side = "q"
	}
	return `<cf-occlusion id="image-occlusion-canvas" shapes="` + html.EscapeString(string(data)) +
		`" active="` + strconv.Itoa(ord) + `" side="` + side + `"></cf-occlusion>`
}

var (
	occlusionTag       = regexp.MustCompile(`<cf-occlusion [^>]*></cf-occlusion>`)
	occlusionHidden    = regexp.MustCompile(`(?is)<div\s+style\s*=\s*["']\s*display\s*:\s*none\s*;?\s*["']\s*>\s*(<cf-occlusion [^>]*></cf-occlusion>)\s*</div>`)
	occlusionContainer = regexp.MustCompile(`(?is)<div\b[^>]*\bid\s*=\s*["']?image-occlusion-container\b["']?[^>]*>`)
	occlusionErr       = regexp.MustCompile(`(?is)<div\b[^>]*\bid\s*=\s*["']?err\b["']?[^>]*>\s*</div>`)
	occlusionToggle    = regexp.MustCompile(`(?is)(<div>\s*)?<button\b[^>]*\bid\s*=\s*["']?toggle\b["']?[^>]*>.*?</button>(\s*</div>)?`)
	divTag             = regexp.MustCompile(`(?i)<(/?)div\b[^>]*>`)
	imgTag             = regexp.MustCompile(`(?is)<img\b[^>]*>`)
	imgSrc             = regexp.MustCompile(`(?is)\bsrc\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
)

// placeOcclusion finishes a rendered side that shows occlusion shapes: the
// <cf-occlusion> element gets the image of Anki's image-occlusion-container
// (or, without one, the side's first image), which is removed with the
// container, and the template's markup for its script goes: the hidden
// <div> around the shapes, the error <div> and the Toggle Masks button.
// Other sides are returned unchanged.
func placeOcclusion(side string) string {
	if !strings.Contains(side, "<cf-occlusion ") {
		return side
	}
	side = occlusionHidden.ReplaceAllString(side, "$1")
	image := ""
	if loc := occlusionContainer.FindStringIndex(side); loc != nil {
		end := closingDiv(side, loc[1])
		image = firstImage(side[loc[1]:end])
		side = side[:loc[0]] + side[end:]
	} else if loc := imgTag.FindStringIndex(side); loc != nil {
		image = firstImage(side[loc[0]:loc[1]])
		side = side[:loc[0]] + side[loc[1]:]
	}
	side = occlusionErr.ReplaceAllString(side, "")
	side = occlusionToggle.ReplaceAllString(side, "")
	return occlusionTag.ReplaceAllStringFunc(side, func(tag string) string {
		return strings.Replace(tag, "<cf-occlusion ", `<cf-occlusion image="`+strings.ReplaceAll(image, `"`, "&quot;")+`" `, 1)
	})
}

// closingDiv returns the position after the </div> that closes the <div>
// whose start tag ends at from, or the end of s when it is not closed.
func closingDiv(s string, from int) int {
	depth := 1
	for _, m := range divTag.FindAllStringSubmatchIndex(s[from:], -1) {
		if m[3] > m[2] {
			depth--
		} else {
			depth++
		}
		if depth == 0 {
			return from + m[1]
		}
	}
	return len(s)
}

// firstImage returns the src of the first <img> in s, or "".
func firstImage(s string) string {
	tag := imgTag.FindString(s)
	m := imgSrc.FindStringSubmatch(tag)
	if m == nil {
		return ""
	}
	return html.UnescapeString(unquote(m[1]))
}
