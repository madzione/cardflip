// Package csv reads delimited text files (CSV, TSV, semicolon-separated)
// for import: it detects the delimiter and a UTF-8 byte order mark, parses
// quoted and multi-line fields, and turns rows into an import batch using a
// column mapping.
package csv

import (
	"bytes"
	"cardflip/golang/internal/usermsg"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"cardflip/golang/core"
)

// File is a parsed delimited text file.
type File struct {
	Delimiter rune
	// Rows holds every record; the header row, when there is one, is Rows[0].
	Rows [][]string
	// Columns is the largest number of fields in a row.
	Columns int
	// Header is set when the file declares its column names (Anki's
	// "#columns:" line); those names are not part of Rows.
	Header bool
	// ColumnNames names the columns when Header is set.
	ColumnNames []string
	// HTML is set when the file declares its fields HTML ("#html:true", as
	// in Cardflip and Anki text exports): line breaks in fields are then
	// kept as they are.
	HTML bool
	// NoHTML is set when the file declares its fields plain text
	// ("#html:false"), so "Allow HTML in fields" starts off.
	NoHTML bool
	// Tags are the tags of a "#tags:" line, for every note of the file.
	Tags []string
	// Roles maps the columns (from 0) that "#notetype column:", "#deck
	// column:", "#tags column:" and "#guid column:" lines name to NoteType,
	// Deck, Tags or GUID.
	Roles map[int]int
	// NoteType and Deck are the names of "#notetype:" and "#deck:" lines,
	// the defaults for rows without a note type or deck of their own.
	NoteType, Deck string
	// Comments counts the lines starting with "#" after the header lines,
	// which are skipped.
	Comments int
}

var candidates = []rune{'\t', ';', ','}

// separators are the separators a "#separator:" line can name.
var separators = map[string]rune{"comma": ',', "semicolon": ';', "tab": '\t', "space": ' ', "pipe": '|', "colon": ':'}

// SeparatorName names a separator: "tab", "comma", "semicolon", "space",
// "pipe", "colon", or the character itself for any other.
func SeparatorName(r rune) string {
	for name, sep := range separators {
		if sep == r {
			return name
		}
	}
	return string(r)
}

// ParseSeparator reads a separator given by name (in any case) or as the
// character itself.
func ParseSeparator(s string) (rune, bool) {
	if r, ok := separators[strings.ToLower(strings.TrimSpace(s))]; ok {
		return r, true
	}
	// A lone space or tab is the separator itself.
	if t := strings.TrimSpace(s); t != "" {
		s = t
	}
	if utf8.RuneCountInString(s) == 1 {
		r, _ := utf8.DecodeRuneInString(s)
		for _, sep := range separators {
			if sep == r {
				return r, true
			}
		}
	}
	return 0, false
}

// Parse reads a delimited text file.
func Parse(data []byte) (*File, error) {
	return ParseWith(data, 0)
}

// ParseWith reads a delimited text file split at sep, or, when sep is 0,
// at the separator the file declares or, without one, the detected one.
func ParseWith(data []byte, sep rune) (*File, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	if !utf8.Valid(data) {
		return nil, errors.New(usermsg.Text("csvNotUtf8", "the file is not UTF-8 text; save it as UTF-8 and try again"))
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")

	// Anki text exports start with "#key:value" lines.
	f := &File{}
	var declared rune
	var columns string
	for strings.HasPrefix(text, "#") {
		line, rest, _ := strings.Cut(text, "\n")
		key, value, ok := strings.Cut(strings.TrimPrefix(line, "#"), ":")
		if !ok {
			break
		}
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "separator":
			declared, _ = ParseSeparator(value)
		case "html":
			f.HTML = strings.EqualFold(strings.TrimSpace(value), "true")
			f.NoHTML = strings.EqualFold(strings.TrimSpace(value), "false")
		case "tags":
			f.Tags = strings.Fields(value)
		case "columns":
			columns = value
			f.Header = true
		case "notetype":
			f.NoteType = strings.TrimSpace(value)
		case "deck":
			f.Deck = strings.TrimSpace(value)
		case "notetype column", "deck column", "tags column", "guid column":
			col, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || col < 1 {
				break
			}
			if f.Roles == nil {
				f.Roles = map[int]int{}
			}
			f.Roles[col-1] = roleNames[strings.Fields(strings.ToLower(key))[0]]
		}
		text = rest
	}
	if strings.TrimSpace(text) == "" {
		return nil, errors.New(usermsg.Text("csvNoRows", "the file has no rows"))
	}

	f.Delimiter = sep
	if f.Delimiter == 0 {
		f.Delimiter = declared
	}
	if f.Delimiter == 0 {
		f.Delimiter = detect(text)
	}
	if f.Header {
		// Column names are split at the declared separator, or at any of
		// the common ones when the file declares none.
		f.ColumnNames = strings.FieldsFunc(columns, func(r rune) bool {
			if declared != 0 {
				return r == declared
			}
			return r == '\t' || r == ',' || r == ';'
		})
	}
	rows, comments, err := read(text, f.Delimiter, -1)
	if err != nil {
		return nil, errors.New(usermsg.Text("fileUnreadable", "the file could not be read: %v", err))
	}
	for _, r := range rows {
		f.Columns = max(f.Columns, len(r))
	}
	for col := range f.Roles {
		if col >= f.Columns {
			delete(f.Roles, col)
		}
	}
	f.Rows = rows
	f.Comments = comments
	return f, nil
}

// read reads up to limit rows (all when negative), skipping blank lines and
// comment lines, and counts the comment lines.
func read(text string, delim rune, limit int) (rows [][]string, comments int, err error) {
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = delim
	r.Comment = '#'
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	r.ReuseRecord = false
	// The reader skips comment and blank lines before a record: count the
	// comment lines among what it read before each record and at the end.
	var offset int64
	countSkipped := func(chunk string) {
		for chunk != "" {
			var line string
			line, chunk, _ = strings.Cut(chunk, "\n")
			switch {
			case strings.HasPrefix(line, "#"):
				comments++
			case line != "":
				return // the record
			}
		}
	}
	for limit < 0 || len(rows) < limit {
		rec, err := r.Read()
		countSkipped(text[offset:r.InputOffset()])
		offset = r.InputOffset()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		if len(rec) == 1 && strings.TrimSpace(rec[0]) == "" {
			continue // blank line
		}
		rows = append(rows, rec)
	}
	return rows, comments, nil
}

// detect picks the delimiter that splits the first 20 rows most
// consistently into more than one column.
func detect(text string) rune {
	best, bestScore := ',', -1
	for _, d := range candidates {
		rows, _, err := read(text, d, 20)
		if err != nil || len(rows) == 0 {
			continue
		}
		counts := map[int]int{}
		for _, r := range rows {
			counts[len(r)]++
		}
		mode, modeRows := 0, 0
		for n, k := range counts {
			if k > modeRows || k == modeRows && n > mode {
				mode, modeRows = n, k
			}
		}
		if mode < 2 {
			continue
		}
		// Rows agreeing on the column count matter most, then more columns.
		score := modeRows*100 + mode
		if score > bestScore {
			best, bestScore = d, score
		}
	}
	return best
}

// Column mapping targets besides field indexes: a column holds each row's
// tags, note type name, deck name or GUID.
const (
	Ignore   = -1
	Tags     = -2
	NoteType = -3
	Deck     = -4
	GUID     = -5
)

var roleNames = map[string]int{"notetype": NoteType, "deck": Deck, "tags": Tags, "guid": GUID}

// DefaultMapping maps column i to field i, and extra columns to nothing.
func DefaultMapping(columns, fields int) []int {
	return (&File{Columns: columns}).Mapping(fields)
}

// Mapping is the file's default column mapping: the columns its header
// lines name get their role, and the other columns map to fields in order,
// up to the given number of fields; extra columns map to nothing.
func (f *File) Mapping(fields int) []int {
	m := make([]int, f.Columns)
	next := 0
	for i := range m {
		switch role, ok := f.Roles[i]; {
		case ok:
			m[i] = role
		case next < fields:
			m[i] = next
			next++
		default:
			m[i] = Ignore
		}
	}
	return m
}

// Options configure Build.
type Options struct {
	// TypeID is the note type of rows without one of their own, among
	// NoteTypes, the collection's note types.
	TypeID    core.ID
	NoteTypes []core.NoteType
	// DeckID is the deck of rows without one of their own, or DeckName
	// when it is set (created if missing).
	DeckID     core.ID
	DeckName   string
	Mapping    []int
	HasHeader  bool
	Duplicates core.DupMode
	// DupScope says where a row's first field is looked for.
	DupScope core.DupScope
	// AllowHTML imports field text as HTML. When it is off, "<", ">" and
	// "&" are kept as typed and line breaks become "<br>".
	AllowHTML bool
	// Tags are added to every note, e.g. the file's Tags.
	Tags []string
}

// plainText turns plain text into HTML that shows it as typed.
var plainText = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\n", "<br>")

// Build turns the file's rows into an import batch. Each row goes into its
// note type and deck, from the note type and deck columns or else the
// defaults; field columns map by position within the row's note type. Rows
// naming a note type that does not exist are skipped, with a warning.
func Build(f *File, opts Options) (*core.ImportBatch, error) {
	types := map[core.ID]*core.NoteType{}
	for i := range opts.NoteTypes {
		types[opts.NoteTypes[i].ID] = &opts.NoteTypes[i]
	}
	def := types[opts.TypeID]
	if def == nil || opts.DeckID == 0 && strings.TrimSpace(opts.DeckName) == "" {
		return nil, errors.New(usermsg.Text("csvChooseTypeAndDeck", "choose a note type and a deck"))
	}
	mapped := false
	roles := map[int]int{}
	for col, m := range opts.Mapping {
		switch {
		case m >= 0:
			mapped = true
		case m == NoteType || m == Deck || m == GUID:
			if _, ok := roles[m]; ok {
				return nil, errors.New(usermsg.Text(roleCode(m), "map only one column to the "+roleLabel(m)))
			}
			roles[m] = col
		}
	}
	if !mapped {
		return nil, errors.New(usermsg.Text("csvMapAField", "map at least one column to a field"))
	}
	_, perRowType := roles[NoteType]
	if !perRowType {
		for _, m := range opts.Mapping {
			if m >= len(def.Fields) {
				return nil, errors.New(usermsg.Text("csvFieldOutOfRange", "column mapped to field %d, but the note type has %d fields", m+1, len(def.Fields)))
			}
		}
	}
	b := &core.ImportBatch{Duplicates: opts.Duplicates, DupScope: opts.DupScope}
	if b.Duplicates == "" {
		b.Duplicates = core.DupSkip
	}
	keys := map[core.ID]string{}
	typeKey := func(nt *core.NoteType) string {
		if k, ok := keys[nt.ID]; ok {
			return k
		}
		k := fmt.Sprintf("csv:%d", nt.ID)
		keys[nt.ID] = k
		b.NoteTypes = append(b.NoteTypes, core.ImportNoteType{Key: k, ExistingID: nt.ID})
		return k
	}
	cell := func(row []string, role int) string {
		if col, ok := roles[role]; ok && col < len(row) {
			return strings.TrimSpace(row[col])
		}
		return ""
	}

	rows := f.Rows
	if opts.HasHeader && len(rows) > 0 && len(f.ColumnNames) == 0 {
		rows = rows[1:]
	}
	var unknown []string
	unknownRows := map[string]int{}
	extra := 0
	for _, row := range rows {
		nt := def
		if name := cell(row, NoteType); name != "" {
			if nt = NoteTypeNamed(opts.NoteTypes, name); nt == nil {
				if unknownRows[name] == 0 {
					unknown = append(unknown, name)
				}
				unknownRows[name]++
				continue
			}
		}
		n := core.ImportNote{TypeKey: typeKey(nt), DeckID: opts.DeckID, DeckName: strings.TrimSpace(opts.DeckName),
			GUID: cell(row, GUID), Fields: make([]string, len(nt.Fields))}
		if deck := cell(row, Deck); deck != "" {
			n.DeckName = deck
		}
		n.Tags = append(n.Tags, opts.Tags...)
		dropped := false
		for col, target := range opts.Mapping {
			if col >= len(row) {
				continue
			}
			v := row[col]
			switch {
			case target == Tags:
				n.Tags = append(n.Tags, v)
			case target >= len(n.Fields):
				// A field the row's note type does not have.
				dropped = dropped || strings.TrimSpace(v) != ""
			case target >= 0 && !opts.AllowHTML:
				n.Fields[target] = plainText.Replace(v)
			case target >= 0 && f.HTML:
				n.Fields[target] = v
			case target >= 0:
				// Field text is HTML; a line break inside a quoted field
				// becomes a visible one.
				n.Fields[target] = strings.ReplaceAll(v, "\n", "<br>")
			}
		}
		if dropped {
			extra++
		}
		b.Notes = append(b.Notes, n)
	}
	for _, name := range unknown {
		b.Skipped += unknownRows[name]
		b.Warnings = append(b.Warnings, usermsg.Tag("csvSkippedUnknownType", usermsg.Args(unknownRows[name], name), fmt.Sprintf("Skipped %s: there is no note type named %q.", rowCount(unknownRows[name]), name)))
	}
	if extra > 0 {
		b.Warnings = append(b.Warnings, usermsg.Tag("csvExtraFields", usermsg.Args(extra), fmt.Sprintf("%s had more fields than their note type; the extra fields were not imported.", rowCount(extra))))
	}
	return b, nil
}

// NoteTypeNamed finds a note type by name: the one with that exact name,
// else the oldest with that name in another case.
func NoteTypeNamed(types []core.NoteType, name string) *core.NoteType {
	var folded *core.NoteType
	for i := range types {
		nt := &types[i]
		if nt.Name == name {
			return nt
		}
		if strings.EqualFold(nt.Name, name) && (folded == nil || nt.ID < folded.ID) {
			folded = nt
		}
	}
	return folded
}

// roleCode is the code of the "map only one column" error for role.
func roleCode(role int) string {
	switch role {
	case NoteType:
		return "csvOneNoteTypeColumn"
	case Deck:
		return "csvOneDeckColumn"
	case GUID:
		return "csvOneGuidColumn"
	}
	return "csvOneTagsColumn"
}

func roleLabel(role int) string {
	switch role {
	case NoteType:
		return "note type"
	case Deck:
		return "deck"
	case GUID:
		return "GUID"
	}
	return "tags"
}

func rowCount(n int) string {
	if n == 1 {
		return "1 row"
	}
	return strconv.Itoa(n) + " rows"
}

// LooksLikeHeader reports whether a row names the given fields, e.g.
// "Front,Back" for Basic, so it is probably a header and not a note.
func LooksLikeHeader(row []string, fields []string) bool {
	if len(row) == 0 {
		return false
	}
	matches := 0
	for _, cell := range row {
		for _, f := range fields {
			if strings.EqualFold(strings.TrimSpace(cell), f) {
				matches++
				break
			}
		}
	}
	return matches > 0 && matches >= len(row)/2+len(row)%2
}
