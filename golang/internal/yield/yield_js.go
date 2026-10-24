//go:build js

package yield

import "time"

// Yield pauses briefly so the browser can render and handle input.
func Yield() { time.Sleep(time.Millisecond) }
