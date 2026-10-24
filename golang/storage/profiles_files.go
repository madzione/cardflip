//go:build !js

package storage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// ProfileDataDir returns the data directory of profile id for OpenBackend:
// dataDir itself for the default profile, dataDir/profiles/<id> for others.
func ProfileDataDir(dataDir, id string) string {
	if id == DefaultProfile {
		return dataDir
	}
	return filepath.Join(dataDir, "profiles", id)
}

// RemoveProfileData deletes everything profile id stored under dataDir.
func RemoveProfileData(dataDir, id string) error {
	dir := ProfileDataDir(dataDir, id)
	if err := RemoveBackend(dir); err != nil {
		return err
	}
	if id == DefaultProfile {
		// dataDir also holds the registry and the other profiles.
		if err := os.Remove(syncLinksFile(dataDir, id)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	return os.RemoveAll(dir)
}

func profilesFile(dataDir string) string { return filepath.Join(dataDir, "profiles.json") }

// ReadProfiles returns the stored profile registry, or ErrNotFound when
// none was written yet.
func ReadProfiles(dataDir string) ([]byte, error) {
	if dataDir == "" {
		return nil, errors.New("storage: no data directory")
	}
	data, err := os.ReadFile(profilesFile(dataDir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

// WriteProfiles replaces the stored profile registry atomically.
func WriteProfiles(dataDir string, data []byte) error {
	if dataDir == "" {
		return errors.New("storage: no data directory")
	}
	return writeAtomic(profilesFile(dataDir), data)
}

func syncFile(dataDir string) string { return filepath.Join(dataDir, "sync.json") }

func syncLinksFile(dataDir, id string) string {
	return filepath.Join(ProfileDataDir(dataDir, id), "sync-links.json")
}

// ReadSync returns the stored sync state of the device (its key, name,
// trusted networks and paired devices), or ErrNotFound when none was
// written yet. Like the profile registry it lives in the data directory,
// outside every collection, so backups do not include it.
func ReadSync(dataDir string) ([]byte, error) { return readDataFile(dataDir, syncFile(dataDir)) }

// WriteSync replaces the stored sync state of the device atomically. The
// file is readable only by its owner, as it holds the device's private key.
func WriteSync(dataDir string, data []byte) error {
	return writeDataFile(dataDir, syncFile(dataDir), data)
}

// ReadSyncLinks returns the stored links of profile id, or ErrNotFound
// when none were written yet. The file sits next to the profile's
// collection backend, not in it, so backups do not include it.
func ReadSyncLinks(dataDir, id string) ([]byte, error) {
	return readDataFile(dataDir, syncLinksFile(dataDir, id))
}

// WriteSyncLinks replaces the stored links of profile id atomically.
func WriteSyncLinks(dataDir, id string, data []byte) error {
	return writeDataFile(dataDir, syncLinksFile(dataDir, id), data)
}

func readDataFile(dataDir, path string) ([]byte, error) {
	if dataDir == "" {
		return nil, errors.New("storage: no data directory")
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

func writeDataFile(dataDir, path string, data []byte) error {
	if dataDir == "" {
		return errors.New("storage: no data directory")
	}
	// writeAtomic renames a temporary file created with mode 0600.
	return writeAtomic(path, data)
}
