//go:build !js

package yield

// Yield does nothing on native platforms.
func Yield() {}
