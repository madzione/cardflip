package storage

import "fmt"

// BackupStore holds backup files next to the collection. Backups are not
// keys of the Backend, so replacing or resetting the collection never
// touches them.
type BackupStore interface {
	// List returns the stored backups, sorted by name.
	List() ([]BackupFile, error)
	// Write stores a backup atomically, replacing one with the same name.
	// It returns after the data is durable.
	Write(name string, data []byte) error
	// Read returns a backup's content, or ErrNotFound.
	Read(name string) ([]byte, error)
	// Delete removes a backup. Deleting a missing backup is not an error.
	Delete(name string) error
}

// BackupFile is a stored backup.
type BackupFile struct {
	Name string
	Size int64
}

// ValidBackupName reports whether name is acceptable as a backup name on
// every backend: 1–100 characters of [a-z0-9._-], not starting with a dot.
// Lower case only keeps names distinct on case-insensitive file systems.
func ValidBackupName(name string) error {
	if name == "" || len(name) > 100 || name[0] == '.' {
		return fmt.Errorf("storage: invalid backup name %q", name)
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-') {
			return fmt.Errorf("storage: invalid backup name %q", name)
		}
	}
	return nil
}
