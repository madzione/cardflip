package main

import "cardflip/golang/fluffy"

//go:generate fluffy generate

func init() {
	fluffy.Start(func() error {
		a := newApp()
		fluffy.Bind("app", &AppService{a: a})
		fluffy.Bind("decks", &DecksService{a: a})
		fluffy.Bind("notes", &NotesService{a: a})
		study, transfer := &StudyService{a: a}, &TransferService{a: a}
		a.resets = []func(){study.reset, transfer.reset}
		fluffy.Bind("study", study)
		fluffy.Bind("browse", &BrowseService{a: a})
		fluffy.Bind("transfer", transfer)
		fluffy.Bind("media", &MediaService{a: a})
		fluffy.Bind("stats", &StatsService{a: a})
		fluffy.Bind("mcp", newMcpService(a))
		fluffy.Bind("sync", newSyncService(a, newSyncUI()))
		return nil
	})
}

// main only runs on the web, native platforms load golang/ as a library.
// Setup belongs in the fluffy.Start function above.
func main() {
	fluffy.Serve()
}
