//go:build js

package storage

import "errors"

// ProfileDataDir returns the data directory of profile id for OpenBackend,
// which names its database: the app's own database for the default
// profile, cardflip-profile-<id> for others.
func ProfileDataDir(dataDir, id string) string {
	if id == DefaultProfile {
		return dataDir
	}
	if dataDir == "" {
		return "profile-" + id
	}
	return dataDir + "-profile-" + id
}

// RemoveProfileData deletes everything profile id stored.
func RemoveProfileData(dataDir, id string) error {
	return RemoveBackend(ProfileDataDir(dataDir, id))
}

// profilesKey is the registry's key in its own database, which no profile
// uses.
const profilesKey = "registry"

func profilesDB(dataDir string) string { return dbName(dataDir) + "-profiles" }

// ReadProfiles returns the stored profile registry, or ErrNotFound when
// none was written yet.
func ReadProfiles(dataDir string) ([]byte, error) {
	b, err := OpenIndexedDB(profilesDB(dataDir))
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return b.Get(profilesKey)
}

// WriteProfiles replaces the stored profile registry atomically.
func WriteProfiles(dataDir string, data []byte) error {
	b, err := OpenIndexedDB(profilesDB(dataDir))
	if err != nil {
		return err
	}
	defer b.Close()
	return b.Put(profilesKey, data)
}

var errSyncUnsupported = errors.New("storage: device sync is not available on the web")

// ReadSync is not available on the web, which cannot sync.
func ReadSync(dataDir string) ([]byte, error) { return nil, errSyncUnsupported }

// WriteSync is not available on the web, which cannot sync.
func WriteSync(dataDir string, data []byte) error { return errSyncUnsupported }

// ReadSyncLinks is not available on the web, which cannot sync.
func ReadSyncLinks(dataDir, id string) ([]byte, error) { return nil, errSyncUnsupported }

// WriteSyncLinks is not available on the web, which cannot sync.
func WriteSyncLinks(dataDir, id string, data []byte) error { return errSyncUnsupported }
