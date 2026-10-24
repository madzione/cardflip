package main

import (
	"cardflip/golang/internal/usermsg"
	"context"
	"errors"
	"sync"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/internal/yield"
	"cardflip/golang/render"
	"cardflip/golang/storage"
)

// errNotOpened is returned by every service until app.Open succeeded.
var errNotOpened = errors.New("not opened: the collection has not been opened yet")

// Status states.
const (
	StatusOK      = "ok"
	StatusFresh   = "fresh"
	StatusCorrupt = "corrupt"
)

// app is the state the services share: the opened collection, or why it
// could not be opened.
type app struct {
	mu     sync.Mutex
	opened bool
	// root is the data directory the profiles are kept under (empty on the
	// web), profiles their registry once loaded, profile the ID of the open
	// profile and dataDir its data directory.
	root     string
	profiles *profileRegistry
	profile  string
	dataDir  string
	backend  storage.Backend
	col      *core.Collection
	status   Status
	// replaced is closed when col is replaced (new collection after a
	// corrupt start), so change streams move to the new one.
	replaced chan struct{}
	now      func() time.Time

	// env is where cards are shown; envChanged is closed when it changes,
	// so change streams tell open card views to render again.
	env        render.Env
	envChanged chan struct{}

	// names are the built-in names in the interface language, which
	// collections are opened with.
	names core.BuiltinNames

	// openBackend is storage.OpenBackend; tests replace it.
	openBackend   func(dataDir string) (storage.Backend, error)
	removeBackend func(dataDir string) error
	persist       func() bool

	// The profile storage functions of package storage; tests replace them.
	readProfiles  func(root string) ([]byte, error)
	writeProfiles func(root string, data []byte) error
	profileDir    func(root, id string) string
	removeProfile func(root, id string) error

	// onOpen are called with the data root once it is known (see
	// rootLoaded), without the app's lock held.
	onOpen     []func(root string)
	onOpenOnce sync.Once

	// resets clear what services keep about the open collection (a study
	// session, staged files) when another profile is opened.
	resets []func()

	backups autoBackups

	// shared holds the collections open in this process by profile, with
	// how many users each has: the open profile and sync sessions serving
	// profiles of linked devices (sync design decision 12). A profile is
	// opened once, however many use it.
	shared map[string]*sharedCollection
	// syncDevice is this device's sync ID, which collections are told when
	// they are opened (see core.Collection.SetDevice); empty until the
	// device has a sync identity.
	syncDevice string
}

// sharedCollection is a collection open in this process and its users.
type sharedCollection struct {
	col     *core.Collection
	backend storage.Backend
	refs    int
}

func newApp() *app {
	return &app{now: time.Now, openBackend: storage.OpenBackend, removeBackend: storage.RemoveBackend,
		persist: storage.RequestPersistence, replaced: make(chan struct{}), envChanged: make(chan struct{}),
		backups:      autoBackups{tick: time.Minute},
		readProfiles: storage.ReadProfiles, writeProfiles: storage.WriteProfiles,
		profileDir: storage.ProfileDataDir, removeProfile: storage.RemoveProfileData,
		shared: map[string]*sharedCollection{}}
}

// Status tells the app what it can show after opening the collection.
type Status struct {
	// State is "ok", "fresh" (a new collection was created) or "corrupt"
	// (stored data could not be read; it was left untouched).
	State   string `json:"state"`
	Message string `json:"message,omitempty"`
	// Persistent is false when the browser may evict the stored data.
	Persistent bool `json:"persistent"`
	// ShowPersistenceNotice is set once, when storage is not persistent.
	ShowPersistenceNotice bool   `json:"showPersistenceNotice"`
	Version               string `json:"version"`
}

// open opens the collection of the last used profile in dataDir.
func (a *app) open(dataDir string) (Status, error) {
	st, err := a.openOnce(dataDir)
	if err == nil {
		a.rootLoaded()
	}
	return st, err
}

func (a *app) openOnce(dataDir string) (Status, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.opened {
		// Hot restart: the collection stays open.
		return a.statusLocked(), nil
	}
	reg, err := a.loadProfilesLocked(dataDir)
	if err != nil {
		return Status{}, err
	}
	return a.openLocked(reg.startProfile()), nil
}

// rootLoaded runs onOpen once the profile registry, and with it the data
// root, was loaded: on start, whether the app opens a collection or shows
// the profile list. A hot restart loads nothing new and runs nothing.
func (a *app) rootLoaded() {
	a.onOpenOnce.Do(func() {
		a.mu.Lock()
		root := a.root
		a.mu.Unlock()
		for _, f := range a.onOpen {
			f(root)
		}
	})
}

// openLocked opens the collection of profile id; none is open. A
// collection a sync session has open is used as it is.
func (a *app) openLocked(id string) Status {
	dir := a.profileDir(a.root, id)
	a.profile, a.dataDir, a.opened = id, dir, true
	if sc := a.shared[id]; sc != nil {
		sc.refs++
		a.backend, a.col = sc.backend, sc.col
		a.status = Status{State: StatusOK, Persistent: a.persist()}
		a.backups.opened(sc.col, sc.backend)
		a.startBackupTimer()
		return a.statusLocked()
	}
	b, err := a.openBackend(dir)
	if err != nil {
		a.status = Status{State: StatusCorrupt, Message: usermsg.Text("storageOpenFailed", "The collection storage could not be opened: %v", err)}
		return a.statusLocked()
	}
	a.backend = b
	persistent := a.persist()
	c, fresh, err := core.Open(b, core.Options{Now: a.now, Yield: yield.Yield, Names: a.names})
	if err != nil {
		a.status = Status{State: StatusCorrupt, Persistent: persistent,
			Message: usermsg.Text("collectionUnreadable", "Your collection could not be read: %v", err)}
		return a.statusLocked()
	}
	a.col = c
	c.SetDevice(a.syncDevice)
	a.shared[id] = &sharedCollection{col: c, backend: b, refs: 1}
	a.status = Status{State: StatusOK, Persistent: persistent}
	if fresh {
		a.status.State = StatusFresh
	}
	a.backups.opened(c, b)
	a.startBackupTimer()
	return a.statusLocked()
}

// closeLocked closes the open collection and its storage, after a backup
// of it that is being made.
func (a *app) closeLocked() {
	a.backups.mu.Lock()
	defer a.backups.mu.Unlock()
	if sc := a.shared[a.profile]; sc != nil && sc.col == a.col && a.col != nil {
		a.releaseLocked(a.profile, sc)
	} else {
		if a.col != nil {
			a.col.Close()
		}
		if a.backend != nil {
			a.backend.Close()
		}
	}
	a.col, a.backend = nil, nil
	a.backups.col = nil
}

// releaseLocked ends one use of profile id's shared collection sc, and
// closes it after the last.
func (a *app) releaseLocked(id string, sc *sharedCollection) {
	if sc.refs--; sc.refs > 0 {
		return
	}
	sc.col.Close()
	sc.backend.Close()
	if a.shared[id] == sc {
		delete(a.shared, id)
	}
}

// acquireCollection returns profile id's collection for a sync session,
// opening it if no one has, and the function that ends the use.
func (a *app) acquireCollection(id string) (*core.Collection, storage.Backend, func(), error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	sc := a.shared[id]
	if sc == nil {
		reg, err := a.registryLocked()
		if err != nil {
			return nil, nil, nil, err
		}
		if reg.find(id) < 0 {
			return nil, nil, nil, errProfileGone()
		}
		b, err := a.openBackend(a.profileDir(a.root, id))
		if err != nil {
			return nil, nil, nil, err
		}
		c, _, err := core.Open(b, core.Options{Now: a.now, Yield: yield.Yield, Names: a.names})
		if err != nil {
			b.Close()
			return nil, nil, nil, err
		}
		c.SetDevice(a.syncDevice)
		sc = &sharedCollection{col: c, backend: b}
		a.shared[id] = sc
	}
	sc.refs++
	var once sync.Once
	return sc.col, sc.backend, func() {
		once.Do(func() {
			a.mu.Lock()
			defer a.mu.Unlock()
			a.releaseLocked(id, sc)
		})
	}, nil
}

// setSyncDevice records this device's sync ID and tells the open
// collections.
func (a *app) setSyncDevice(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.syncDevice = id
	for _, sc := range a.shared {
		sc.col.SetDevice(id)
	}
}

func (a *app) statusLocked() Status {
	s := a.status
	s.Version = appVersion
	if a.col != nil && !s.Persistent {
		s.ShowPersistenceNotice = !a.col.Settings().PersistenceNoticeShown
	}
	return s
}

// collection returns the opened collection.
func (a *app) collection() (*core.Collection, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.col == nil {
		if a.opened && a.status.State == StatusCorrupt {
			return nil, errors.New(a.status.Message)
		}
		return nil, errNotOpened
	}
	return a.col, nil
}

func (a *app) reload() (Status, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.opened {
		return Status{}, errNotOpened
	}
	if a.backend != nil {
		a.backend.Close()
	}
	b, err := a.openBackend(a.dataDir)
	if err != nil {
		return Status{}, err
	}
	a.backend = b
	c, fresh, err := core.Open(b, core.Options{Now: a.now, Yield: yield.Yield, Names: a.names})
	if err != nil {
		if a.col != nil {
			a.col.Close()
			a.col = nil
		}
		a.status = Status{State: StatusCorrupt, Persistent: a.status.Persistent,
			Message: usermsg.Text("collectionUnreadable", "Your collection could not be read: %v", err)}
		return a.statusLocked(), nil
	}
	a.setCollection(c)
	a.status.State, a.status.Message = StatusOK, ""
	if fresh {
		a.status.State = StatusFresh
	}
	return a.statusLocked(), nil
}

// startFresh deletes unreadable data and creates a new collection; the
// user confirmed it on the error screen.
func (a *app) startFresh() (Status, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.opened {
		return Status{}, errNotOpened
	}
	if a.col != nil {
		return Status{}, core.NewError("nothingToReplace", "the collection is readable; there is nothing to replace")
	}
	if a.backend == nil {
		b, err := a.openBackend(a.dataDir)
		if err != nil {
			return Status{}, errors.New(usermsg.Text("storageOpenFailed", "The collection storage could not be opened: %v", err))
		}
		a.backend = b
	}
	if err := resetStorage(a.backend); err != nil {
		return Status{}, err
	}
	c, _, err := core.Open(a.backend, core.Options{Now: a.now, Yield: yield.Yield, Names: a.names})
	if err != nil {
		return Status{}, err
	}
	a.setCollection(c)
	a.status.State, a.status.Message = StatusFresh, ""
	return a.statusLocked(), nil
}

// restoreWith replaces the collection with restore, after backing up the
// current one. When the stored collection could not be read, the user chose
// to restore over it: a fresh collection is created first.
func (a *app) restoreWith(restore func(c *core.Collection) error) error {
	c, err := a.collection()
	if err != nil {
		if errors.Is(err, errNotOpened) {
			return err
		}
		if _, err := a.startFresh(); err != nil {
			return err
		}
		if c, err = a.collection(); err != nil {
			return err
		}
		if err := restore(c); err != nil {
			return err
		}
		a.mu.Lock()
		a.status.State = StatusOK
		a.mu.Unlock()
		return nil
	}
	if err := a.backupBeforeOperation(); err != nil {
		return err
	}
	return restore(c)
}

func (a *app) setCollection(c *core.Collection) {
	// The caller dealt with the old backend (reload closed it).
	if sc := a.shared[a.profile]; sc != nil && sc.col == a.col {
		delete(a.shared, a.profile)
	}
	if a.col != nil {
		a.col.Close()
	}
	a.col = c
	c.SetDevice(a.syncDevice)
	a.shared[a.profile] = &sharedCollection{col: c, backend: a.backend, refs: 1}
	a.backups.opened(c, a.backend)
	close(a.replaced)
	a.replaced = make(chan struct{})
}

func resetStorage(b storage.Backend) error {
	keys, err := b.List("")
	if err != nil {
		return err
	}
	for _, k := range keys {
		if err := b.Delete(k); err != nil {
			return err
		}
	}
	return nil
}

func (a *app) setBuiltinNames(names core.BuiltinNames) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.names = names
	if a.col != nil {
		a.col.SetBuiltinNames(names)
	}
}

// renderEnv returns where cards are shown.
func (a *app) renderEnv() render.Env {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.env
}

func (a *app) setRenderEnv(env render.Env) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if env == a.env {
		return
	}
	a.env = env
	close(a.envChanged)
	a.envChanged = make(chan struct{})
}

// changes streams collection changes until ctx ends, following the
// collection when it is replaced, and a change of kind "environment" when
// the render environment changes. The first subscription is made before it
// returns, so no change after the call is missed.
func (a *app) changes(ctx context.Context) <-chan core.Change {
	out := make(chan core.Change)
	subscribe := func() (<-chan core.Change, func(), <-chan struct{}) {
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.col == nil {
			return nil, func() {}, a.replaced
		}
		events, cancel := a.col.Subscribe()
		return events, cancel, a.replaced
	}
	events, cancel, replaced := subscribe()
	envChanged := func() <-chan struct{} {
		a.mu.Lock()
		defer a.mu.Unlock()
		return a.envChanged
	}
	env := envChanged()
	go func() {
		defer close(out)
		for {
		loop:
			for {
				select {
				case <-ctx.Done():
					cancel()
					return
				case <-replaced:
					cancel()
					break loop
				case <-env:
					env = envChanged()
					select {
					case out <- core.Change{Kind: "environment"}:
					case <-ctx.Done():
						cancel()
						return
					}
				case ch, ok := <-events:
					if !ok {
						events = nil
						continue
					}
					select {
					case out <- ch:
					case <-ctx.Done():
						cancel()
						return
					}
				}
			}
			events, cancel, replaced = subscribe()
			// A new collection: everything changed.
			select {
			case out <- core.Change{Kind: "replaced", All: true}:
			case <-ctx.Done():
				cancel()
				return
			}
		}
	}()
	return out
}
