// Package yield lets long loops give the browser a chance to run. Go wasm
// runs on the page's main thread and only returns to it when every
// goroutine is blocked, so a long import would freeze the UI; a short
// sleep every few hundred items keeps it responsive. On native platforms
// Go runs on its own threads and Yield does nothing.
package yield

// pause is Yield; tests replace it to count calls.
var pause = Yield

// Every returns a function that yields once every n calls.
func Every(n int) func() {
	i := 0
	return func() {
		if i++; i >= n {
			i = 0
			pause()
		}
	}
}
