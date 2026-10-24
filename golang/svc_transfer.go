package main

import (
	"archive/zip"
	"bytes"
	"cardflip/golang/internal/usermsg"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"cardflip/golang/backup"
	"cardflip/golang/core"
	exporter "cardflip/golang/exporter/apkg"
	textexport "cardflip/golang/exporter/text"
	"cardflip/golang/importer/apkg"
	"cardflip/golang/importer/csv"
	"cardflip/golang/internal/yield"
	"cardflip/golang/render"
)

// TransferService imports files and exports and restores backups. Files
// are staged first: by path on native platforms (Go reads the file), as
// bytes on the web.
type TransferService struct {
	a      *app
	mu     sync.Mutex
	next   int64
	stages map[int64]*stage
}

type stage struct {
	info StageInfo
	data []byte
}

// StageInfo describes a staged file.
type StageInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Kind is "csv", "apkg", "colpkg" (an Anki collection package) or
	// "backup".
	Kind string `json:"kind"`
	Size int    `json:"size"`
}

// reset drops the staged files; another profile was opened.
func (s *TransferService) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stages = nil
}

// StagePath stages a file by path (native platforms).
func (s *TransferService) StagePath(filePath string) (StageInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return StageInfo{}, errors.New(usermsg.Text("fileUnreadable", "the file could not be read: %v", err))
	}
	return s.StageBytes(filepath.Base(filePath), data)
}

// StageBytes stages a file's content (web).
func (s *TransferService) StageBytes(name string, data []byte) (StageInfo, error) {
	kind, err := kindOfFile(name, data)
	if err != nil {
		return StageInfo{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stages == nil {
		s.stages = map[int64]*stage{}
	}
	s.next++
	info := StageInfo{ID: s.next, Name: name, Kind: kind, Size: len(data)}
	s.stages[info.ID] = &stage{info: info, data: data}
	return info, nil
}

func kindOfFile(name string, data []byte) (string, error) {
	switch strings.ToLower(path.Ext(name)) {
	case ".apkg":
		return "apkg", nil
	case ".colpkg":
		return "colpkg", nil
	case ".cardflip":
		return "backup", nil
	case ".csv", ".tsv", ".txt":
		return "csv", nil
	}
	if zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data))); err == nil {
		for _, f := range zr.File {
			switch f.Name {
			case "manifest.json":
				return "backup", nil
			case "collection.anki21b", "collection.anki21", "collection.anki2":
				return "apkg", nil
			}
		}
		return "", core.NewError("notPackageOrBackup", "this file is neither an Anki package nor a Cardflip backup")
	}
	return "csv", nil
}

func (s *TransferService) stage(id int64) (*stage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.stages[id]
	if !ok {
		return nil, core.NewError("fileGone", "the file is no longer available; please choose it again")
	}
	return st, nil
}

// Discard forgets a staged file.
func (s *TransferService) Discard(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.stages, id)
}

// CsvPreview shows the first rows of a delimited file and the default
// import settings.
type CsvPreview struct {
	// Delimiter is "comma", "semicolon", "tab", "space", "pipe", "colon" or
	// the character itself.
	Delimiter string     `json:"delimiter"`
	Columns   int        `json:"columns"`
	Rows      [][]string `json:"rows"`
	TotalRows int        `json:"totalRows"`
	// HeaderRow is set when the first row looks like column names.
	HeaderRow   bool     `json:"headerRow"`
	ColumnNames []string `json:"columnNames"`
	// TypeID and DeckID are the defaults for rows without a note type or
	// deck of their own: the file's "#notetype:" and "#deck:" ones when they
	// exist. DeckName is the file's deck when it does not exist yet; it is
	// created by an import that uses it (ImportOptions.DeckName).
	TypeID   int64  `json:"typeId"`
	DeckID   int64  `json:"deckId"`
	DeckName string `json:"deckName"`
	// Mapping maps each column to a field index, -1 (ignore), -2 (tags),
	// -3 (note type), -4 (deck) or -5 (GUID).
	Mapping []int `json:"mapping"`
	// NoteTypes describes the note type column, when Mapping has one.
	NoteTypes CsvNoteTypes `json:"noteTypes"`
	// Warnings are problems with the file's header lines.
	Warnings []string `json:"warnings"`
	// AllowHTML is the default of "Allow HTML in fields": on, unless the
	// file declares "#html:false".
	AllowHTML bool `json:"allowHtml"`
	// Tags are the tags of the file's "#tags:" line, added to every note.
	Tags []string `json:"tags"`
}

// ApkgInfo summarises an Anki package before importing it.
type ApkgInfo struct {
	Format string `json:"format"`
	Notes  int    `json:"notes"`
	Cards  int    `json:"cards"`
	// Decks are the decks with cards; for a collection package, every deck
	// apart from filtered ones.
	Decks []string `json:"decks"`
	// Media counts the images, audio and video files that are imported.
	Media int `json:"media"`
	// Unsupported counts audio and video files in formats the app cannot
	// play, which are skipped.
	Unsupported int `json:"unsupported"`
}

// BackupInfo describes a backup before restoring it.
type BackupInfo struct {
	Format     int       `json:"format"`
	AppVersion string    `json:"appVersion"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Inspection is what Inspect found in a staged file.
type Inspection struct {
	Stage  StageInfo   `json:"stage"`
	Csv    *CsvPreview `json:"csv"`
	Apkg   *ApkgInfo   `json:"apkg"`
	Backup *BackupInfo `json:"backup"`
}

// Inspect reads a staged file and describes it.
func (s *TransferService) Inspect(id int64) (Inspection, error) {
	st, err := s.stage(id)
	if err != nil {
		return Inspection{}, err
	}
	out := Inspection{Stage: st.info}
	switch st.info.Kind {
	case "backup":
		a, err := backup.Read(st.data)
		if err != nil {
			return out, err
		}
		out.Backup = &BackupInfo{Format: a.Manifest.Format, AppVersion: a.Manifest.AppVersion, CreatedAt: a.Manifest.CreatedAt}
	case "apkg", "colpkg":
		p, err := apkg.Read(st.data, yield.Yield)
		if err != nil {
			return out, err
		}
		info := &ApkgInfo{Format: p.Format, Notes: len(p.Notes), Cards: len(p.Cards)}
		used := map[int64]bool{}
		for _, c := range p.Cards {
			used[c.DeckID] = true
			if c.ODeckID != 0 {
				used[c.ODeckID] = true
			}
		}
		for id, d := range p.Decks {
			if (used[id] || st.info.Kind == "colpkg") && !d.Filtered {
				info.Decks = append(info.Decks, d.Name)
			}
		}
		sort.Strings(info.Decks)
		b, err := apkg.Convert(p, false)
		if err != nil {
			return out, err
		}
		info.Media = len(b.Media)
		info.Unsupported = len(p.Media) - len(b.Media)
		out.Apkg = info
	default:
		p, err := s.csvPreview(st, "")
		if err != nil {
			return out, err
		}
		out.Csv = &p
	}
	return out, nil
}

// PreviewCsv reads a staged delimited file again, split at another
// separator: "comma", "semicolon", "tab", "space", "pipe", "colon", or ""
// for the file's own or the detected one.
func (s *TransferService) PreviewCsv(id int64, separator string) (CsvPreview, error) {
	st, err := s.stage(id)
	if err != nil {
		return CsvPreview{}, err
	}
	if st.info.Kind != "csv" {
		return CsvPreview{}, core.NewError("notTextFile", "this file is not a text file")
	}
	return s.csvPreview(st, separator)
}

func (s *TransferService) csvPreview(st *stage, separator string) (CsvPreview, error) {
	c, err := s.a.collection()
	if err != nil {
		return CsvPreview{}, err
	}
	f, err := parseCsv(st.data, separator)
	if err != nil {
		return CsvPreview{}, err
	}
	defaults, err := (&NotesService{a: s.a}).Defaults(0)
	if err != nil {
		return CsvPreview{}, err
	}
	p := CsvPreview{
		Delimiter: csv.SeparatorName(f.Delimiter), Columns: f.Columns, TotalRows: len(f.Rows),
		ColumnNames: f.ColumnNames, TypeID: defaults.TypeID, DeckID: defaults.DeckID,
		AllowHTML: !f.NoHTML, Tags: f.Tags, Warnings: []string{},
	}
	types := c.NoteTypes()
	if id, ok := c.BasicNoteType(); ok {
		p.TypeID = int64(id)
	}
	if f.NoteType != "" {
		if nt := csv.NoteTypeNamed(types, f.NoteType); nt != nil {
			p.TypeID = int64(nt.ID)
		} else {
			p.Warnings = append(p.Warnings, usermsg.Text("csvNoteTypeMissing", "The file's note type %q does not exist; rows without a note type of their own use the one chosen here.", f.NoteType))
		}
	}
	if f.Deck != "" {
		if id, ok := deckNamed(c, f.Deck); ok {
			p.DeckID = int64(id)
		} else {
			p.DeckName = f.Deck
		}
	}
	var fields []string
	if nt, err := c.NoteType(core.ID(p.TypeID)); err == nil {
		fields = nt.Fields
	}
	p.HeaderRow = len(f.Rows) > 0 && len(f.ColumnNames) == 0 && f.Roles == nil && csv.LooksLikeHeader(f.Rows[0], fields)
	p.Mapping = f.Mapping(len(fields))
	if slices.Contains(p.Mapping, csv.NoteType) {
		// Field columns map by position within each row's note type: map
		// as many as the largest of them has.
		p.NoteTypes = csvNoteTypes(types, f, p.Mapping, p.HeaderRow, core.ID(p.TypeID))
		p.Mapping = f.Mapping(len(p.NoteTypes.Fields))
	}
	if p.NoteTypes.Unknown == nil {
		p.NoteTypes = CsvNoteTypes{Unknown: []string{}, Fields: fields}
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}
	p.Rows = f.Rows
	if len(p.Rows) > 10 {
		p.Rows = p.Rows[:10]
	}
	return p, nil
}

// CsvNoteTypes describes the note types of a file's rows.
type CsvNoteTypes struct {
	// Unknown names the note types of the note type column that do not
	// exist, in the order of the rows; their rows are skipped.
	Unknown []string `json:"unknown"`
	// UnknownRows counts the rows naming them.
	UnknownRows int `json:"unknownRows"`
	// Fields names the field columns: the fields of the rows' note type when
	// they all have the same, else "Field 1", "Field 2", … as many as the
	// note type with the most fields has.
	Fields []string `json:"fields"`
}

// CsvNoteTypes describes the note types of a staged file's rows, read with
// the separator, mapping, header row and default note type of options.
func (s *TransferService) CsvNoteTypes(id int64, options ImportOptions) (CsvNoteTypes, error) {
	st, err := s.stage(id)
	if err != nil {
		return CsvNoteTypes{}, err
	}
	if st.info.Kind != "csv" {
		return CsvNoteTypes{}, core.NewError("notTextFile", "this file is not a text file")
	}
	c, err := s.a.collection()
	if err != nil {
		return CsvNoteTypes{}, err
	}
	f, err := parseCsv(st.data, options.Separator)
	if err != nil {
		return CsvNoteTypes{}, err
	}
	return csvNoteTypes(c.NoteTypes(), f, options.Mapping, options.HasHeader, core.ID(options.TypeID)), nil
}

func csvNoteTypes(types []core.NoteType, f *csv.File, mapping []int, hasHeader bool, typeID core.ID) CsvNoteTypes {
	out := CsvNoteTypes{Unknown: []string{}, Fields: []string{}}
	col := slices.Index(mapping, csv.NoteType)
	rows := f.Rows
	if hasHeader && len(rows) > 0 && len(f.ColumnNames) == 0 {
		rows = rows[1:]
	}
	var used []*core.NoteType
	seen := map[string]bool{}
	use := func(nt *core.NoteType) {
		if !slices.ContainsFunc(used, func(u *core.NoteType) bool { return u.ID == nt.ID }) {
			used = append(used, nt)
		}
	}
	for _, row := range rows {
		name := ""
		if col >= 0 && col < len(row) {
			name = strings.TrimSpace(row[col])
		}
		if name == "" {
			if i := slices.IndexFunc(types, func(nt core.NoteType) bool { return nt.ID == typeID }); i >= 0 {
				use(&types[i])
			}
			continue
		}
		if nt := csv.NoteTypeNamed(types, name); nt != nil {
			use(nt)
			continue
		}
		out.UnknownRows++
		if !seen[name] {
			seen[name] = true
			out.Unknown = append(out.Unknown, name)
		}
	}
	if len(used) == 1 {
		out.Fields = append(out.Fields, used[0].Fields...)
		return out
	}
	n := 0
	for _, nt := range used {
		n = max(n, len(nt.Fields))
	}
	for i := range n {
		out.Fields = append(out.Fields, fmt.Sprintf("Field %d", i+1))
	}
	return out
}

// deckNamed finds a deck by its full name, ignoring case and extra spaces.
func deckNamed(c *core.Collection, name string) (core.ID, bool) {
	name = strings.Join(strings.Fields(name), " ")
	for _, d := range c.Decks() {
		if strings.EqualFold(d.Name, name) {
			return d.ID, true
		}
	}
	return 0, false
}

// parseCsv parses a delimited file, split at the named separator or, when
// it is empty, at the file's own or the detected one.
func parseCsv(data []byte, separator string) (*csv.File, error) {
	if separator == "" {
		return csv.Parse(data)
	}
	sep, ok := csv.ParseSeparator(separator)
	if !ok {
		return nil, core.NewError("unknownSeparator", "unknown separator %q", separator)
	}
	return csv.ParseWith(data, sep)
}

// ImportOptions configure an import.
type ImportOptions struct {
	// CSV: the note type and deck of rows without their own, or the deck
	// named DeckName (created if missing) instead of DeckID; the column
	// mapping (see CsvPreview.Mapping); whether the first row is a header;
	// and what to do with a row whose first field matches a note:
	// "skip" (default), "update" or "duplicate", within its note type
	// ("noteType", the default) or its note type and deck ("deck"). A row
	// whose GUID matches a note of its note type updates it.
	TypeID         int64  `json:"typeId"`
	DeckID         int64  `json:"deckId"`
	DeckName       string `json:"deckName"`
	Mapping        []int  `json:"mapping"`
	HasHeader      bool   `json:"hasHeader"`
	Duplicates     string `json:"duplicates"`
	DuplicateScope string `json:"duplicateScope"`
	// CSV: the separator, as for PreviewCsv; whether field text is HTML
	// (see CsvPreview.AllowHTML), otherwise it is imported as typed; and
	// tags added to every note (see CsvPreview.Tags).
	Separator string   `json:"separator"`
	AllowHTML bool     `json:"allowHtml"`
	Tags      []string `json:"tags"`
	// Anki packages: keep Anki's scheduling, or import every card as New.
	KeepScheduling bool `json:"keepScheduling"`
	// Anki packages: when to update notes and note types that exist
	// already, "ifNewer" (default, also when empty), "always" or "never";
	// and whether to merge a note type whose fields or card types changed
	// into the existing one instead of adding it as a separate note type.
	UpdateNotes     string `json:"updateNotes"`
	UpdateNoteTypes string `json:"updateNoteTypes"`
	MergeNoteTypes  bool   `json:"mergeNoteTypes"`
}

// ImportSummary reports what an import did.
type ImportSummary struct {
	NotesAdded   int `json:"notesAdded"`
	NotesUpdated int `json:"notesUpdated"`
	NotesSkipped int `json:"notesSkipped"`
	// NotesSkippedOlder counts the notes not updated because the package's
	// version was not modified after the collection's ("if newer").
	NotesSkippedOlder int `json:"notesSkippedOlder"`
	CardsAdded        int `json:"cardsAdded"`
	// NoteTypesAdded, NoteTypesUpdated and NoteTypesMerged count the note
	// types created, existing ones given new templates or styling, and
	// existing ones that gained fields or card types.
	NoteTypesAdded   int `json:"noteTypesAdded"`
	NoteTypesUpdated int `json:"noteTypesUpdated"`
	NoteTypesMerged  int `json:"noteTypesMerged"`
	// Reviews counts the review history entries added.
	Reviews int `json:"reviews"`
	Media   int `json:"media"`
	// CommentLines counts the comment lines skipped in a text file.
	CommentLines int      `json:"commentLines"`
	Warnings     []string `json:"warnings"`
}

// ImportEvent is progress, then a final summary or error.
type ImportEvent struct {
	// Stage is "reading", "notes", "history", "media" or "saving".
	Stage   string         `json:"stage"`
	Done    int            `json:"done"`
	Total   int            `json:"total"`
	Summary *ImportSummary `json:"summary"`
	// Error is set when the import failed; the collection is unchanged.
	Error string `json:"error"`
}

// Import imports a staged CSV/TSV file or Anki package, streaming progress
// and ending with a summary or an error. An Anki collection package
// replaces the whole collection (data-import spec), after a backup of the
// current one; the app settings are kept.
func (s *TransferService) Import(ctx context.Context, id int64, options ImportOptions) <-chan ImportEvent {
	out := make(chan ImportEvent, 16)
	go func() {
		defer close(out)
		send := func(e ImportEvent) {
			select {
			case out <- e:
			case <-ctx.Done():
			}
		}
		sum, err := s.runImport(id, options, func(stage string, done, total int) {
			// Progress is best effort: never block the import on a slow UI.
			select {
			case out <- ImportEvent{Stage: stage, Done: done, Total: total}:
			default:
			}
		})
		if err != nil {
			send(ImportEvent{Error: err.Error()})
			return
		}
		s.Discard(id)
		send(ImportEvent{Stage: "done", Done: 1, Total: 1, Summary: &sum})
	}()
	return out
}

func (s *TransferService) runImport(id int64, o ImportOptions, progress core.Progress) (ImportSummary, error) {
	st, err := s.stage(id)
	if err != nil {
		return ImportSummary{}, err
	}
	if st.info.Kind == "colpkg" {
		return s.replaceCollection(st, progress)
	}
	c, err := s.a.collection()
	if err != nil {
		return ImportSummary{}, err
	}
	var sum core.ImportSummary
	switch st.info.Kind {
	case "apkg":
		sum, err = apkg.Import(c, st.data, apkg.Options{KeepScheduling: o.KeepScheduling,
			UpdateNotes: core.UpdatePolicy(o.UpdateNotes), UpdateNoteTypes: core.UpdatePolicy(o.UpdateNoteTypes),
			MergeNoteTypes: o.MergeNoteTypes}, progress, yield.Yield)
	case "csv":
		progress("reading", 0, 1)
		f, perr := parseCsv(st.data, o.Separator)
		if perr != nil {
			return ImportSummary{}, perr
		}
		if _, nerr := c.NoteType(core.ID(o.TypeID)); nerr != nil {
			return ImportSummary{}, nerr
		}
		b, berr := csv.Build(f, csv.Options{TypeID: core.ID(o.TypeID), NoteTypes: c.NoteTypes(),
			DeckID: core.ID(o.DeckID), DeckName: o.DeckName,
			Mapping: o.Mapping, HasHeader: o.HasHeader, Duplicates: core.DupMode(o.Duplicates),
			DupScope: core.DupScope(o.DuplicateScope), AllowHTML: o.AllowHTML, Tags: o.Tags})
		if berr != nil {
			return ImportSummary{}, berr
		}
		sum, err = c.Import(b, progress)
		sum.CommentLines = f.Comments
	default:
		return ImportSummary{}, core.NewError("backupNotImportable", "a backup cannot be imported; restore it from Settings instead")
	}
	if err != nil {
		return ImportSummary{}, err
	}
	return importSummary(sum), nil
}

// replaceCollection replaces the collection with an Anki collection
// package. The package is read before anything changes, so a damaged one
// leaves the collection, and the backups, as they are.
func (s *TransferService) replaceCollection(st *stage, progress core.Progress) (ImportSummary, error) {
	progress("reading", 0, 1)
	b, err := apkg.ReadCollection(st.data, yield.Yield)
	if err != nil {
		return ImportSummary{}, err
	}
	var sum core.ImportSummary
	err = s.a.restoreWith(func(c *core.Collection) error {
		sum, err = c.ReplaceWithImport(b, progress)
		return err
	})
	if err != nil {
		return ImportSummary{}, err
	}
	return importSummary(sum), nil
}

func importSummary(sum core.ImportSummary) ImportSummary {
	out := ImportSummary(sum)
	if out.Warnings == nil {
		out.Warnings = []string{}
	}
	return out
}

// BackupFile is an exported backup.
type BackupFile struct {
	// Name is the suggested file name, cardflip-<YYYY-MM-DD>.cardflip.
	Name string `json:"name"`
	Data []byte `json:"data"`
}

// ExportBackup writes the whole collection to a .cardflip backup.
func (s *TransferService) ExportBackup() (BackupFile, error) {
	c, err := s.a.collection()
	if err != nil {
		return BackupFile{}, err
	}
	now := s.a.now()
	data, err := backup.Export(c, appVersion, now, yield.Yield)
	if err != nil {
		return BackupFile{}, err
	}
	return BackupFile{Name: backup.FileName(now), Data: data}, nil
}

// Restore replaces the collection with a staged backup. The current
// collection stays unchanged if the backup is invalid. It also works from
// the error screen, when the stored collection could not be read.
func (s *TransferService) Restore(id int64) error {
	st, err := s.stage(id)
	if err != nil {
		return err
	}
	a, err := backup.Read(st.data)
	if err != nil {
		return err
	}
	err = s.a.restoreWith(func(c *core.Collection) error {
		return c.Restore(a.Collection, a.Revlog, a.Deletions, a.Media)
	})
	if err != nil {
		return err
	}
	s.Discard(id)
	return nil
}

// ExportOptions configure an Anki package export (data-export spec).
type ExportOptions struct {
	// Scheduling keeps scheduling and review history; without it every
	// card is exported as new. Collection exports always keep it.
	Scheduling bool `json:"scheduling"`
	// Presets exports the deck presets. Collection exports always do.
	Presets bool `json:"presets"`
	Media   bool `json:"media"`
	// Legacy writes the format older Anki versions read.
	Legacy bool `json:"legacy"`
}

// ExportFile is an exported package.
type ExportFile struct {
	// Name is the suggested file name, e.g. Spanish.apkg.
	Name string `json:"name"`
	Data []byte `json:"data"`
}

// ExportEvent is progress, then the exported file or an error.
type ExportEvent struct {
	// Stage is "collection", "media", "saving", "cards" (rendering cards
	// for a text export) or, with File, "done".
	Stage string      `json:"stage"`
	Done  int         `json:"done"`
	Total int         `json:"total"`
	File  *ExportFile `json:"file"`
	// Error is set when the export failed.
	Error string `json:"error"`
}

// ExportDeck exports a deck and its subdecks as an .apkg named after the
// deck, streaming progress.
func (s *TransferService) ExportDeck(ctx context.Context, deckID int64, options ExportOptions) <-chan ExportEvent {
	return s.export(ctx, core.ExportScope{DeckID: core.ID(deckID)}, options, func(c *core.Collection) string {
		d, _ := c.Deck(core.ID(deckID))
		return exporter.FileName(d.Name, "apkg")
	})
}

// ExportNotes exports the notes of the cards, with all their cards, as an
// .apkg, streaming progress.
func (s *TransferService) ExportNotes(ctx context.Context, cardIDs []int64, options ExportOptions) <-chan ExportEvent {
	c, err := s.a.collection()
	if err != nil {
		return exportFailed(err)
	}
	scope := core.ExportScope{NoteIDs: notesOf(c, cardIDs)}
	return s.export(ctx, scope, options, func(*core.Collection) string { return "Selected notes.apkg" })
}

// ExportCollection exports the whole collection, with scheduling, review
// history and presets, as a .colpkg, streaming progress.
func (s *TransferService) ExportCollection(ctx context.Context, options ExportOptions) <-chan ExportEvent {
	options.Scheduling, options.Presets = true, true
	return s.export(ctx, core.ExportScope{All: true}, options, func(*core.Collection) string {
		return "collection-" + s.a.now().Format("2006-01-02") + ".colpkg"
	})
}

func exportFailed(err error) <-chan ExportEvent {
	out := make(chan ExportEvent, 1)
	out <- ExportEvent{Error: err.Error()}
	close(out)
	return out
}

func (s *TransferService) export(ctx context.Context, scope core.ExportScope, o ExportOptions, name func(*core.Collection) string) <-chan ExportEvent {
	return s.stream(ctx, func(progress exporter.Progress) (ExportFile, error) {
		return s.runExport(scope, o, name, progress)
	})
}

// stream runs an export in the background, streaming its progress and
// then the file or the error.
func (s *TransferService) stream(ctx context.Context, run func(exporter.Progress) (ExportFile, error)) <-chan ExportEvent {
	out := make(chan ExportEvent, 16)
	go func() {
		defer close(out)
		send := func(e ExportEvent) {
			select {
			case out <- e:
			case <-ctx.Done():
			}
		}
		file, err := run(func(stage string, done, total int) {
			// Progress is best effort: never block the export on a slow UI.
			select {
			case out <- ExportEvent{Stage: stage, Done: done, Total: total}:
			default:
			}
		})
		if err != nil {
			send(ExportEvent{Error: err.Error()})
			return
		}
		send(ExportEvent{Stage: "done", Done: 1, Total: 1, File: &file})
	}()
	return out
}

func (s *TransferService) runExport(scope core.ExportScope, o ExportOptions, name func(*core.Collection) string, progress exporter.Progress) (ExportFile, error) {
	c, err := s.a.collection()
	if err != nil {
		return ExportFile{}, err
	}
	set, err := c.Export(scope)
	if err != nil {
		return ExportFile{}, err
	}
	var reviews []core.ReviewEntry
	if o.Scheduling {
		if reviews, err = c.ExportReviews(set.Cards); err != nil {
			return ExportFile{}, err
		}
	}
	data, err := exporter.Write(exporter.Input{Set: set, Reviews: reviews, Media: c.Media, Now: s.a.now()},
		exporter.Options{Scheduling: o.Scheduling, Presets: o.Presets, Media: o.Media, Legacy: o.Legacy},
		progress, yield.Yield)
	if err != nil {
		return ExportFile{}, err
	}
	return ExportFile{Name: name(c), Data: data}, nil
}

// TextScope selects what a text export contains: a deck and its subdecks,
// the selected cards, or the whole collection.
type TextScope struct {
	DeckID  int64   `json:"deckId"`
	CardIDs []int64 `json:"cardIds"`
	All     bool    `json:"all"`
}

// TextOptions choose the columns a notes text export adds to the fields.
type TextOptions struct {
	Tags     bool `json:"tags"`
	Deck     bool `json:"deck"`
	NoteType bool `json:"noteType"`
	GUID     bool `json:"guid"`
}

// ExportNotesText exports notes as tab-separated text, one note per line
// with its fields as stored (data-export spec): the notes of a deck and its
// subdecks, of the selected cards, or of the whole collection. The file
// imports again into Cardflip or Anki.
func (s *TransferService) ExportNotesText(ctx context.Context, scope TextScope, options TextOptions) <-chan ExportEvent {
	return s.stream(ctx, func(progress exporter.Progress) (ExportFile, error) {
		c, err := s.a.collection()
		if err != nil {
			return ExportFile{}, err
		}
		progress("collection", 0, 1)
		es, name := s.textScope(c, scope, "notes")
		if len(scope.CardIDs) > 0 {
			es = core.ExportScope{NoteIDs: notesOf(c, scope.CardIDs)}
		}
		set, err := c.Export(es)
		if err != nil {
			return ExportFile{}, err
		}
		progress("saving", 0, 1)
		o := textexport.NoteOptions{Tags: options.Tags, Deck: options.Deck, NoteType: options.NoteType, GUID: options.GUID}
		return ExportFile{Name: name, Data: textexport.WriteNotes(set, o, yield.Yield)}, nil
	})
}

// ExportCardsText exports cards as tab-separated text, one card per line
// with its question and answer as plain text (data-export spec): the cards
// of a deck and its subdecks, the selected cards, or every card.
func (s *TransferService) ExportCardsText(ctx context.Context, scope TextScope) <-chan ExportEvent {
	return s.stream(ctx, func(progress exporter.Progress) (ExportFile, error) {
		c, err := s.a.collection()
		if err != nil {
			return ExportFile{}, err
		}
		if !scope.All && scope.DeckID == 0 && len(scope.CardIDs) == 0 {
			return ExportFile{}, core.NewError("noCardsSelected", "No cards are selected.")
		}
		progress("collection", 0, 1)
		es, name := s.textScope(c, scope, "cards")
		set, err := c.Export(es)
		if err != nil {
			return ExportFile{}, err
		}
		cards := make([]render.Card, 0, len(set.Cards))
		for i, card := range set.Cards {
			if i%100 == 0 {
				progress("cards", i, len(set.Cards))
				yield.Yield()
			}
			r, err := c.RenderCard(card.ID, render.Env{})
			if err != nil {
				continue // deleted since
			}
			cards = append(cards, r)
		}
		progress("saving", 0, 1)
		return ExportFile{Name: name, Data: textexport.WriteCards(cards, yield.Yield)}, nil
	})
}

// textScope returns the export scope of a text export and its file name:
// the deck's name, "Selected <what>.txt" or "<what>-<date>.txt".
func (s *TransferService) textScope(c *core.Collection, scope TextScope, what string) (core.ExportScope, string) {
	switch {
	case scope.All:
		return core.ExportScope{All: true}, what + "-" + s.a.now().Format("2006-01-02") + ".txt"
	case scope.DeckID != 0:
		name := "deck.txt"
		if d, err := c.Deck(core.ID(scope.DeckID)); err == nil {
			name = exporter.FileName(d.Name, "txt")
		}
		return core.ExportScope{DeckID: core.ID(scope.DeckID)}, name
	}
	ids := make([]core.ID, len(scope.CardIDs))
	for i, id := range scope.CardIDs {
		ids[i] = core.ID(id)
	}
	return core.ExportScope{CardIDs: ids}, "Selected " + what + ".txt"
}

// MediaService serves card images and stores media added in the editor.
type MediaService struct{ a *app }

// Get returns a media file's bytes.
func (s *MediaService) Get(name string) ([]byte, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	return c.Media(name)
}

// Add stores a media file added in the editor and returns the name to
// refer to it by. Same content as a stored file reuses that file; a taken
// name gets a numbered suffix; a name without a stem (".png") is named
// after the content. Files no note uses are removed by a later compaction.
func (s *MediaService) Add(name string, data []byte) (string, error) {
	c, err := s.a.collection()
	if err != nil {
		return "", err
	}
	return c.AddMedia(name, data)
}

// AddPastedImage stores an image pasted from the clipboard (app-shell spec:
// paste images as PNG): as JPEG, or as PNG when asPng is set. An image
// already in that format is stored unchanged; transparent parts become
// white in a JPEG. The file is named after its content.
func (s *MediaService) AddPastedImage(data []byte, asPng bool) (string, error) {
	c, err := s.a.collection()
	if err != nil {
		return "", err
	}
	data, ext, err := convertPastedImage(data, asPng)
	if err != nil {
		return "", err
	}
	return c.AddMedia(ext, data)
}

// jpegQuality is the quality of pasted images stored as JPEG.
const jpegQuality = 90

func convertPastedImage(data []byte, asPng bool) ([]byte, string, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", core.NewError("clipboardImageFormat", "the clipboard image is not a PNG, JPEG or GIF image")
	}
	var out bytes.Buffer
	if asPng {
		if format == "png" {
			return data, ".png", nil
		}
		if err := png.Encode(&out, img); err != nil {
			return nil, "", err
		}
		return out.Bytes(), ".png", nil
	}
	if format == "jpeg" {
		return data, ".jpg", nil
	}
	flat := image.NewRGBA(img.Bounds())
	draw.Draw(flat, flat.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(flat, flat.Bounds(), img, img.Bounds().Min, draw.Over)
	if err := jpeg.Encode(&out, flat, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return nil, "", err
	}
	return out.Bytes(), ".jpg", nil
}
