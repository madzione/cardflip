// Package search parses browser queries and matches cards against them.
//
// Terms separated by spaces must all match (AND). A term is free
// text, matched case-insensitively within each field with HTML stripped and
// entities decoded, a "quoted phrase", or a qualifier: deck:<name> (with
// subdecks; deck:<name>::* only the subdecks; cards in filtered decks match
// their home deck too), deck:current (the deck most recently studied),
// deck:filtered (cards in filtered decks), tag:<tag> (with child tags), tag:none, note:<note
// type>, preset:<deck options preset>, card:<card type name or number>,
// is:new|due|review|learn|suspended|buried|buried-manually|buried-sibling
// (relearning cards are both learn and review), flag:0 (no flag) to flag:7,
// or nid:<IDs> and cid:<IDs>, comma-separated. A leading "-" negates a
// term.
//
// prop:<property><op><number> compares ivl (days), due (days from today),
// reps, lapses, pos (new card position), s (stability in days), d
// (difficulty, 0 to 1) or r (retrievability, 0 to 1); cards without the
// property do not match.
//
// In text, "_" matches one character and "*" any run. w:<word> matches
// whole words, nc:<text> ignores accents, sc:<text> ignores cloze markup and
// re:<expr> matches a regular expression against the fields as stored.
// <field>:<text> matches a field whose whole text is <text>, and
// <field>:re:<expr> and tag:re:<expr> match a regular expression; field
// names may contain "*".
//
// Event terms count study days back from today, which is day 1: added:N
// and edited:N match cards added, or whose note was edited, in the last N
// days; rated:N cards answered in them (at most 365), rated:N:R with
// rating R (1 to 4); introduced:N cards first answered in them; resched:N
// cards with a manual entry (a change of scheduling) in them.
// prop:rated<op>D and prop:resched<op>D compare the day of the latest
// answer or manual entry with D, where 0 is today and -1 yesterday; <op> is
// one of =, !=, <, >, <= and >=.
//
// "or" between terms matches either, and "and" may be written explicitly;
// AND binds more tightly than OR. Parentheses group terms, and "-(…)"
// negates a group. Quoted text is one term with spaces and parentheses
// taken literally; \", \(, \), \-, \:, \_, \* and \\ stand for the
// literal characters.
package search

import (
	"cardflip/golang/internal/usermsg"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"cardflip/golang/render"
	"cardflip/golang/scheduler"
)

// Card is what a term is matched against. The caller fills one value per
// card; strings are compared lower-cased.
type Card struct {
	// Fields are the note's fields with HTML stripped, entities decoded,
	// whitespace collapsed and letters lower-cased.
	Fields []string
	// Raw are the note's fields as stored, for regular expressions.
	Raw []string
	// FieldNames are the note type's field names, lower-cased, in the
	// order of Fields.
	FieldNames []string
	// Excluded marks the fields that unqualified terms (text, w: and re:)
	// skip, in the order of Fields; nil when there are none. Terms naming
	// a field still match it.
	Excluded []bool
	// Deck is the card's deck. HomeDeck is the deck a card in a filtered
	// deck returns to (Filtered is set then), else Deck.
	Deck     string
	HomeDeck string
	Filtered bool
	Tags     []string
	NoteType string
	State    State
	// Preset is the name of the deck's options preset.
	Preset string
	// Due is set when the card is due in the current study day.
	Due       bool
	Suspended bool
	// Flag is 0 for no flag, or 1 to 7.
	Flag int
	// BuriedManually is set while the user has the card buried,
	// BuriedSibling while sibling burying keeps it out of today's study
	// because another card of its note was answered today. Only queries
	// for which UsesSiblingBurial reports true look at BuriedSibling.
	BuriedManually, BuriedSibling bool
	// Interval is a Review card's interval in days, DueDay the study day a
	// card other than New is due on, relative to today (negative when
	// overdue), and Position a New card's position. Stability (in days),
	// Difficulty (0 to 1) and Retrievability (0 to 1) come from the FSRS
	// memory state, retrievability at search time. Each is NaN for a card
	// without it.
	Interval, DueDay, Position            float64
	Stability, Difficulty, Retrievability float64
	Reps, Lapses                          int
	// CardOrd is the card type's number from 1, or the cloze number for
	// cloze notes; CardName is the card type's name.
	CardOrd  int
	CardName string
	// NoteID and CardID are the note's and the card's IDs.
	NoteID, CardID int64
	// Added is when the card was created, Edited when its note last
	// changed.
	Added, Edited time.Time
	// Reviews are the card's review log entries, oldest first. Only queries
	// for which UsesReviews reports true look at them.
	Reviews []LogEntry
}

// LogEntry is one review log entry of a card.
type LogEntry struct {
	// Time is in Unix milliseconds.
	Time int64
	// Rating is 1 to 4 for an answer, or 0 for a manual entry (a change of
	// scheduling).
	Rating int
}

// State is a card's scheduling state as far as is: cares.
type State int

const (
	New State = iota
	Learning
	Review
	Relearning
)

type kind int

const (
	text kind = iota
	deck
	tag
	note
	preset
	isNew
	isDue
	isReview
	isLearn
	isSuspended
	flag
	word // w:
	field
	// regex matches the stored fields, tagRegex the tags.
	regex
	tagRegex
	// The event kinds match a time in [since, until), or outside it when
	// outside is set (see term).
	added
	edited
	rated
	introduced
	resched
	propRated
	propResched
	// propNum compares a numeric card property (prop:ivl and the like).
	propNum
	isBuried // manually or by sibling burying
	isBuriedManually
	isBuriedSibling
	cardOrd
	cardName
	tagNone
	deckCurrent
	deckFiltered
	noteID
	cardID
)

// property is a numeric card property that prop: compares.
type property int

const (
	propIvl property = iota
	propDue
	propReps
	propLapses
	propPos
	propS
	propD
	propR
)

// properties maps prop: names to properties, with an example comparison
// for error messages.
var properties = map[string]struct {
	p       property
	example string
}{
	"ivl": {propIvl, ">=30"}, "due": {propDue, "<0"}, "reps": {propReps, ">10"}, "lapses": {propLapses, ">3"},
	"pos": {propPos, "<100"}, "s": {propS, ">21"}, "d": {propD, ">0.7"}, "r": {propR, "<0.8"},
}

// of returns the card's value of the property, or NaN when it has none.
func (p property) of(c *Card) float64 {
	switch p {
	case propIvl:
		return c.Interval
	case propDue:
		return c.DueDay
	case propReps:
		return float64(c.Reps)
	case propLapses:
		return float64(c.Lapses)
	case propPos:
		return c.Position
	case propS:
		return c.Stability
	case propD:
		return c.Difficulty
	case propR:
		return c.Retrievability
	}
	return math.NaN()
}

type term struct {
	kind  kind
	value string // lower-cased
	flag  int
	// re, when set, replaces value for text, word, field and the regex
	// kinds.
	re *regexp.Regexp
	// fold ignores accents (nc: or Options.IgnoreAccents); cloze strips
	// cloze markup (sc:).
	fold, cloze bool
	// name is a field term's field name, lower-cased; nameRE is set when
	// it has a "*". raw matches the stored field (<field>:re:).
	name   string
	nameRE *regexp.Regexp
	raw    bool
	// since and until bound an event kind's time in Unix milliseconds, and
	// outside inverts the range (prop:…!=). rating, when not 0, is the
	// rating a rated term wants.
	since, until int64
	outside      bool
	rating       int
	// prop, cmp and num are a propNum term's property, comparison and
	// number; num is also a cardOrd term's number.
	prop property
	cmp  string
	num  float64
	// ids are a noteID or cardID term's IDs.
	ids map[int64]bool
}

type op int

const (
	opTerm op = iota
	opAnd
	opOr
)

// node is an expression tree: a term, or the AND or OR of its children.
type node struct {
	op       op
	negate   bool
	term     term
	children []*node
}

// Query is a parsed query. The zero Query matches every card.
type Query struct {
	root *node
}

// Error reports an invalid term, or a syntax error such as an unclosed
// parenthesis.
type Error struct {
	// Term is the invalid term as typed, or "" for a syntax error.
	Term string
	// Pos is the 1-based character position of the term or syntax error.
	Pos int
	// Msg says what is wrong, in English; Code and Args tag it so the app
	// can show it translated (see package usermsg).
	Msg  string
	Code string
	Args []string
}

// newError returns an *Error with the message format makes of args.
func newError(term string, pos int, code, format string, args ...any) *Error {
	return &Error{Term: term, Pos: pos, Msg: fmt.Sprintf(format, args...), Code: code, Args: usermsg.Args(args...)}
}

// Error returns the tagged message with the term or position.
func (e *Error) Error() string {
	msg := usermsg.Tag(e.Code, e.Args, e.Msg)
	if e.Term == "" {
		return usermsg.Text("searchErrorAt", "%s (character %d)", msg, e.Pos)
	}
	return usermsg.Text("searchErrorIn", "%s: %s", e.Term, msg)
}

// Options change how a query matches.
type Options struct {
	// IgnoreAccents makes every text and field term ignore accents and
	// other combining marks, as nc: does.
	IgnoreAccents bool
	// Now and RolloverHour place the study days that event terms count; a
	// zero Now means the current time.
	Now          time.Time
	RolloverHour int
	// CurrentDeck is the name of the deck deck:current matches, with its
	// subdecks; "" matches no card.
	CurrentDeck string
}

// maxRatedDays is the most days rated: looks back, as in Anki.
const maxRatedDays = 365

// Parse parses a query with default options.
func Parse(q string) (Query, error) { return ParseWith(q, Options{}) }

// ParseWith parses a query.
func ParseWith(q string, o Options) (Query, error) {
	p := parser{toks: tokenize(q), opts: o}
	root, err := p.or()
	if err != nil {
		return Query{}, err
	}
	if t := p.peek(); t != nil {
		// or() stops only at the end or at a ")" with no group open.
		return Query{}, newError("", t.pos, "searchParenNotOpened", "a closing parenthesis has no opening one")
	}
	return Query{root: root}, nil
}

type tokKind int

const (
	tWord tokKind = iota
	tOpen
	tClose
	tAnd
	tOr
)

type token struct {
	kind tokKind
	pos  int    // 1-based character position
	raw  string // as typed
	// text is a word with quotes removed. Escapes are kept as typed; each
	// kind of term resolves them (see unescape and wildcard).
	text string
	// colon is the byte offset in text of the first unescaped colon, or -1.
	// A quoted colon counts, so "deck:My Deck" is a deck term.
	colon int
	// negate is set by a leading "-" on a word or "(".
	negate bool
}

// escapable are the characters a backslash makes literal. A backslash
// before anything else is kept as typed, except that in text "\_" and "\*"
// are a literal "_" and "*".
const escapable = `"()-:\`

// tokenize splits a query into words, parentheses and the operators "and"
// and "or" (any case). Whitespace and unquoted, unescaped parentheses end a
// word; a "-" before a word or "(" negates it.
func tokenize(q string) []token {
	rs := []rune(q)
	isBreak := func(i int) bool {
		return i >= len(rs) || unicode.IsSpace(rs[i]) || rs[i] == '(' || rs[i] == ')'
	}
	var toks []token
	for i := 0; i < len(rs); {
		switch {
		case unicode.IsSpace(rs[i]):
			i++
			continue
		case rs[i] == '(':
			toks = append(toks, token{kind: tOpen, pos: i + 1, raw: "("})
			i++
			continue
		case rs[i] == '-' && i+1 < len(rs) && rs[i+1] == '(':
			toks = append(toks, token{kind: tOpen, pos: i + 1, raw: "-(", negate: true})
			i += 2
			continue
		case rs[i] == ')':
			toks = append(toks, token{kind: tClose, pos: i + 1, raw: ")"})
			i++
			continue
		}
		t := token{kind: tWord, pos: i + 1, colon: -1}
		start := i
		if rs[i] == '-' && !isBreak(i+1) {
			t.negate = true
			i++
		}
		var b strings.Builder
		quoted := false
		for ; i < len(rs) && (quoted || !isBreak(i)); i++ {
			switch r := rs[i]; {
			case r == '\\' && i+1 < len(rs) && strings.ContainsRune(escapable, rs[i+1]):
				b.WriteRune(r)
				i++
				b.WriteRune(rs[i])
			case r == '"':
				quoted = !quoted
			case r == ':' && t.colon < 0:
				t.colon = b.Len()
				b.WriteRune(r)
			default:
				b.WriteRune(r)
			}
		}
		t.raw = string(rs[start:i])
		t.text = b.String()
		if !norm.NFC.IsNormalString(t.text) {
			// Stored text is in NFC, so terms match in NFC. Positions stay
			// those of the query as typed. A colon composes with nothing,
			// so its offset is that of the normalised text before it.
			if t.colon >= 0 {
				t.colon = len(norm.NFC.String(t.text[:t.colon]))
			}
			t.text = norm.NFC.String(t.text)
		}
		switch {
		case strings.EqualFold(t.raw, "and"):
			t.kind = tAnd
		case strings.EqualFold(t.raw, "or"):
			t.kind = tOr
		}
		toks = append(toks, t)
	}
	return toks
}

// parser is a recursive-descent parser over tokens. AND, explicit or
// implicit, binds more tightly than OR:
//
//	or    = and { "or" and }
//	and   = unary { ["and"] unary }
//	unary = word | ["-"] "(" or ")"
type parser struct {
	toks []token
	i    int
	opts Options
}

func (p *parser) peek() *token {
	if p.i < len(p.toks) {
		return &p.toks[p.i]
	}
	return nil
}

// or parses alternatives up to the end or a ")"; it returns nil when there
// are none.
func (p *parser) or() (*node, error) {
	var alts []*node
	var lastOr *token
	for {
		a, err := p.and()
		if err != nil {
			return nil, err
		}
		if a == nil {
			if lastOr != nil {
				return nil, needsTerms(lastOr)
			}
			if t := p.peek(); t != nil && t.kind == tOr {
				return nil, needsTerms(t)
			}
			return nil, nil
		}
		alts = append(alts, a)
		t := p.peek()
		if t == nil || t.kind != tOr {
			break
		}
		lastOr = t
		p.i++
	}
	if len(alts) == 1 {
		return alts[0], nil
	}
	return &node{op: opOr, children: alts}, nil
}

// and parses terms and groups up to "or", ")" or the end; it returns nil
// when there are none.
func (p *parser) and() (*node, error) {
	var all []*node
	for {
		t := p.peek()
		if t == nil || t.kind == tOr || t.kind == tClose {
			break
		}
		if t.kind == tAnd {
			p.i++
			if n := p.peek(); len(all) == 0 || n == nil || n.kind == tAnd || n.kind == tOr || n.kind == tClose {
				return nil, needsTerms(t)
			}
			continue
		}
		n, err := p.unary()
		if err != nil {
			return nil, err
		}
		all = append(all, n)
	}
	switch len(all) {
	case 0:
		return nil, nil
	case 1:
		return all[0], nil
	}
	return &node{op: opAnd, children: all}, nil
}

// unary parses a word or a group.
func (p *parser) unary() (*node, error) {
	t := &p.toks[p.i]
	p.i++
	if t.kind == tWord {
		tm, err := parseTerm(t, p.opts)
		if err != nil {
			return nil, err
		}
		return &node{op: opTerm, term: tm, negate: t.negate}, nil
	}
	// tOpen
	if n := p.peek(); n != nil && n.kind == tClose {
		return nil, newError("", t.pos, "searchParenEmpty", "empty parentheses")
	}
	inner, err := p.or()
	if err != nil {
		return nil, err
	}
	if n := p.peek(); n == nil || inner == nil {
		return nil, newError("", t.pos, "searchParenNotClosed", "a parenthesis is not closed")
	}
	p.i++
	inner.negate = inner.negate != t.negate
	return inner, nil
}

func needsTerms(t *token) error {
	return newError("", t.pos, "searchOperatorAlone", "%q needs a search term on both sides", strings.ToLower(t.raw))
}

func parseTerm(tk *token, o Options) (term, error) {
	t := term{fold: o.IgnoreAccents}
	bad := func(code, format string, args ...any) (term, error) {
		return t, newError(tk.raw, tk.pos, code, format, args...)
	}
	value := tk.text
	if tk.colon > 0 {
		key := strings.ToLower(unescape(tk.text[:tk.colon]))
		value = tk.text[tk.colon+1:]
		v := strings.ToLower(strings.TrimSpace(unescape(value)))
		switch key {
		case "tag":
			if expr, ok := cutPrefixFold(value, "re:"); ok {
				t.kind = tagRegex
				return t, t.compileRegex(tk, expr)
			}
			fallthrough
		case "deck", "note", "preset":
			if v == "" {
				return bad("searchNeedsName", "needs a name, e.g. %s:Spanish", key)
			}
			t.value = v
			t.kind = map[string]kind{"deck": deck, "tag": tag, "note": note, "preset": preset}[key]
			switch {
			case key == "tag" && v == "none":
				t.kind = tagNone
			case key == "deck" && v == "current":
				t.kind = deckCurrent
				t.value = strings.ToLower(o.CurrentDeck)
			case key == "deck" && v == "filtered":
				t.kind = deckFiltered
			}
			return t, nil
		case "is":
			k, ok := map[string]kind{"new": isNew, "due": isDue, "review": isReview, "learn": isLearn,
				"suspended": isSuspended, "buried": isBuried, "buried-manually": isBuriedManually,
				"buried-sibling": isBuriedSibling}[v]
			if !ok {
				return bad("searchUnknownState", "unknown state; use is:new, is:due, is:review, is:learn, is:suspended, is:buried, "+
					"is:buried-manually or is:buried-sibling")
			}
			t.kind = k
			return t, nil
		case "card":
			if v == "" {
				return bad("searchNeedsCardType", "needs a card type name or number, e.g. card:Reverse or card:2")
			}
			if n, err := strconv.Atoi(v); err == nil {
				if n < 1 {
					return bad("searchCardTypeNumber", "card type numbers start at 1, e.g. card:2")
				}
				t.kind = cardOrd
				t.num = float64(n)
				return t, nil
			}
			t.kind = cardName
			re, lit, wild := wildcard(strings.ToLower(strings.TrimSpace(value)), ".", ".*")
			if wild {
				t.re = regexp.MustCompile("(?s)^" + re + "$")
			} else {
				t.value = lit
			}
			return t, nil
		case "nid", "cid":
			t.kind = map[string]kind{"nid": noteID, "cid": cardID}[key]
			t.ids = map[int64]bool{}
			for s := range strings.SplitSeq(v, ",") {
				id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
				if err != nil {
					return bad("searchNeedsIds", "needs an ID or a comma-separated list of IDs, e.g. %s:1712345678901,1712345678902", key)
				}
				t.ids[id] = true
			}
			return t, nil
		case "flag":
			if len(v) != 1 || v[0] < '0' || v[0] > '7' {
				return bad("searchUnknownFlag", "unknown flag; use flag:0 (no flag) to flag:7")
			}
			t.kind = flag
			t.flag = int(v[0] - '0')
			return t, nil
		case "re":
			t.kind = regex
			return t, t.compileRegex(tk, value)
		case "w":
			if value == "" {
				return bad("searchNeedsWord", "needs a word, e.g. w:dog")
			}
			t.kind = word
			re, _, _ := wildcard(t.prepare(value), wordChar, wordChar+"*")
			t.re = regexp.MustCompile(`(?s)(?:^|[^\pL\pM\pN_])` + re + `(?:[^\pL\pM\pN_]|$)`)
			return t, nil
		case "added", "edited", "rated", "introduced", "resched":
			return parseEvent(t, key, v, o, bad)
		case "prop":
			return parseProp(t, v, o, bad)
		case "nc":
			t.fold = true
		case "sc":
			t.cloze = true
		default:
			t.kind = field
			t.name = key
			if strings.Contains(key, "*") {
				parts := strings.Split(key, "*")
				for i := range parts {
					parts[i] = regexp.QuoteMeta(parts[i])
				}
				t.nameRE = regexp.MustCompile("^" + strings.Join(parts, ".*") + "$")
			}
			if expr, ok := cutPrefixFold(value, "re:"); ok {
				t.raw = true
				return t, t.compileRegex(tk, expr)
			}
			re, lit, wild := wildcard(t.prepare(value), ".", ".*")
			if wild {
				t.re = regexp.MustCompile("(?s)^" + re + "$")
			} else {
				t.value = lit
			}
			return t, nil
		}
	}
	t.kind = text
	if value == "" {
		return bad("searchEmptyTerm", "empty search term")
	}
	re, lit, wild := wildcard(t.prepare(value), ".", ".*")
	if wild {
		t.re = regexp.MustCompile("(?s)" + re)
	} else {
		t.value = lit
	}
	return t, nil
}

// parseEvent parses the value of an event term with key: a number of days,
// and for rated: an optional rating after a colon.
func parseEvent(t term, key, v string, o Options, bad func(code, format string, args ...any) (term, error)) (term, error) {
	t.kind = map[string]kind{"added": added, "edited": edited, "rated": rated, "introduced": introduced, "resched": resched}[key]
	days, rating, hasRating := strings.Cut(v, ":")
	if hasRating && key != "rated" {
		return bad("searchNeedsDays", "needs a number of days, e.g. %s:7", key)
	}
	n, err := strconv.Atoi(days)
	if err != nil || n < 1 {
		return bad("searchNeedsDaysToday", "needs a number of days, e.g. %s:7 (1 is today)", key)
	}
	if key == "rated" && n > maxRatedDays {
		return bad("searchTooManyDays", "the number of days must be at most %d", maxRatedDays)
	}
	if hasRating {
		r, err := strconv.Atoi(rating)
		if err != nil || r < 1 || r > 4 {
			return bad("searchRating", "the rating must be 1 (Again), 2 (Hard), 3 (Good) or 4 (Easy), e.g. rated:7:1")
		}
		t.rating = r
	}
	t.since, t.until = dayStart(o, 1-n), math.MaxInt64
	return t, nil
}

// parseProp parses the value of a prop: term, e.g. "ivl>=30" or
// "rated>=-3".
func parseProp(t term, v string, o Options, bad func(code, format string, args ...any) (term, error)) (term, error) {
	name, rest := v, ""
	if i := strings.IndexAny(v, "=!<>"); i >= 0 {
		name, rest = v[:i], v[i:]
	}
	var cmp string
	for _, c := range []string{"!=", "<=", ">=", "=", "<", ">"} {
		if strings.HasPrefix(rest, c) {
			cmp = c
			break
		}
	}
	switch name {
	case "rated":
		t.kind = propRated
	case "resched":
		t.kind = propResched
	case "ease":
		return bad("searchNoEase", "Cardflip schedules with FSRS, which has no ease; search the difficulty with prop:d instead, e.g. prop:d>0.7")
	default:
		p, ok := properties[name]
		if !ok {
			return bad("searchUnknownProperty", "unknown property; use prop: with ivl, due, reps, lapses, pos, s, d, r, rated or resched, e.g. prop:ivl>=30")
		}
		n, err := strconv.ParseFloat(rest[len(cmp):], 64)
		if cmp == "" || err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
			return bad("searchNeedsComparison", "needs a comparison (=, !=, <, >, <= or >=) with a number, e.g. prop:%s%s", name, p.example)
		}
		t.kind, t.prop, t.cmp, t.num = propNum, p.p, cmp, n
		return t, nil
	}
	d, err := strconv.Atoi(rest[len(cmp):])
	if cmp == "" || err != nil {
		return bad("searchNeedsDayComparison", "needs a comparison with a number of days, e.g. prop:%s=-1 (0 is today, -1 yesterday)", name)
	}
	// Day d is [dayStart(d), dayStart(d+1)).
	lo, hi := int64(math.MinInt64), int64(math.MaxInt64)
	switch cmp {
	case "=", "!=":
		lo, hi = dayStart(o, d), dayStart(o, d+1)
		t.outside = cmp == "!="
	case "<":
		hi = dayStart(o, d)
	case "<=":
		hi = dayStart(o, d+1)
	case ">":
		lo = dayStart(o, d+1)
	case ">=":
		lo = dayStart(o, d)
	}
	t.since, t.until = lo, hi
	return t, nil
}

// dayStart returns when the study day offset days from today began, in Unix
// milliseconds; offset 0 is today, -1 yesterday.
func dayStart(o Options, offset int) int64 {
	now := o.Now
	if now.IsZero() {
		now = time.Now()
	}
	y, m, d := scheduler.DayStart(now, o.RolloverHour).Date()
	return time.Date(y, m, d+offset, o.RolloverHour, 0, 0, 0, now.Location()).UnixMilli()
}

// wordChar matches one character of a word.
const wordChar = `[\pL\pM\pN_]`

// prepare lower-cases a text value and, when the term ignores accents,
// removes them, as the card's text is.
func (t *term) prepare(s string) string {
	s = strings.ToLower(s)
	if t.fold {
		s = foldAccents(s)
	}
	return s
}

// compileRegex sets t.re from a regular expression, which ignores case
// unless it starts with "(?-i)". Escapes are passed to the expression as
// typed, which reads "\(" and the like as the literal character too.
func (t *term) compileRegex(tk *token, expr string) error {
	if expr == "" {
		return newError(tk.raw, tk.pos, "searchNeedsRegex", "needs a regular expression, e.g. re:^\\d+$")
	}
	re, err := regexp.Compile("(?i)" + expr)
	if err != nil {
		msg := strings.TrimPrefix(err.Error(), "error parsing regexp: ")
		return newError(tk.raw, tk.pos, "searchInvalidRegex", "invalid regular expression: %s", msg)
	}
	t.re = re
	return nil
}

func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):], true
	}
	return s, false
}

// unescape resolves the escapes in escapable; other backslashes stay as
// typed.
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && strings.IndexByte(escapable, s[i+1]) >= 0 {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// wildcard converts a text value to a regular expression in which "_" is
// one and "*" is run. A backslash makes "_", "*" and the escapable
// characters literal and is otherwise kept as typed. lit is the value with
// escapes resolved, and wild reports whether it had a wildcard.
func wildcard(s, one, run string) (re, lit string, wild bool) {
	var r, l strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s) && strings.IndexByte(escapable+"_*", s[i+1]) >= 0:
			i++
			c = s[i]
		case c == '_':
			r.WriteString(one)
			wild = true
			continue
		case c == '*':
			r.WriteString(run)
			wild = true
			continue
		}
		// Bytes of a multi-byte character are never special, so copying
		// them one at a time keeps them intact.
		r.WriteString(regexp.QuoteMeta(string([]byte{c})))
		l.WriteByte(c)
	}
	return r.String(), l.String(), wild
}

// foldAccents removes accents and other combining marks: it decomposes s
// (NFD) and drops the non-spacing marks.
func foldAccents(s string) string {
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		return s
	}
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Empty reports whether the query has no terms.
func (q Query) Empty() bool { return q.root == nil }

// UsesReviews reports whether matching looks at Card.Reviews, so callers
// can skip filling them otherwise.
func (q Query) UsesReviews() bool {
	return q.uses(rated, introduced, resched, propRated, propResched)
}

// UsesSiblingBurial reports whether matching looks at Card.BuriedSibling,
// so callers can skip computing it otherwise.
func (q Query) UsesSiblingBurial() bool { return q.uses(isBuried, isBuriedSibling) }

// uses reports whether the query has a term of one of kinds.
func (q Query) uses(kinds ...kind) bool { return q.root != nil && q.root.uses(kinds) }

func (n *node) uses(kinds []kind) bool {
	if n.op == opTerm {
		return slices.Contains(kinds, n.term.kind)
	}
	for _, ch := range n.children {
		if ch.uses(kinds) {
			return true
		}
	}
	return false
}

// excluded reports whether unqualified terms skip field i.
func (c *Card) excluded(i int) bool { return i < len(c.Excluded) && c.Excluded[i] }

// Match reports whether c matches the query.
func (q Query) Match(c *Card) bool {
	return q.root == nil || q.root.match(c)
}

func (n *node) match(c *Card) bool {
	var m bool
	switch n.op {
	case opTerm:
		m = (&n.term).match(c)
	case opAnd:
		m = true
		for _, ch := range n.children {
			if !ch.match(c) {
				m = false
				break
			}
		}
	case opOr:
		for _, ch := range n.children {
			if ch.match(c) {
				m = true
				break
			}
		}
	}
	return m != n.negate
}

func (t *term) match(c *Card) bool {
	switch t.kind {
	case text, word:
		for i, f := range c.Fields {
			if !c.excluded(i) && t.matchText(f) {
				return true
			}
		}
		return false
	case field:
		for i, name := range c.FieldNames {
			if i >= len(c.Fields) || !t.matchName(name) {
				continue
			}
			if t.raw {
				if i < len(c.Raw) && t.re.MatchString(c.Raw[i]) {
					return true
				}
			} else if t.matchText(c.Fields[i]) {
				return true
			}
		}
		return false
	case regex:
		for i, f := range c.Raw {
			if !c.excluded(i) && t.re.MatchString(f) {
				return true
			}
		}
		return false
	case tagRegex:
		for _, tg := range c.Tags {
			if t.re.MatchString(tg) {
				return true
			}
		}
		return false
	case deck:
		return nameOrChild(c.Deck, t.value) || c.Filtered && nameOrChild(c.HomeDeck, t.value)
	case tag:
		for _, tg := range c.Tags {
			if nameOrChild(tg, t.value) {
				return true
			}
		}
		return false
	case note:
		return strings.EqualFold(c.NoteType, t.value)
	case preset:
		return strings.EqualFold(c.Preset, t.value)
	case isNew:
		return c.State == New
	case isDue:
		return c.Due
	case isReview:
		return c.State == Review || c.State == Relearning
	case isLearn:
		return c.State == Learning || c.State == Relearning
	case isSuspended:
		return c.Suspended
	case isBuried:
		return c.BuriedManually || c.BuriedSibling
	case isBuriedManually:
		return c.BuriedManually
	case isBuriedSibling:
		return c.BuriedSibling
	case propNum:
		return compare(t.prop.of(c), t.cmp, t.num)
	case cardOrd:
		return float64(c.CardOrd) == t.num
	case cardName:
		name := strings.ToLower(c.CardName)
		if t.re != nil {
			return t.re.MatchString(name)
		}
		return name == t.value
	case tagNone:
		return len(c.Tags) == 0
	case deckCurrent:
		in := func(name string) bool {
			d := strings.ToLower(name)
			return d == t.value || strings.HasPrefix(d, t.value+"::")
		}
		return t.value != "" && (in(c.Deck) || c.Filtered && in(c.HomeDeck))
	case deckFiltered:
		return c.Filtered
	case noteID:
		return t.ids[c.NoteID]
	case cardID:
		return t.ids[c.CardID]
	case flag:
		return c.Flag == t.flag
	case added:
		return t.inRange(c.Added.UnixMilli())
	case edited:
		return !c.Edited.IsZero() && t.inRange(c.Edited.UnixMilli())
	case rated, resched:
		// Newest first: entries before the range end the search.
		for i := len(c.Reviews) - 1; i >= 0; i-- {
			r := c.Reviews[i]
			if r.Time < t.since {
				break
			}
			if t.kind == resched && r.Rating == 0 || t.kind == rated && r.Rating > 0 && (t.rating == 0 || r.Rating == t.rating) {
				return true
			}
		}
		return false
	case introduced:
		for _, r := range c.Reviews {
			if r.Rating > 0 {
				return t.inRange(r.Time)
			}
		}
		return false
	case propRated, propResched:
		for i := len(c.Reviews) - 1; i >= 0; i-- {
			if r := c.Reviews[i]; (r.Rating > 0) == (t.kind == propRated) {
				return t.inRange(r.Time)
			}
		}
		return false
	}
	return false
}

// compare reports whether v cmp n holds; a NaN v (a missing property) never
// matches.
func compare(v float64, cmp string, n float64) bool {
	if math.IsNaN(v) {
		return false
	}
	switch cmp {
	case "=":
		return v == n
	case "!=":
		return v != n
	case "<":
		return v < n
	case "<=":
		return v <= n
	case ">":
		return v > n
	case ">=":
		return v >= n
	}
	return false
}

// inRange reports whether an event term matches a time in Unix
// milliseconds.
func (t *term) inRange(ms int64) bool {
	return (ms >= t.since && ms < t.until) != t.outside
}

// matchText matches one field's text: a substring for text terms, the
// whole field for field terms.
func (t *term) matchText(s string) bool {
	if t.cloze {
		s = render.StripCloze(s)
	}
	if t.fold {
		s = foldAccents(s)
	}
	switch {
	case t.re != nil:
		return t.re.MatchString(s)
	case t.kind == field:
		return s == t.value
	}
	return strings.Contains(s, t.value)
}

func (t *term) matchName(name string) bool {
	if t.nameRE != nil {
		return t.nameRE.MatchString(name)
	}
	return name == t.name
}

// nameOrChild matches a "::" hierarchical name against want or a child of
// it, ignoring case. A trailing "*" in want matches any suffix.
func nameOrChild(name, want string) bool {
	name = strings.ToLower(name)
	if prefix, ok := strings.CutSuffix(want, "*"); ok {
		return strings.HasPrefix(name, prefix)
	}
	return name == want || strings.HasPrefix(name, want+"::")
}
