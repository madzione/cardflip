// Package usermsg makes texts for people that the app can show in their
// language, while Go itself stays locale-agnostic.
//
// A text is tagged with a code and its arguments next to its English
// wording:
//
//	"\x02" + {"code":"deckNameTaken","args":["Spanish"]} + "\x03" + English + "\x04"
//
// The app replaces each tagged text with the translation of its code, made
// with the arguments, or with the English wording when it has no
// translation for the code. Text around a tagged text, such as the context a
// wrapping error adds, is shown as it is. Arguments may be tagged texts
// themselves (an error inside a warning, say); the app translates them too.
package usermsg

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// The delimiters of a tagged text. Neither the JSON part (encoding/json
// escapes control characters) nor the English part (see Text) contains
// them.
const (
	start = "\x02"
	mid   = "\x03"
	end   = "\x04"
)

type tag struct {
	Code string   `json:"code"`
	Args []string `json:"args,omitempty"`
}

// Text returns the text fmt.Sprintf makes of format and args, tagged with
// code and args.
func Text(code, format string, args ...any) string {
	return Tag(code, Args(args...), Sprintf(format, args...))
}

// Tag tags the English text msg with code and args.
func Tag(code string, args []string, msg string) string {
	b, err := json.Marshal(tag{Code: code, Args: args})
	if err != nil {
		panic(err) // strings always marshal
	}
	return start + string(b) + mid + Plain(msg) + end
}

// Args returns args as Tag takes them: each formatted with %v. Tagged texts,
// and errors whose message is one, keep their tags.
func Args(args ...any) []string {
	if len(args) == 0 {
		return nil
	}
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = fmt.Sprint(a)
	}
	return out
}

// Sprintf is fmt.Sprintf with the tagged texts among args (and in errors
// and other values that format as one) replaced by their English wording.
func Sprintf(format string, args ...any) string {
	var plain []any
	for i, a := range args {
		if s := fmt.Sprint(a); strings.Contains(s, start) {
			if plain == nil {
				plain = slices.Clone(args)
			}
			plain[i] = Plain(s)
		}
	}
	if plain == nil {
		plain = args
	}
	return fmt.Sprintf(format, plain...)
}

// Plain returns s with each tagged text replaced by its English wording.
func Plain(s string) string {
	var b strings.Builder
	for {
		i := strings.Index(s, start)
		if i < 0 {
			break
		}
		j := strings.Index(s[i:], mid)
		k := strings.Index(s[i:], end)
		if j < 0 || k < j {
			break
		}
		b.WriteString(s[:i])
		b.WriteString(s[i+j+len(mid) : i+k])
		s = s[i+k+len(end):]
	}
	if b.Len() == 0 {
		return s
	}
	b.WriteString(s)
	return b.String()
}

// Code returns the code of the first tagged text in s, or "" if s has none.
func Code(s string) string {
	_, t, ok := first(s)
	if !ok {
		return ""
	}
	return t.Code
}

// Arguments returns the arguments of the first tagged text in s.
func Arguments(s string) []string {
	_, t, _ := first(s)
	return t.Args
}

func first(s string) (english string, t tag, ok bool) {
	i := strings.Index(s, start)
	if i < 0 {
		return "", tag{}, false
	}
	j := strings.Index(s[i:], mid)
	k := strings.Index(s[i:], end)
	if j < 0 || k < j {
		return "", tag{}, false
	}
	if json.Unmarshal([]byte(s[i+len(start):i+j]), &t) != nil {
		return "", tag{}, false
	}
	return s[i+j+len(mid) : i+k], t, true
}
