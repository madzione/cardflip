package search

import (
	"cardflip/golang/internal/usermsg"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func TestParseAndMatch(t *testing.T) {
	cards := map[string]*Card{
		"spanish verb":    {Fields: split("comer\x1fto eat"), Deck: "Spanish", Tags: []string{"verb"}, NoteType: "Basic", State: Review, Due: true, Flag: 1, Preset: "Languages"},
		"subdeck verb":    {Fields: split("ir\x1fto go"), Deck: "Spanish::Irregular", Tags: []string{"Verb"}, NoteType: "Basic", State: New, Preset: "Default"},
		"suspended verb":  {Fields: split("ser\x1fto be"), Deck: "Spanish", Tags: []string{"verb"}, NoteType: "Basic", State: Review, Suspended: true, Flag: 7},
		"spanish noun":    {Fields: split("perro\x1fdog"), Deck: "Spanish", Tags: []string{"noun"}, NoteType: "Basic", State: Learning, Due: true},
		"other deck verb": {Fields: split("manger\x1fto eat"), Deck: "Spanish Extra", Tags: []string{"verb"}, NoteType: "Basic", State: Relearning, Preset: "Languages"},
		"cloze":           {Fields: split("{{c1::paris}} is in france\x1f"), Deck: "Geo", Tags: []string{"lang::fr"}, NoteType: "Cloze", State: New, Preset: "Exam prep"},
	}
	cases := []struct {
		query string
		want  []string
	}{
		// Combined search (spec): unsuspended cards in Spanish or its
		// subdecks whose note has tag verb.
		{"deck:Spanish tag:verb -is:suspended", []string{"spanish verb", "subdeck verb"}},
		{"to eat", []string{"spanish verb", "other deck verb"}},
		{`"to eat"`, []string{"spanish verb", "other deck verb"}},
		{`"is in"`, []string{"cloze"}},
		{"EAT", []string{"spanish verb", "other deck verb"}},
		{"-eat", []string{"subdeck verb", "suspended verb", "spanish noun", "cloze"}},
		{`deck:"spanish extra"`, []string{"other deck verb"}},
		{"deck:span*", []string{"spanish verb", "subdeck verb", "suspended verb", "spanish noun", "other deck verb"}},
		{"tag:lang", []string{"cloze"}},
		{"note:cloze", []string{"cloze"}},
		{"is:new", []string{"subdeck verb", "cloze"}},
		{"is:due", []string{"spanish verb", "spanish noun"}},
		{"is:review", []string{"spanish verb", "suspended verb", "other deck verb"}},
		{"is:learn", []string{"spanish noun", "other deck verb"}},
		{"is:suspended", []string{"suspended verb"}},
		// Search red flags (spec).
		{"flag:1", []string{"spanish verb"}},
		{"flag:7", []string{"suspended verb"}},
		{"-flag:1", []string{"subdeck verb", "suspended verb", "spanish noun", "other deck verb", "cloze"}},
		// Unflagged cards (spec).
		{"deck:Spanish flag:0", []string{"subdeck verb", "spanish noun"}},
		// Search by preset (spec), ignoring case; quotes allow spaces.
		{"preset:Languages", []string{"spanish verb", "other deck verb"}},
		{"preset:LANGUAGES", []string{"spanish verb", "other deck verb"}},
		{`preset:"exam prep"`, []string{"cloze"}},
		{"preset:exam", nil},
		{"-preset:languages", []string{"subdeck verb", "suspended verb", "spanish noun", "cloze"}},
		{"", []string{"spanish verb", "subdeck verb", "suspended verb", "spanish noun", "other deck verb", "cloze"}},
		{"http://x", nil},
	}
	order := []string{"spanish verb", "subdeck verb", "suspended verb", "spanish noun", "other deck verb", "cloze"}
	for _, c := range cases {
		q, err := Parse(c.query)
		if err != nil {
			t.Fatalf("%q: %v", c.query, err)
		}
		var got []string
		for _, name := range order {
			if q.Match(cards[name]) {
				got = append(got, name)
			}
		}
		if len(got) != len(c.want) {
			t.Errorf("%q matched %v, want %v", c.query, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q matched %v, want %v", c.query, got, c.want)
				break
			}
		}
	}
}

func TestInvalidQualifier(t *testing.T) {
	// Invalid qualifier (spec): an error for that term, not an empty list.
	for _, q := range []string{"is:bogus", "verb is:bogus", "-is:nope", "deck:", `tag:""`, "preset:", `preset:""`, "flag:8", "flag:", "flag:red", "flag:11"} {
		_, err := Parse(q)
		var e *Error
		if !errors.As(err, &e) {
			t.Fatalf("%q: err = %v", q, err)
		}
		t.Logf("%q → %v", q, err)
	}
	if _, err := Parse("is:bogus"); err.(*Error).Term != "is:bogus" {
		t.Fatal("error does not name the term")
	}
}

func TestTokenize(t *testing.T) {
	type tok struct {
		kind   tokKind
		text   string
		colon  int
		negate bool
	}
	w := func(text string) tok { return tok{kind: tWord, text: text, colon: -1} }
	cases := []struct {
		query string
		want  []tok
	}{
		{`a  b`, []tok{w("a"), w("b")}},
		{`"a (b)"`, []tok{w("a (b)")}},
		{`"a b"c`, []tok{w("a bc")}},
		{`deck:"My Deck"`, []tok{{kind: tWord, text: "deck:My Deck", colon: 4}}},
		// A quoted colon still starts a qualifier.
		{`"deck:x"`, []tok{{kind: tWord, text: "deck:x", colon: 4}}},
		{`"sc:a b"`, []tok{{kind: tWord, text: "sc:a b", colon: 2}}},
		{`deck:"a:b"`, []tok{{kind: tWord, text: "deck:a:b", colon: 4}}},
		{`"unclosed quote`, []tok{w("unclosed quote")}},
		// Escapes are kept for the term to resolve; an escaped quote or
		// colon is literal.
		{`\"a\"`, []tok{w(`\"a\"`)}},
		{`"say \"hi\""`, []tok{w(`say \"hi\"`)}},
		{`f\(x\)`, []tok{w(`f\(x\)`)}},
		{`\-ing`, []tok{w(`\-ing`)}},
		{`a\:b`, []tok{w(`a\:b`)}},
		{`a\\b`, []tok{w(`a\\b`)}},
		{`a\\:b`, []tok{{kind: tWord, text: `a\\:b`, colon: 3}}},
		{`a\b`, []tok{w(`a\b`)}},
		{`a\`, []tok{w(`a\`)}},
		// Negation and parentheses.
		{`-a`, []tok{{kind: tWord, text: "a", colon: -1, negate: true}}},
		{`-"a b"`, []tok{{kind: tWord, text: "a b", colon: -1, negate: true}}},
		{`- a`, []tok{w("-"), w("a")}},
		{`well-known`, []tok{w("well-known")}},
		{`-(a)`, []tok{{kind: tOpen, colon: 0, negate: true}, w("a"), {kind: tClose}}},
		{`(a b)c`, []tok{{kind: tOpen}, w("a"), w("b"), {kind: tClose}, w("c")}},
		{`f(x)`, []tok{w("f"), {kind: tOpen}, w("x"), {kind: tClose}}},
		// Operators, but not when quoted or negated.
		{`a OR b And c`, []tok{w("a"), {kind: tOr, text: "OR", colon: -1}, w("b"), {kind: tAnd, text: "And", colon: -1}, w("c")}},
		{`"or" -or`, []tok{w("or"), {kind: tWord, text: "or", colon: -1, negate: true}}},
	}
	for _, c := range cases {
		var got []tok
		for _, t := range tokenize(c.query) {
			got = append(got, tok{kind: t.kind, text: t.text, colon: t.colon, negate: t.negate})
		}
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q → %v, want %v", c.query, got, c.want)
		}
	}
}

func TestBoolean(t *testing.T) {
	cards := map[string]*Card{
		"es verb": {Fields: split("comer\x1fto eat"), Deck: "Spanish", Tags: []string{"verb"}, Due: true},
		"es noun": {Fields: split("perro\x1fdog"), Deck: "Spanish", Tags: []string{"noun"}},
		"fr verb": {Fields: split("manger\x1fto eat"), Deck: "French", Tags: []string{"verb"}, Due: true},
		"de adj":  {Fields: split("schnell\x1ffast"), Deck: "German", Tags: []string{"adj"}, Due: true},
		"math":    {Fields: split("f(x) = x²\x1f-ing"), Deck: "A", Tags: []string{"x"}},
		"a y":     {Fields: split("y\x1f"), Deck: "A", Tags: []string{"y"}},
		"b x":     {Fields: split("\x1f"), Deck: "B", Tags: []string{"x"}},
	}
	order := []string{"es verb", "es noun", "fr verb", "de adj", "math", "a y", "b x"}
	cases := []struct {
		query string
		want  []string
	}{
		// Either deck (spec).
		{"(deck:Spanish or deck:French) is:due", []string{"es verb", "fr verb"}},
		// Negated group (spec).
		{"-(tag:verb or tag:noun)", []string{"de adj", "math", "a y", "b x"}},
		// Precedence (spec): AND binds more tightly than OR.
		{"deck:A tag:x or tag:y", []string{"math", "a y"}},
		{"deck:A and tag:x or tag:y", []string{"math", "a y"}},
		{"tag:y or deck:A tag:x", []string{"math", "a y"}},
		{"deck:A (tag:x or tag:y)", []string{"math", "a y"}},
		{"(deck:A or deck:B) tag:x", []string{"math", "b x"}},
		{"tag:x OR tag:noun Or tag:adj", []string{"es noun", "de adj", "math", "b x"}},
		{"deck:spanish AND tag:verb", []string{"es verb"}},
		// Nested groups.
		{"((deck:spanish or deck:french) (tag:verb or tag:adj)) or deck:b", []string{"es verb", "fr verb", "b x"}},
		{"-(deck:spanish -(tag:noun))", []string{"es noun", "fr verb", "de adj", "math", "a y", "b x"}},
		{"-(-tag:x)", []string{"math", "b x"}},
		{"((tag:x))", []string{"math", "b x"}},
		{"-(deck:a or deck:b) -(is:due)", []string{"es noun"}},
		// Literal parenthesis and leading hyphen (spec).
		{`"f(x)"`, []string{"math"}},
		{`f\(x\)`, []string{"math"}},
		{`\-ing`, []string{"math"}},
		{`-ing`, []string{"es verb", "es noun", "fr verb", "de adj", "a y", "b x"}},
		// Quoted operators are words.
		{`"or"`, nil},
	}
	for _, c := range cases {
		q, err := Parse(c.query)
		if err != nil {
			t.Fatalf("%q: %v", c.query, err)
		}
		var got []string
		for _, name := range order {
			if q.Match(cards[name]) {
				got = append(got, name)
			}
		}
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q matched %v, want %v", c.query, got, c.want)
		}
	}
}

func TestSyntaxErrors(t *testing.T) {
	cases := []struct {
		query string
		pos   int
		msg   string
	}{
		// Unclosed group (spec).
		{"(deck:Spanish or deck:French", 1, "a parenthesis is not closed"},
		{"a (b (c)", 3, "a parenthesis is not closed"},
		{"(", 1, "a parenthesis is not closed"},
		{"-(a", 1, "a parenthesis is not closed"},
		{"a)", 2, "a closing parenthesis has no opening one"},
		{"(a)) b", 4, "a closing parenthesis has no opening one"},
		{")", 1, "a closing parenthesis has no opening one"},
		{"a or", 3, `"or" needs a search term on both sides`},
		{"or a", 1, `"or" needs a search term on both sides`},
		{"a OR or b", 3, `"or" needs a search term on both sides`},
		{"(a or) b", 4, `"or" needs a search term on both sides`},
		{"a and", 3, `"and" needs a search term on both sides`},
		{"And a", 1, `"and" needs a search term on both sides`},
		{"a and or b", 3, `"and" needs a search term on both sides`},
		{"a ()", 3, "empty parentheses"},
		{"-( )", 1, "empty parentheses"},
		{"(())", 2, "empty parentheses"},
	}
	for _, c := range cases {
		_, err := Parse(c.query)
		var e *Error
		if !errors.As(err, &e) {
			t.Errorf("%q: err = %v", c.query, err)
			continue
		}
		if e.Term != "" || e.Pos != c.pos || e.Msg != c.msg {
			t.Errorf("%q: %+v, want pos %d %q", c.query, *e, c.pos, c.msg)
		}
	}
	_, err := Parse("(deck:Spanish or deck:French")
	if got, want := usermsg.Plain(err.Error()), "a parenthesis is not closed (character 1)"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	// Term errors keep naming the term, now with its position.
	_, err = Parse("(a or is:bogus)")
	if e := err.(*Error); e.Term != "is:bogus" || e.Pos != 7 || usermsg.Plain(err.Error()) != "is:bogus: unknown state; use is:new, is:due, is:review, is:learn, is:suspended, is:buried, is:buried-manually or is:buried-sibling" {
		t.Errorf("term error = %+v", e)
	}
}

// split turns "\x1f"-separated text into fields.
func split(s string) []string { return strings.Split(s, "\x1f") }

// matchNames returns the names of the cards, in order, that query matches.
func matchNames(t *testing.T, query string, o Options, cards map[string]*Card, order []string) []string {
	t.Helper()
	q, err := ParseWith(query, o)
	if err != nil {
		t.Fatalf("%q: %v", query, err)
	}
	var got []string
	for _, name := range order {
		if q.Match(cards[name]) {
			got = append(got, name)
		}
	}
	return got
}

func TestWildcardsAndWords(t *testing.T) {
	cards := map[string]*Card{
		"dog":   {Fields: split("dog\x1fanimal")},
		"dig":   {Fields: split("dig\x1fverb")},
		"doggy": {Fields: split("doggy\x1fanimal")},
		"lit":   {Fields: split("snake_case\x1f2*3")},
		"cross": {Fields: split("d\x1fg")},
	}
	order := []string{"dog", "dig", "doggy", "lit", "cross"}
	cases := []struct {
		query string
		want  []string
	}{
		// Single-character wildcard (spec).
		{"d_g", []string{"dog", "dig", "doggy"}},
		{"d*g", []string{"dog", "dig", "doggy"}},
		{"do*y", []string{"doggy"}},
		{"*", []string{"dog", "dig", "doggy", "lit", "cross"}},
		// Escaped wildcards are literal.
		{`e\_c`, []string{"lit"}},
		{`2\*3`, []string{"lit"}},
		{`d\_g`, nil},
		{"e_c", []string{"lit"}},
		// Whole word (spec).
		{"w:dog", []string{"dog"}},
		{"w:dog*", []string{"dog", "doggy"}},
		{"w:*ggy", []string{"doggy"}},
		{"w:og", nil},
		{"w:case", nil},
		{"w:snake_case", []string{"lit"}},
		{"-w:dog", []string{"dig", "doggy", "lit", "cross"}},
		{`"w:dog animal"`, nil},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, Options{}, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q matched %v, want %v", c.query, got, c.want)
		}
	}
}

func TestFieldSearch(t *testing.T) {
	names := []string{"front", "back"}
	card := func(front, back string) *Card {
		return &Card{Fields: []string{front, back}, Raw: []string{front, back}, FieldNames: names}
	}
	cards := map[string]*Card{
		"dog":    card("dog", "chien"),
		"a dog":  card("a dog", ""),
		"digits": card("123", "x"),
		"other":  {Fields: []string{"dog"}, Raw: []string{"dog"}, FieldNames: []string{"text"}},
		"amp":    {Fields: []string{"salt & pepper", "x"}, Raw: []string{"salt &amp; pepper", "x"}, FieldNames: names},
	}
	order := []string{"dog", "a dog", "digits", "other", "amp"}
	cases := []struct {
		query string
		want  []string
	}{
		// Exact field (spec).
		{"front:dog", []string{"dog"}},
		{"FRONT:DOG", []string{"dog"}},
		{"front:*dog*", []string{"dog", "a dog"}},
		{"front:d_g", []string{"dog"}},
		{`"front:a dog"`, []string{"a dog"}},
		{`front:"a dog"`, []string{"a dog"}},
		// Field name wildcards.
		{"fr*:dog", []string{"dog"}},
		{"*:dog", []string{"dog", "other"}},
		{"*t:dog", []string{"dog", "other"}},
		// Empty field (spec), non-empty, and any note with the field.
		{"back:", []string{"a dog"}},
		{"back:_*", []string{"dog", "digits", "amp"}},
		{"back:*", []string{"dog", "a dog", "digits", "amp"}},
		{"-back:*", []string{"other"}},
		// Unknown field matches nothing.
		{"extra:dog", nil},
		// Entity decoding (spec), in text and field terms.
		{`"salt & pepper"`, []string{"amp"}},
		{`"front:salt & pepper"`, []string{"amp"}},
		{"amp", nil},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, Options{}, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q matched %v, want %v", c.query, got, c.want)
		}
	}
}

func TestExcludedFields(t *testing.T) {
	names := []string{"word", "notes"}
	cards := map[string]*Card{
		"excluded": {Fields: []string{"go", "irregular"}, Raw: []string{"go", "<i>irregular</i>"}, FieldNames: names,
			Excluded: []bool{false, true}},
		"included": {Fields: []string{"be", "irregular"}, Raw: []string{"be", "<i>irregular</i>"}, FieldNames: names},
		"both": {Fields: []string{"irregular", "irregular"}, Raw: []string{"irregular", "irregular"}, FieldNames: names,
			Excluded: []bool{false, true}},
	}
	order := []string{"excluded", "included", "both"}
	cases := []struct {
		query string
		want  []string
	}{
		// Notes field excluded (spec): free text skips it unless another
		// field matches.
		{"irregular", []string{"included", "both"}},
		{"irreg*", []string{"included", "both"}},
		{"w:irregular", []string{"included", "both"}},
		{"re:<i>", []string{"included"}},
		{"nc:irregular", []string{"included", "both"}},
		// Searches naming the field still look in it.
		{"notes:irregular", []string{"excluded", "included", "both"}},
		{"*:irregular", []string{"excluded", "included", "both"}},
		{"notes:re:<i>", []string{"excluded", "included"}},
		{"go", []string{"excluded"}},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, Options{}, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q matched %v, want %v", c.query, got, c.want)
		}
	}
}

func TestAccentsAndCloze(t *testing.T) {
	cards := map[string]*Card{
		"über":  {Fields: split("über alles\x1f"), FieldNames: []string{"front", "back"}},
		"café":  {Fields: split("café\x1f"), FieldNames: []string{"front", "back"}},
		"nfd":   {Fields: split("café\x1f"), FieldNames: []string{"front", "back"}},
		"cloze": {Fields: split("{{c1::capital::city}} of {{c2,3::france}}\x1fextra"), FieldNames: []string{"text", "extra"}},
		"plain": {Fields: split("uber\x1f"), FieldNames: []string{"front", "back"}},
	}
	order := []string{"über", "café", "nfd", "cloze", "plain"}
	cases := []struct {
		query  string
		accent bool
		want   []string
	}{
		// Ignore accents (spec).
		{"nc:uber", false, []string{"über", "plain"}},
		{"nc:über", false, []string{"über", "plain"}},
		{"nc:cafe", false, []string{"café", "nfd"}},
		{"nc:ub*es", false, []string{"über"}},
		{"uber", false, []string{"plain"}},
		{"cafe", false, []string{"nfd"}},
		// Setting on (spec): every text and field term ignores accents.
		{"cafe", true, []string{"café", "nfd"}},
		{"front:cafe", true, []string{"café", "nfd"}},
		{"w:uber", true, []string{"über", "plain"}},
		{"front:cafe", false, nil},
		// Across a cloze (spec).
		{`"sc:capital of france"`, false, []string{"cloze"}},
		{`"capital of france"`, false, nil},
		{"sc:city", false, nil},
		{"city", false, []string{"cloze"}},
		{`"sc:of fr*"`, false, []string{"cloze"}},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, Options{IgnoreAccents: c.accent}, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q (ignore accents %v) matched %v, want %v", c.query, c.accent, got, c.want)
		}
	}
}

func TestRegexSearch(t *testing.T) {
	names := []string{"front", "back"}
	cards := map[string]*Card{
		"digits": {Fields: []string{"123", "x"}, Raw: []string{"123", "x"}, FieldNames: names, Tags: []string{"Num::Int"}},
		"mixed":  {Fields: []string{"12a", "456"}, Raw: []string{"12a", "456"}, FieldNames: names, Tags: []string{"num"}},
		"html":   {Fields: []string{"dog", "x"}, Raw: []string{"<b>Dog</b>", "x"}, FieldNames: names, Tags: []string{"animal"}},
	}
	order := []string{"digits", "mixed", "html"}
	cases := []struct {
		query string
		want  []string
	}{
		// Field regex (spec).
		{`front:re:^\d+$`, []string{"digits"}},
		{`re:^\d+$`, []string{"digits", "mixed"}},
		// Matched against the stored HTML, ignoring case unless (?-i).
		{"re:<b>dog", []string{"html"}},
		{`"re:(?-i)<b>dog"`, nil},
		{`"re:(?-i)<b>Dog"`, []string{"html"}},
		{`"front:re:^(dog|\d+)$"`, []string{"digits"}},
		{`re:^x\)?$`, []string{"digits", "html"}},
		// Tags.
		{"tag:re:^num", []string{"digits", "mixed"}},
		{"tag:re:int$", []string{"digits"}},
		{`"tag:re:(?-i)^num$"`, []string{"mixed"}},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, Options{}, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q matched %v, want %v", c.query, got, c.want)
		}
	}
	// An invalid expression is an error naming the term.
	for _, q := range []string{"re:[a", `"front:re:("`, `"tag:re:a{2,1}"`, "re:", "w:", "nc:", "sc:"} {
		_, err := Parse(q)
		var e *Error
		if !errors.As(err, &e) || e.Term != q {
			t.Errorf("%q: err = %v", q, err)
			continue
		}
		t.Logf("%q → %v", q, err)
	}
}

func TestUnescape(t *testing.T) {
	for in, want := range map[string]string{`a\:b`: "a:b", `\\`: `\`, `\_`: `\_`, `\\\_`: `\\_`, `\"x\"`: `"x"`, `a\`: `a\`} {
		if got := unescape(in); got != want {
			t.Errorf("unescape(%q) = %q, want %q", in, got, want)
		}
	}
	if _, lit, wild := wildcard(`a\\_b\_c\(`, ".", ".*"); lit != `a\b_c(` || !wild {
		t.Errorf("wildcard lit = %q wild=%v", lit, wild)
	}
}

func TestEventTerms(t *testing.T) {
	// Today at 10:00 with rollover 04:00: today is the study day that began
	// at 04:00, yesterday 02:00 belongs to the study day before yesterday.
	loc := time.FixedZone("test", 2*3600)
	now := time.Date(2026, 10, 5, 10, 0, 0, 0, loc)
	at := func(day, hour int) time.Time { return time.Date(2026, 10, day, hour, 0, 0, 0, loc) }
	ms := func(day, hour int) int64 { return at(day, hour).UnixMilli() }
	o := Options{Now: now, RolloverHour: 4}
	cards := map[string]*Card{
		"added today":   {Added: at(5, 4), Edited: at(5, 4)},
		"added 02:00":   {Added: at(5, 2), Edited: at(5, 9)},
		"yesterday 2am": {Added: at(4, 2)},
		"failed today": {Added: at(1, 12), Reviews: []LogEntry{
			{Time: ms(2, 12), Rating: 3}, {Time: ms(5, 9), Rating: 1}}},
		"good yesterday": {Added: at(1, 12), Reviews: []LogEntry{
			{Time: ms(1, 12), Rating: 1}, {Time: ms(3, 12), Rating: 0}, {Time: ms(4, 12), Rating: 3}}},
		"rescheduled": {Added: at(1, 12), Reviews: []LogEntry{
			{Time: ms(1, 12), Rating: 0}, {Time: ms(5, 5), Rating: 0}}},
	}
	order := []string{"added today", "added 02:00", "yesterday 2am", "failed today", "good yesterday", "rescheduled"}
	cases := []struct {
		query string
		want  []string
	}{
		// Added today (spec); before rollover (spec).
		{"added:1", []string{"added today"}},
		{"added:2", []string{"added today", "added 02:00"}},
		{"added:3", []string{"added today", "added 02:00", "yesterday 2am"}},
		{"edited:1", []string{"added today", "added 02:00"}},
		{"rated:1", []string{"failed today"}},
		{"rated:2", []string{"failed today", "good yesterday"}},
		// Failed this week (spec).
		{"rated:7:1", []string{"failed today", "good yesterday"}},
		{"rated:2:1", []string{"failed today"}},
		{"rated:2:3", []string{"good yesterday"}},
		{"-rated:7", []string{"added today", "added 02:00", "yesterday 2am", "rescheduled"}},
		{"introduced:4", []string{"failed today"}},
		{"introduced:5", []string{"failed today", "good yesterday"}},
		{"resched:1", []string{"rescheduled"}},
		{"resched:3", []string{"good yesterday", "rescheduled"}},
		{"prop:rated=0", []string{"failed today"}},
		{"prop:rated=-1", []string{"good yesterday"}},
		{"prop:rated>-1", []string{"failed today"}},
		{"prop:rated>=-1", []string{"failed today", "good yesterday"}},
		{"prop:rated<0", []string{"good yesterday"}},
		{"prop:rated<=0", []string{"failed today", "good yesterday"}},
		{"prop:rated!=-1", []string{"failed today"}},
		{"prop:resched=0", []string{"rescheduled"}},
		{"prop:resched=-2", []string{"good yesterday"}},
		{"-prop:resched=0", []string{"added today", "added 02:00", "yesterday 2am", "failed today", "good yesterday"}},
		{"prop:RATED=0", []string{"failed today"}},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, o, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q = %v, want %v", c.query, got, c.want)
		}
	}
	// A rollover hour of 06:00 puts 05:00 today in yesterday.
	o.RolloverHour = 6
	if got := matchNames(t, "resched:1", o, cards, order); got != nil {
		t.Errorf("resched:1 with rollover 06:00 = %v", got)
	}
}

func TestEventTermErrors(t *testing.T) {
	cases := map[string]string{
		// Out of range (spec).
		"rated:400":         "rated:400: the number of days must be at most 365",
		"rated:0":           "rated:0: needs a number of days, e.g. rated:7 (1 is today)",
		"rated:-1":          "rated:-1: needs a number of days, e.g. rated:7 (1 is today)",
		"rated:7:0":         "rated:7:0: the rating must be 1 (Again), 2 (Hard), 3 (Good) or 4 (Easy), e.g. rated:7:1",
		"rated:7:":          "rated:7:: the rating must be 1 (Again), 2 (Hard), 3 (Good) or 4 (Easy), e.g. rated:7:1",
		"introduced:1:1":    "introduced:1:1: needs a number of days, e.g. introduced:7",
		"added:":            "added:: needs a number of days, e.g. added:7 (1 is today)",
		"prop:rated":        "prop:rated: needs a comparison with a number of days, e.g. prop:rated=-1 (0 is today, -1 yesterday)",
		"prop:foo>3":        "prop:foo>3: unknown property; use prop: with ivl, due, reps, lapses, pos, s, d, r, rated or resched, e.g. prop:ivl>=30",
		"prop:rated=x":      "prop:rated=x: needs a comparison with a number of days, e.g. prop:rated=-1 (0 is today, -1 yesterday)",
		"prop:resched==1":   "prop:resched==1: needs a comparison with a number of days, e.g. prop:resched=-1 (0 is today, -1 yesterday)",
		"deck:x rated:1000": "rated:1000: the number of days must be at most 365",
	}
	for query, want := range cases {
		_, err := Parse(query)
		if err == nil || usermsg.Plain(err.Error()) != want {
			t.Errorf("%q: err = %v, want %q", query, err, want)
		}
	}
	for _, query := range []string{"rated:365", "rated:1:4", "added:10000", "prop:rated>=-365"} {
		if _, err := Parse(query); err != nil {
			t.Errorf("%q: %v", query, err)
		}
	}
}

func TestUsesReviews(t *testing.T) {
	for query, want := range map[string]bool{
		"": false, "added:1 edited:2": false, "deck:x": false,
		"rated:1": true, "deck:x (a or -introduced:3)": true, "resched:1": true, "prop:rated=0": true, "prop:resched<0": true,
	} {
		q, err := Parse(query)
		if err != nil {
			t.Fatal(err)
		}
		if q.UsesReviews() != want {
			t.Errorf("%q: UsesReviews = %v", query, !want)
		}
	}
}

func TestPropertyTerms(t *testing.T) {
	nan := math.NaN()
	newCard := func(pos float64) *Card {
		return &Card{State: New, Interval: nan, DueDay: nan, Position: pos, Stability: nan, Difficulty: nan, Retrievability: nan}
	}
	cards := map[string]*Card{
		"new":     newCard(5),
		"young":   {State: Review, Interval: 3, DueDay: 1, Position: nan, Stability: 4.5, Difficulty: 0.3, Retrievability: 0.95, Reps: 4},
		"mature":  {State: Review, Interval: 30, DueDay: 12, Position: nan, Stability: 40, Difficulty: 0.5, Retrievability: 0.9, Reps: 9, Lapses: 1},
		"overdue": {State: Review, Interval: 60, DueDay: -3, Position: nan, Stability: 50, Difficulty: 0.8, Retrievability: 0.7, Reps: 12, Lapses: 4},
		"relearn": {State: Relearning, Interval: nan, DueDay: 0, Position: nan, Stability: 2, Difficulty: 0.9, Retrievability: 0.79, Reps: 15, Lapses: 5},
	}
	order := []string{"new", "young", "mature", "overdue", "relearn"}
	cases := []struct {
		query string
		want  []string
	}{
		// Long intervals (spec); New and relearning cards have no interval.
		{"prop:ivl>=30", []string{"mature", "overdue"}},
		{"prop:ivl<30", []string{"young"}},
		{"prop:ivl!=30", []string{"young", "overdue"}},
		{"-prop:ivl>=30", []string{"new", "young", "relearn"}},
		// Overdue (spec).
		{"prop:due<0", []string{"overdue"}},
		{"prop:due=0", []string{"relearn"}},
		{"prop:due>=1", []string{"young", "mature"}},
		{"prop:reps>10", []string{"overdue", "relearn"}},
		{"prop:reps=0", []string{"new"}},
		{"prop:lapses>=4", []string{"overdue", "relearn"}},
		{"prop:lapses<=1", []string{"new", "young", "mature"}},
		{"prop:pos=5", []string{"new"}},
		{"prop:pos>0", []string{"new"}},
		{"prop:s>21", []string{"mature", "overdue"}},
		{"prop:s<=4.5", []string{"young", "relearn"}},
		{"prop:d>0.7", []string{"overdue", "relearn"}},
		{"prop:D<0.4", []string{"young"}},
		// At risk (spec).
		{"prop:r<0.8", []string{"overdue", "relearn"}},
		{"prop:r>=.9", []string{"young", "mature"}},
		{"prop:ivl>=30 or prop:lapses>4", []string{"mature", "overdue", "relearn"}},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, Options{}, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q = %v, want %v", c.query, got, c.want)
		}
	}
}

func TestPropertyTermErrors(t *testing.T) {
	cases := map[string]string{
		"prop:ease>2":   "prop:ease>2: Cardflip schedules with FSRS, which has no ease; search the difficulty with prop:d instead, e.g. prop:d>0.7",
		"prop:ease":     "prop:ease: Cardflip schedules with FSRS, which has no ease; search the difficulty with prop:d instead, e.g. prop:d>0.7",
		"prop:factor>2": "prop:factor>2: unknown property; use prop: with ivl, due, reps, lapses, pos, s, d, r, rated or resched, e.g. prop:ivl>=30",
		"prop:":         "prop:: unknown property; use prop: with ivl, due, reps, lapses, pos, s, d, r, rated or resched, e.g. prop:ivl>=30",
		"prop:ivl":      "prop:ivl: needs a comparison (=, !=, <, >, <= or >=) with a number, e.g. prop:ivl>=30",
		"prop:ivl>=":    "prop:ivl>=: needs a comparison (=, !=, <, >, <= or >=) with a number, e.g. prop:ivl>=30",
		"prop:r<x":      "prop:r<x: needs a comparison (=, !=, <, >, <= or >=) with a number, e.g. prop:r<0.8",
		"prop:d=nan":    "prop:d=nan: needs a comparison (=, !=, <, >, <= or >=) with a number, e.g. prop:d>0.7",
		"prop:due==1":   "prop:due==1: needs a comparison (=, !=, <, >, <= or >=) with a number, e.g. prop:due<0",
	}
	for query, want := range cases {
		_, err := Parse(query)
		if err == nil || usermsg.Plain(err.Error()) != want {
			t.Errorf("%q: err = %v, want %q", query, err, want)
		}
	}
	for _, query := range []string{"prop:ivl=1", "prop:due>-7", "prop:s>=1.5", "prop:rated=-1"} {
		if _, err := Parse(query); err != nil {
			t.Errorf("%q: %v", query, err)
		}
	}
}

func TestBurialAndStateTerms(t *testing.T) {
	cards := map[string]*Card{
		"new":        {State: New},
		"learning":   {State: Learning},
		"review":     {State: Review},
		"lapsed":     {State: Relearning},
		"manually":   {State: Review, BuriedManually: true},
		"by sibling": {State: New, BuriedSibling: true},
	}
	order := []string{"new", "learning", "review", "lapsed", "manually", "by sibling"}
	cases := []struct {
		query string
		want  []string
	}{
		{"is:buried", []string{"manually", "by sibling"}},
		{"is:buried-manually", []string{"manually"}},
		{"is:buried-sibling", []string{"by sibling"}},
		{"-is:buried is:review", []string{"review", "lapsed"}},
		{"is:learn", []string{"learning", "lapsed"}},
		{"is:review", []string{"review", "lapsed", "manually"}},
		// Lapsed cards (spec).
		{"is:learn is:review", []string{"lapsed"}},
		{"is:review -is:learn", []string{"review", "manually"}},
		{"is:learn -is:review", []string{"learning"}},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, Options{}, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q = %v, want %v", c.query, got, c.want)
		}
	}
	if _, err := Parse("is:buried-today"); err == nil {
		t.Error("is:buried-today accepted")
	}
}

func TestUsesSiblingBurial(t *testing.T) {
	for query, want := range map[string]bool{
		"is:buried": true, "a or -is:buried-sibling": true, "is:buried-manually": false, "deck:x": false, "": false,
	} {
		q, err := Parse(query)
		if err != nil {
			t.Fatal(err)
		}
		if q.UsesSiblingBurial() != want {
			t.Errorf("%q: UsesSiblingBurial = %v", query, !want)
		}
	}
}

// Cards away from home (card-browser spec): deck:filtered matches the
// cards in filtered decks, and deck: matches them by their home deck too.
func TestFilteredDeckTerms(t *testing.T) {
	cards := map[string]*Card{
		"home":    {Deck: "Spanish", HomeDeck: "Spanish"},
		"away":    {Deck: "Cram", HomeDeck: "Spanish::Verbs", Filtered: true},
		"french":  {Deck: "Cram", HomeDeck: "French", Filtered: true},
		"literal": {Deck: "filtered", HomeDeck: "filtered"},
	}
	order := []string{"home", "away", "french", "literal"}
	o := Options{CurrentDeck: "Spanish"}
	for _, c := range []struct {
		query string
		want  []string
	}{
		{"deck:filtered", []string{"away", "french"}},
		{"-deck:filtered", []string{"home", "literal"}},
		{"deck:Spanish", []string{"home", "away"}},
		{"deck:Spanish -deck:filtered", []string{"home"}},
		{"deck:Cram", []string{"away", "french"}},
		{"deck:current", []string{"home", "away"}},
		{"deck:French", []string{"french"}},
	} {
		if got := matchNames(t, c.query, o, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q = %v, want %v", c.query, got, c.want)
		}
	}
}

func TestCardTagDeckAndIDTerms(t *testing.T) {
	cards := map[string]*Card{
		"forward": {Deck: "French", Tags: []string{"verb"}, CardOrd: 1, CardName: "Card 1", NoteID: 10, CardID: 101},
		"reverse": {Deck: "French", Tags: []string{"verb"}, CardOrd: 2, CardName: "Card 2", NoteID: 10, CardID: 102},
		"subdeck": {Deck: "French::Verbs", CardOrd: 1, CardName: "Recognition", NoteID: 11, CardID: 111},
		"cloze 2": {Deck: "French Extra", Tags: []string{"none"}, CardOrd: 2, CardName: "Cloze", NoteID: 12, CardID: 122},
		"spanish": {Deck: "Spanish", CardOrd: 1, CardName: "Card 1", NoteID: 13, CardID: 131},
	}
	order := []string{"forward", "reverse", "subdeck", "cloze 2", "spanish"}
	o := Options{CurrentDeck: "French"}
	cases := []struct {
		query string
		want  []string
	}{
		// Reverse cards (spec): card type 2 and cloze 2.
		{"card:2", []string{"reverse", "cloze 2"}},
		{"card:1", []string{"forward", "subdeck", "spanish"}},
		{`"card:card 2"`, []string{"reverse"}},
		{"card:cloze", []string{"cloze 2"}},
		{"card:rec*", []string{"subdeck"}},
		{"card:card*", []string{"forward", "reverse", "spanish"}},
		{"card:card_1", []string{"forward", "spanish"}},
		{"card:3", nil},
		// tag:none ignores a tag named "none".
		{"tag:none", []string{"subdeck", "spanish"}},
		{"-tag:none", []string{"forward", "reverse", "cloze 2"}},
		// The current deck with its subdecks, not decks sharing its prefix.
		{"deck:current", []string{"forward", "reverse", "subdeck"}},
		{"deck:CURRENT -card:2", []string{"forward", "subdeck"}},
		// Without subdecks (spec).
		{"deck:French -deck:French::*", []string{"forward", "reverse"}},
		{"deck:french::*", []string{"subdeck"}},
		// By card IDs (spec).
		{"cid:101,102", []string{"forward", "reverse"}},
		{"cid:122", []string{"cloze 2"}},
		{"nid:10", []string{"forward", "reverse"}},
		{`"nid:11, 13"`, []string{"subdeck", "spanish"}},
		{"-cid:101,102,111", []string{"cloze 2", "spanish"}},
		{"cid:999", nil},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, o, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q = %v, want %v", c.query, got, c.want)
		}
	}
	// No deck studied yet: deck:current matches nothing.
	if got := matchNames(t, "deck:current", Options{}, cards, order); got != nil {
		t.Errorf("deck:current without a current deck = %v", got)
	}
	errs := map[string]string{
		"card:":    "card:: needs a card type name or number, e.g. card:Reverse or card:2",
		"card:0":   "card:0: card type numbers start at 1, e.g. card:2",
		"cid:":     "cid:: needs an ID or a comma-separated list of IDs, e.g. cid:1712345678901,1712345678902",
		"nid:1,x":  "nid:1,x: needs an ID or a comma-separated list of IDs, e.g. nid:1712345678901,1712345678902",
		"nid:1,,2": "nid:1,,2: needs an ID or a comma-separated list of IDs, e.g. nid:1712345678901,1712345678902",
	}
	for query, want := range errs {
		if _, err := Parse(query); err == nil || usermsg.Plain(err.Error()) != want {
			t.Errorf("%q: err = %v, want %q", query, err, want)
		}
	}
}

// Search matches either form (spec): stored text is in NFC, and a query
// typed in decomposed form is normalised before matching. Error positions
// count the characters as typed.
func TestDecomposedQuery(t *testing.T) {
	const cafe, café = "café", "café"
	cards := map[string]*Card{
		"café": {Fields: split("un " + café + "\x1fcoffee"), Raw: split("un " + café + "\x1fcoffee"), FieldNames: []string{"front", "back"},
			Deck: café, Tags: []string{café + "::noir"}},
		"field": {Fields: split("x\x1fy"), FieldNames: []string{café, "back"}},
		"plain": {Fields: split("cafe\x1f"), FieldNames: []string{"front", "back"}},
	}
	order := []string{"café", "field", "plain"}
	cases := []struct {
		query string
		want  []string
	}{
		{cafe, []string{"café"}},
		{`"un ` + cafe + `"`, []string{"café"}},
		{"front:*" + cafe, []string{"café"}},
		{"w:" + cafe, []string{"café"}},
		{"re:" + cafe + "$", []string{"café"}},
		{"deck:" + cafe, []string{"café"}},
		{"tag:" + cafe, []string{"café"}},
		{cafe + ":x", []string{"field"}},
		{"-" + cafe, []string{"field", "plain"}},
	}
	for _, c := range cases {
		if got := matchNames(t, c.query, Options{}, cards, order); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q matched %v, want %v", c.query, got, c.want)
		}
	}
	_, err := Parse(cafe + " )")
	var e *Error
	if !errors.As(err, &e) || e.Pos != 7 {
		t.Fatalf("error = %v, want one at character 7", err)
	}
}
