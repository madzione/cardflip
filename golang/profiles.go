package main

import (
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/text/unicode/norm"

	"cardflip/golang/core"
	"cardflip/golang/internal/yield"
	"cardflip/golang/storage"
)

// Profile is a separate collection with its own media, settings, backups
// and review history.
type Profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ProfileList lists the profiles by name.
type ProfileList struct {
	Profiles []Profile `json:"profiles"`
	// Current is the ID of the open profile; empty before one is opened.
	Current string `json:"current"`
	// ShowAtStartup is set when the app shows the profile list on start
	// instead of opening the last used profile.
	ShowAtStartup bool `json:"showAtStartup"`
}

// profileRegistry is the stored list of profiles, kept outside every
// collection (storage.WriteProfiles).
type profileRegistry struct {
	Profiles      []Profile `json:"profiles"`
	Last          string    `json:"last"`
	ShowAtStartup bool      `json:"showAtStartup,omitempty"`
	// Next numbers the next new profile's ID. IDs are never reused, so a
	// new profile never finds data a deleted one left behind.
	Next int `json:"next"`
}

// defaultRegistry is the registry of an installation without one: its
// collection, where it always was, is the profile Default.
func defaultRegistry() *profileRegistry {
	return &profileRegistry{Profiles: []Profile{{ID: storage.DefaultProfile, Name: "Default"}},
		Last: storage.DefaultProfile, Next: 2}
}

func (r *profileRegistry) find(id string) int {
	return slices.IndexFunc(r.Profiles, func(p Profile) bool { return p.ID == id })
}

// startProfile is the profile opened on start: the last used one.
func (r *profileRegistry) startProfile() string {
	if r.find(r.Last) >= 0 {
		return r.Last
	}
	return r.Profiles[0].ID
}

func (r *profileRegistry) clone() *profileRegistry {
	out := *r
	out.Profiles = slices.Clone(r.Profiles)
	return &out
}

// checkName returns name as it is stored, or why it cannot be the name of
// profile id (empty for a new profile).
func (r *profileRegistry) checkName(id, name string) (string, error) {
	name = norm.NFC.String(strings.TrimSpace(name))
	if name == "" {
		return "", core.NewError("profileNameEmpty", "A profile needs a name.")
	}
	for _, p := range r.Profiles {
		if p.ID != id && strings.EqualFold(p.Name, name) {
			return "", core.NewError("profileNameTaken", "A profile named %q already exists.", p.Name)
		}
	}
	return name, nil
}

func errProfileGone() error {
	return core.NewError("profileGone", "This profile no longer exists.")
}

// loadProfilesLocked reads the registry kept in root once.
func (a *app) loadProfilesLocked(root string) (*profileRegistry, error) {
	if a.profiles != nil {
		return a.profiles, nil
	}
	data, err := a.readProfiles(root)
	reg := defaultRegistry()
	switch {
	case errors.Is(err, storage.ErrNotFound):
	case err != nil:
		return nil, errors.New(usermsg.Text("profilesUnreadable", "The list of profiles could not be read: %v", err))
	default:
		reg = &profileRegistry{}
		if err := json.Unmarshal(data, reg); err != nil || len(reg.Profiles) == 0 {
			if err == nil {
				err = errors.New("no profiles")
			}
			return nil, errors.New(usermsg.Text("profilesUnreadable", "The list of profiles could not be read: %v", err))
		}
	}
	a.root, a.profiles = root, reg
	return reg, nil
}

// saveProfilesLocked stores reg and makes it the registry.
func (a *app) saveProfilesLocked(reg *profileRegistry) error {
	data, err := json.Marshal(reg)
	if err != nil {
		return err
	}
	if err := a.writeProfiles(a.root, data); err != nil {
		return errors.New(usermsg.Text("profilesNotSaved", "The list of profiles could not be saved: %v", err))
	}
	a.profiles = reg
	return nil
}

// registryLocked returns the registry, which Open or LoadProfiles loaded.
func (a *app) registryLocked() (*profileRegistry, error) {
	if a.profiles == nil {
		return nil, errNotOpened
	}
	return a.profiles, nil
}

func (a *app) profileListLocked() ProfileList {
	out := ProfileList{Profiles: slices.Clone(a.profiles.Profiles), ShowAtStartup: a.profiles.ShowAtStartup}
	if a.opened {
		out.Current = a.profile
	}
	slices.SortStableFunc(out.Profiles, func(x, y Profile) int {
		return strings.Compare(strings.ToLower(x.Name), strings.ToLower(y.Name))
	})
	return out
}

// switchProfile closes the open collection, if any, and opens profile id's.
func (a *app) switchProfile(id string) (Status, error) {
	a.mu.Lock()
	reg, err := a.registryLocked()
	if err != nil {
		a.mu.Unlock()
		return Status{}, err
	}
	if reg.find(id) < 0 {
		a.mu.Unlock()
		return Status{}, errProfileGone()
	}
	if a.opened && a.profile == id {
		defer a.mu.Unlock()
		return a.statusLocked(), nil
	}
	if reg.Last != id {
		next := reg.clone()
		next.Last = id
		if err := a.saveProfilesLocked(next); err != nil {
			a.mu.Unlock()
			return Status{}, err
		}
	}
	a.closeLocked()
	st := a.openLocked(id)
	// Change streams follow to the new collection.
	close(a.replaced)
	a.replaced = make(chan struct{})
	resets := a.resets
	a.mu.Unlock()
	// Services take their own locks, never while holding app.mu.
	for _, reset := range resets {
		reset()
	}
	return st, nil
}

// profileCardCount counts the cards of profile id without opening it in
// the app. A profile that stored nothing yet has none.
func (a *app) profileCardCount(id string) (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	reg, err := a.registryLocked()
	if err != nil {
		return 0, err
	}
	if reg.find(id) < 0 {
		return 0, errProfileGone()
	}
	if a.opened && a.profile == id {
		if a.col == nil {
			return 0, errors.New(a.status.Message)
		}
		return a.col.CardCount(), nil
	}
	b, err := a.openBackend(a.profileDir(a.root, id))
	if err != nil {
		return 0, errors.New(usermsg.Text("storageOpenFailed", "The collection storage could not be opened: %v", err))
	}
	defer b.Close()
	if keys, err := b.List(""); err != nil {
		return 0, err
	} else if len(keys) == 0 {
		return 0, nil
	}
	c, _, err := core.Open(b, core.Options{Now: a.now, Yield: yield.Yield, Names: a.names})
	if err != nil {
		return 0, errors.New(usermsg.Text("collectionUnreadable", "Your collection could not be read: %v", err))
	}
	defer c.Close()
	return c.CardCount(), nil
}

// LoadProfiles reads the profiles kept in dataDir (empty on the web)
// without opening a collection, so that the app can show the profile list
// on start; SwitchProfile then opens one. An installation from before
// profiles has one, Default, with its collection.
func (s *AppService) LoadProfiles(dataDir string) (ProfileList, error) {
	list, err := s.loadProfiles(dataDir)
	if err == nil {
		s.a.rootLoaded()
	}
	return list, err
}

func (s *AppService) loadProfiles(dataDir string) (ProfileList, error) {
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	if _, err := s.a.loadProfilesLocked(dataDir); err != nil {
		return ProfileList{}, err
	}
	return s.a.profileListLocked(), nil
}

// Profiles lists the profiles.
func (s *AppService) Profiles() (ProfileList, error) {
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	if _, err := s.a.registryLocked(); err != nil {
		return ProfileList{}, err
	}
	return s.a.profileListLocked(), nil
}

// AddProfile adds a profile with an empty collection, made when it is first
// opened. Names are unique, ignoring case.
func (s *AppService) AddProfile(name string) (Profile, error) {
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	reg, err := s.a.registryLocked()
	if err != nil {
		return Profile{}, err
	}
	name, err = reg.checkName("", name)
	if err != nil {
		return Profile{}, err
	}
	next := reg.clone()
	p := Profile{ID: fmt.Sprintf("p%d", next.Next), Name: name}
	next.Next++
	next.Profiles = append(next.Profiles, p)
	if err := s.a.saveProfilesLocked(next); err != nil {
		return Profile{}, err
	}
	return p, nil
}

// RenameProfile renames profile id.
func (s *AppService) RenameProfile(id, name string) error {
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	reg, err := s.a.registryLocked()
	if err != nil {
		return err
	}
	i := reg.find(id)
	if i < 0 {
		return errProfileGone()
	}
	name, err = reg.checkName(id, name)
	if err != nil {
		return err
	}
	next := reg.clone()
	next.Profiles[i].Name = name
	return s.a.saveProfilesLocked(next)
}

// ProfileCardCount returns how many cards profile id has, for the
// confirmation before deleting it.
func (s *AppService) ProfileCardCount(id string) (int, error) { return s.a.profileCardCount(id) }

// DeleteProfile deletes profile id and all its data: collection, media and
// backups. The open profile and the only one cannot be deleted. The app
// only calls it after the user confirmed.
func (s *AppService) DeleteProfile(id string) error {
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	reg, err := s.a.registryLocked()
	if err != nil {
		return err
	}
	i := reg.find(id)
	switch {
	case i < 0:
		return errProfileGone()
	case s.a.opened && s.a.profile == id:
		return core.NewError("profileOpen", "The open profile cannot be deleted. Open another profile first.")
	case len(reg.Profiles) == 1:
		return core.NewError("profileOnly", "The only profile cannot be deleted.")
	}
	// The data first: when that fails, the profile is still listed and
	// deleting it can be tried again.
	if err := s.a.removeProfile(s.a.root, id); err != nil {
		return err
	}
	next := reg.clone()
	next.Profiles = slices.Delete(next.Profiles, i, i+1)
	if next.Last == id {
		next.Last = next.Profiles[0].ID
		if s.a.opened {
			next.Last = s.a.profile
		}
	}
	return s.a.saveProfilesLocked(next)
}

// SwitchProfile closes the open collection and opens profile id's, as the
// app does on start; the study session and staged files are dropped. It
// becomes the last used profile.
func (s *AppService) SwitchProfile(id string) (Status, error) { return s.a.switchProfile(id) }

// SetShowProfilesAtStartup sets whether the app shows the profile list on
// start instead of opening the last used profile.
func (s *AppService) SetShowProfilesAtStartup(on bool) error {
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	reg, err := s.a.registryLocked()
	if err != nil {
		return err
	}
	next := reg.clone()
	next.ShowAtStartup = on
	return s.a.saveProfilesLocked(next)
}
