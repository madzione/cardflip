package render

import (
	"regexp"
	"strconv"
	"strings"
)

// Env is where a card is shown. It sets the card's root classes and decides
// which @media rules apply.
type Env struct {
	// Night is set while the app shows the dark theme.
	Night bool `json:"night"`
	// Platform holds the platform classes separated by spaces, such as
	// "android mobile"; empty on a desktop browser.
	Platform string `json:"platform"`
	// WidthPx and HeightPx are the card's display area in logical pixels,
	// 0 while unknown.
	WidthPx  int `json:"widthPx"`
	HeightPx int `json:"heightPx"`
}

var platformClasses = map[string]bool{
	"win": true, "mac": true, "linux": true, "android": true, "iphone": true, "ipad": true, "mobile": true,
}

// rootClasses returns the classes of the card's root element: "card",
// "cardN" for card type number ord (none for 0), the night mode classes
// and the platform classes.
func (e Env) rootClasses(ord int) []string {
	classes := []string{"card"}
	if ord > 0 {
		classes = append(classes, "card"+strconv.Itoa(ord))
	}
	if e.Night {
		classes = append(classes, "nightMode", "night_mode")
	}
	for _, p := range strings.Fields(e.Platform) {
		if platformClasses[p] && !contains(classes, p) {
			classes = append(classes, p)
		}
	}
	return classes
}

// matchMedia reports whether a media query list such as
// "screen and (max-width:600px),(prefers-color-scheme:dark)" applies.
func (e Env) matchMedia(list string) bool {
	for _, q := range strings.Split(strings.ToLower(list), ",") {
		if e.matchQuery(q) {
			return true
		}
	}
	return false
}

var mediaToken = regexp.MustCompile(`\([^()]*\)|[^\s()]+`)

// matchQuery evaluates one media query: an optional "not" or "only", an
// optional media type, and features joined by "and". Anything it does not
// understand makes the query false, also after "not".
func (e Env) matchQuery(q string) bool {
	toks := mediaToken.FindAllString(q, -1)
	if len(toks) == 0 {
		return false
	}
	negate := false
	switch toks[0] {
	case "not":
		negate, toks = true, toks[1:]
	case "only":
		toks = toks[1:]
	}
	if len(toks) == 0 {
		return false
	}
	match := true
	if !strings.HasPrefix(toks[0], "(") {
		switch toks[0] {
		case "all", "screen":
		case "print", "speech":
			match = false
		default:
			return false
		}
		toks = toks[1:]
		if len(toks) > 0 {
			if toks[0] != "and" {
				return false
			}
			toks = toks[1:]
			if len(toks) == 0 {
				return false
			}
		}
	}
	for i, t := range toks {
		if i%2 == 1 {
			if t != "and" || i == len(toks)-1 {
				return false
			}
			continue
		}
		ok, known := e.matchFeature(t)
		if !known {
			return false
		}
		match = match && ok
	}
	return match != negate
}

// matchFeature evaluates a feature such as "(max-width:600px)"; known is
// false when the feature, its value or the needed size is unknown.
func (e Env) matchFeature(f string) (ok, known bool) {
	if !strings.HasPrefix(f, "(") || !strings.HasSuffix(f, ")") {
		return false, false
	}
	name, value, found := strings.Cut(strings.Join(strings.Fields(f[1:len(f)-1]), ""), ":")
	if !found {
		return false, false
	}
	switch name {
	case "prefers-color-scheme":
		switch value {
		case "dark":
			return e.Night, true
		case "light":
			return !e.Night, true
		}
	case "min-width", "max-width":
		px, ok := cssPixels(value)
		if !ok || e.WidthPx <= 0 {
			return false, false
		}
		if name == "min-width" {
			return float64(e.WidthPx) >= px, true
		}
		return float64(e.WidthPx) <= px, true
	case "orientation":
		if e.WidthPx <= 0 || e.HeightPx <= 0 {
			return false, false
		}
		switch value {
		case "portrait":
			return e.HeightPx >= e.WidthPx, true
		case "landscape":
			return e.WidthPx > e.HeightPx, true
		}
	}
	return false, false
}

var cssLength = regexp.MustCompile(`^([0-9]*\.?[0-9]+)(px|em|rem)?$`)

// cssPixels converts a length in px, em or rem (16 px) to pixels.
func cssPixels(v string) (float64, bool) {
	m := cssLength.FindStringSubmatch(v)
	if m == nil {
		return 0, false
	}
	n, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	switch m[2] {
	case "px":
		return n, true
	case "em", "rem":
		return n * 16, true
	}
	return n, n == 0
}
