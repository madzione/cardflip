//go:build !js

package storage

// RequestPersistence asks the platform to keep stored data under storage
// pressure. Files in the app data directory are never evicted.
func RequestPersistence() bool { return true }
