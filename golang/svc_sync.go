package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"cardflip/golang/backup"
	"cardflip/golang/core"
	"cardflip/golang/internal/yield"
	"cardflip/golang/storage"
	cfsync "cardflip/golang/sync"
	"cardflip/golang/syncnet"
)

// SyncUI is how Go asks the app's user during pairing and the first sync
// of a link.
//
//fluffy:dart
type SyncUI interface {
	// PairingRequest asks whether to accept a pairing request from the
	// device named deviceName.
	PairingRequest(ctx context.Context, deviceName string) (bool, error)
	// ConfirmPairing shows the pairing code and asks whether the device
	// named deviceName shows the same.
	ConfirmPairing(ctx context.Context, deviceName, code string) (bool, error)
	// PairingFinished tells whether pairing with deviceName succeeded.
	PairingFinished(deviceName string, paired bool)
	// ChooseFirstSync asks how to combine the open profile with the linked
	// one at its first sync: "merge", "replaceOther" or "replaceThis", or
	// "" to decide later.
	ChooseFirstSync(ctx context.Context, link LinkInfo) (string, error)
}

// Scheduling of automatic syncs (sync design decision 11); variables for
// tests.
var (
	// syncDebounce is how long after the last local change the open
	// profile syncs.
	syncDebounce = 30 * time.Second
	// syncBackgroundBudget is how long the sync when the app goes to the
	// background may take.
	syncBackgroundBudget = 5 * time.Second
)

// Device name limits (device-pairing spec: Device identity).
const maxDeviceName = 40

var errSyncUnsupported = errors.New("device sync is not available on the web")

// syncState is sync.json in the data root: device state, the same for every
// profile and never part of a collection or backup.
type syncState struct {
	Key      []byte            `json:"key"`
	Name     string            `json:"name"`
	Networks []TrustedNetwork  `json:"networks,omitempty"`
	Paired   []syncnet.Device  `json:"paired,omitempty"`
	Revoked  []syncnet.Revoked `json:"revoked,omitempty"`
}

// TrustedNetwork is a network sync may use: a Wi-Fi network by its name,
// or another network by its subnet.
type TrustedNetwork struct {
	ID string `json:"id"`
	// WiFi is the Wi-Fi name; empty for a network known by its subnet.
	WiFi   string `json:"wifi,omitempty"`
	Subnet string `json:"subnet,omitempty"`
	// Wired is set for a cable network.
	Wired bool `json:"wired,omitempty"`
}

// Network kinds the app reports (SetNetwork).
const (
	NetworkWiFi     = "wifi"
	NetworkEthernet = "ethernet"
	NetworkCellular = "cellular"
	NetworkOther    = "other"
	NetworkNone     = "none"
)

// currentNetwork is the network the device is on, as the app reported it.
type currentNetwork struct {
	Kind   string
	WiFi   string
	Subnet string
}

// id is the network's identity: "wifi:<name>" or "subnet:<CIDR>"; empty
// for cellular networks, which are never trusted, and for no network.
func (n currentNetwork) id() string {
	switch {
	case n.Kind == NetworkCellular || n.Kind == NetworkNone || n.Kind == "":
		return ""
	case n.Kind == NetworkWiFi && n.WiFi != "":
		return "wifi:" + n.WiFi
	case n.Subnet != "":
		return "subnet:" + n.Subnet
	}
	return ""
}

// linkRecord is a link as sync-links.json stores it.
type linkRecord struct {
	cfsync.Link
	// RemoteName is the linked profile's name on the other device.
	RemoteName string `json:"remoteName"`
	// Error is why the last sync of the link failed (a tagged text).
	Error string `json:"error,omitempty"`
}

// seenDevice is a device the app found on the network.
type seenDevice struct {
	Name    string
	Addr    string
	Pairing bool
}

// NetworkInfo describes the current network for the Sync section.
type NetworkInfo struct {
	Kind    string `json:"kind"`
	WiFi    string `json:"wifi,omitempty"`
	Subnet  string `json:"subnet,omitempty"`
	ID      string `json:"id,omitempty"`
	Trusted bool   `json:"trusted"`
}

// PairedDeviceInfo describes a paired device for the Sync section.
type PairedDeviceInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Reachable bool   `json:"reachable"`
	// LastSynced is when the open profile last synced with it.
	LastSynced time.Time `json:"lastSynced,omitzero"`
}

// DiscoveredDevice is an unpaired device visible for pairing.
type DiscoveredDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// SyncStatus is the state of device sync, streamed to the app.
type SyncStatus struct {
	// Supported is false on the web; Ready once the device has its sync
	// identity (the Sync section was opened).
	Supported  bool   `json:"supported"`
	Ready      bool   `json:"ready"`
	DeviceID   string `json:"deviceId,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
	// DefaultName is what the device is called until the user names it.
	Network NetworkInfo      `json:"network"`
	Trusted []TrustedNetwork `json:"trusted"`
	// Listening is set while sync accepts connections, on Port, which the
	// app announces; Address is this device's address for "Pair a device".
	Listening bool   `json:"listening"`
	Port      int    `json:"port,omitempty"`
	Address   string `json:"address,omitempty"`
	// Pairing is set while this device is visible for pairing.
	Pairing    bool               `json:"pairing"`
	Paired     []PairedDeviceInfo `json:"paired"`
	Discovered []DiscoveredDevice `json:"discovered"`
	// Links are the open profile's links.
	Links []LinkInfo `json:"links"`
	// Syncing is set while the open profile syncs, with Progress.
	Syncing  bool             `json:"syncing"`
	Progress *cfsync.Progress `json:"progress,omitempty"`
	// LastSynced is the open profile's last finished sync; Error why the
	// last one failed (a tagged text), empty if it did not.
	LastSynced time.Time `json:"lastSynced,omitzero"`
	Error      string    `json:"error,omitempty"`
}

// SyncService syncs the open profile with linked profiles on paired
// devices on trusted local networks (device-pairing and device-sync
// specs). Nothing runs until the Sync section initialises it, and no
// network code runs before a network is trusted.
type SyncService struct {
	a  *app
	ui SyncUI

	mu          sync.Mutex
	root        string
	state       syncState
	node        *syncnet.Node
	network     currentNetwork
	defaultName string
	listener    *syncnet.Listener
	conns       map[*syncnet.Conn]bool
	seen        map[string]seenDevice
	foreground  bool
	// outgoing are the peers this device has a session with that it
	// started; syncing counts the open profile's running sessions.
	outgoing map[string]bool
	syncing  int
	progress *cfsync.Progress
	lastErr  string
	running  bool // an automatic sync of the open profile is running
	again    bool // and another is due after it
	debounce *time.Timer
	changed  chan struct{}

	// The storage functions and the platform check; tests replace them.
	read       func(root string) ([]byte, error)
	write      func(root string, data []byte) error
	readLinks  func(root, profile string) ([]byte, error)
	writeLinks func(root, profile string, data []byte) error
	linksMu    sync.Mutex
	supported  bool
	// serving, if set, is called when a session a peer started has opened
	// the collection of profile (tests).
	serving func(profile string)
}

func newSyncService(a *app, ui SyncUI) *SyncService {
	s := &SyncService{a: a, ui: ui, conns: map[*syncnet.Conn]bool{}, seen: map[string]seenDevice{}, outgoing: map[string]bool{},
		changed: make(chan struct{}), foreground: true,
		read: storage.ReadSync, write: storage.WriteSync, readLinks: storage.ReadSyncLinks, writeLinks: storage.WriteSyncLinks,
		supported: syncnet.Built}
	a.onOpen = append(a.onOpen, s.opened)
	return s
}

// Supported reports whether this platform can sync: every platform but the
// web.
func (s *SyncService) Supported() bool { return s.supported }

// opened loads the device's sync state once the app knows its data root.
// A device that never opened the Sync section has none, and nothing runs.
func (s *SyncService) opened(root string) {
	if !s.supported {
		return
	}
	s.mu.Lock()
	s.root = root
	data, err := s.read(root)
	if err == nil {
		var st syncState
		if err = json.Unmarshal(data, &st); err == nil {
			err = s.loadLocked(st)
		}
	}
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		s.lastErr = core.NewError("syncStateUnreadable", "The sync settings could not be read: %v", err).Error()
	}
	s.notifyLocked()
	s.mu.Unlock()
	go s.watchChanges()
}

// loadLocked starts the device's node from st.
func (s *SyncService) loadLocked(st syncState) error {
	id, err := syncnet.IdentityFromSeed(st.Key)
	if err != nil {
		return err
	}
	n := syncnet.NewNode(id, st.Name, s.ui, s.a.now, st.Paired, st.Revoked)
	n.Save = s.save
	n.OnUnpaired = s.removeLinksOf
	n.OnSession = s.serve
	n.OnPaired = func(syncnet.Device) { s.notify() }
	s.state, s.node = st, n
	s.a.setSyncDevice(id.ID())
	s.applyNetworkLocked()
	return nil
}

// save writes sync.json.
func (s *SyncService) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *SyncService) saveLocked() error {
	if s.node == nil {
		return nil
	}
	st := s.state
	st.Name, st.Paired, st.Revoked = s.node.Name(), s.node.Devices(), s.node.RevokedKeys()
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	if err := s.write(s.root, data); err != nil {
		return core.NewError("syncStateNotSaved", "The sync settings could not be saved: %v", err)
	}
	s.state = st
	s.notifyLocked()
	return nil
}

// Init creates the device's sync identity, if it has none, named
// defaultName (the computer's or phone's name; empty uses the host name).
// The Sync section calls it when it opens.
func (s *SyncService) Init(defaultName string) (SyncStatus, error) {
	if !s.supported {
		return SyncStatus{}, errSyncUnsupported
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.root == "" {
		return s.statusLocked(), errNotOpened
	}
	s.defaultName = defaultName
	if s.node != nil {
		return s.statusLocked(), nil
	}
	name := strings.TrimSpace(defaultName)
	if name == "" {
		name, _ = os.Hostname()
	}
	if name == "" {
		name = "Cardflip"
	}
	if utf8.RuneCountInString(name) > maxDeviceName {
		name = string([]rune(name)[:maxDeviceName])
	}
	if err := s.loadLocked(syncState{Key: syncnet.NewKey(), Name: name}); err != nil {
		return s.statusLocked(), err
	}
	if err := s.saveLocked(); err != nil {
		s.node = nil
		return s.statusLocked(), err
	}
	return s.statusLocked(), nil
}

// ready returns the node, or an error before Init.
func (s *SyncService) readyLocked() (*syncnet.Node, error) {
	if !s.supported {
		return nil, errSyncUnsupported
	}
	if s.node == nil {
		return nil, core.NewError("syncNotSetUp", "Open Settings → Sync first.")
	}
	return s.node, nil
}

// SetDeviceName renames this device, 1 to 40 characters; paired devices
// see the new name at their next connection.
func (s *SyncService) SetDeviceName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxDeviceName {
		return core.NewError("syncDeviceNameLength", "The device name must be 1 to %d characters long.", maxDeviceName)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err := s.readyLocked()
	if err != nil {
		return err
	}
	n.SetName(name)
	return s.saveLocked()
}

// SetNetwork tells sync which network the device is on: kind is "wifi",
// "ethernet", "cellular", "other" or "none"; wifi the Wi-Fi name, if the
// app may read it; subnet the network's subnet in CIDR notation, if the
// app knows it (Go finds it otherwise). Sync starts or stops listening
// accordingly.
func (s *SyncService) SetNetwork(kind, wifi, subnet string) {
	if !s.supported {
		return
	}
	if subnet == "" && kind != NetworkCellular && kind != NetworkNone {
		subnet, _ = syncnet.LocalNetwork()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.network = currentNetwork{Kind: kind, WiFi: wifi, Subnet: subnet}
	s.applyNetworkLocked()
	s.notifyLocked()
}

// trustedLocked reports whether the current network is trusted.
func (s *SyncService) trustedLocked() bool {
	id := s.network.id()
	return id != "" && slices.ContainsFunc(s.state.Networks, func(n TrustedNetwork) bool { return n.ID == id })
}

// applyNetworkLocked listens while the current network is trusted, and
// otherwise stops listening and closes every connection (device-pairing
// spec: Trusted networks only).
func (s *SyncService) applyNetworkLocked() {
	if s.node != nil && s.trustedLocked() {
		if s.listener == nil {
			l, err := syncnet.Listen(s.node.Identity(), syncnet.DefaultPort, s.node.Verify, s.accept)
			if err != nil {
				s.lastErr = core.NewError("syncListenFailed", "Sync could not start: %v", err).Error()
				return
			}
			s.listener = l
			go s.trigger()
		}
		return
	}
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
	for c := range s.conns {
		c.Close()
	}
	if s.node != nil {
		s.node.OpenPairing(0)
	}
}

// accept handles a connection the listener accepted.
func (s *SyncService) accept(c *syncnet.Conn) {
	if !s.track(c) {
		c.Close()
		return
	}
	defer s.untrack(c)
	s.node.Handle(c)
}

// track records an open connection; it reports false when the network is
// no longer trusted.
func (s *SyncService) track(c *syncnet.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.trustedLocked() {
		return false
	}
	s.conns[c] = true
	return true
}

func (s *SyncService) untrack(c *syncnet.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, c)
}

// TrustNetwork trusts the current network.
func (s *SyncService) TrustNetwork() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.readyLocked(); err != nil {
		return err
	}
	id := s.network.id()
	if id == "" {
		return core.NewError("syncNoNetwork", "This device is not on a network that can be trusted.")
	}
	if !s.trustedLocked() {
		n := TrustedNetwork{ID: id, Wired: s.network.Kind == NetworkEthernet}
		if strings.HasPrefix(id, "wifi:") {
			n.WiFi = s.network.WiFi
		} else {
			n.Subnet = s.network.Subnet
		}
		s.state.Networks = append(s.state.Networks, n)
		if err := s.saveLocked(); err != nil {
			s.state.Networks = s.state.Networks[:len(s.state.Networks)-1]
			return err
		}
	}
	s.applyNetworkLocked()
	return nil
}

// UntrustNetwork stops trusting network id; on it, sync stops at once.
func (s *SyncService) UntrustNetwork(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.readyLocked(); err != nil {
		return err
	}
	old := s.state.Networks
	s.state.Networks = slices.DeleteFunc(slices.Clone(old), func(n TrustedNetwork) bool { return n.ID == id })
	if err := s.saveLocked(); err != nil {
		s.state.Networks = old
		return err
	}
	s.applyNetworkLocked()
	return nil
}

// StartPairing makes this device visible for pairing for 5 minutes, or
// until StopPairing.
func (s *SyncService) StartPairing() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err := s.readyLocked()
	if err != nil {
		return err
	}
	if !s.trustedLocked() {
		return core.NewError("syncUntrusted", "Trust this network first.")
	}
	n.OpenPairing(syncnet.PairingWindow)
	s.notifyLocked()
	time.AfterFunc(syncnet.PairingWindow, s.notify)
	return nil
}

// StopPairing makes this device invisible for pairing.
func (s *SyncService) StopPairing() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.node != nil {
		s.node.OpenPairing(0)
		s.notifyLocked()
	}
}

// PairWith asks the device at host and port to pair (device-pairing spec:
// Pair a device, Pair by address).
func (s *SyncService) PairWith(ctx context.Context, host string, port int) error {
	s.mu.Lock()
	n, err := s.readyLocked()
	trusted := s.trustedLocked()
	s.mu.Unlock()
	if err != nil {
		return err
	}
	if !trusted {
		return core.NewError("syncUntrusted", "Trust this network first.")
	}
	if host == "" || port <= 0 || port > 65535 {
		return core.NewError("syncBadAddress", "Enter the address and port shown on the other device.")
	}
	_, err = n.Pair(ctx, net.JoinHostPort(host, strconv.Itoa(port)))
	return err
}

// Unpair unpairs a device and removes the links to it; the device is told
// now if it can be reached, otherwise when it next connects.
func (s *SyncService) Unpair(ctx context.Context, deviceID string) error {
	s.mu.Lock()
	n, err := s.readyLocked()
	addr := s.seen[deviceID].Addr
	s.mu.Unlock()
	if err != nil {
		return err
	}
	return n.Unpair(ctx, deviceID, addr)
}

// Seen tells sync about a device the app found on the network: its ID,
// name and address, and whether it is visible for pairing.
func (s *SyncService) Seen(deviceID, name, host string, port int, pairing bool) {
	if !s.supported {
		return
	}
	s.mu.Lock()
	_, known := s.seen[deviceID]
	s.seen[deviceID] = seenDevice{Name: name, Addr: net.JoinHostPort(host, strconv.Itoa(port)), Pairing: pairing}
	s.notifyLocked()
	s.mu.Unlock()
	if !known {
		// A linked device became reachable.
		go s.trigger()
	}
}

// Lost tells sync that a device left the network.
func (s *SyncService) Lost(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.seen, deviceID)
	s.notifyLocked()
}

// Foreground tells sync that the app is in the foreground: the open
// profile syncs.
func (s *SyncService) Foreground() {
	s.mu.Lock()
	s.foreground = true
	s.mu.Unlock()
	go s.trigger()
}

// Background tells sync that the app went to the background: the open
// profile syncs once more, for at most 5 seconds.
func (s *SyncService) Background() {
	s.mu.Lock()
	s.foreground = false
	if s.debounce != nil {
		s.debounce.Stop()
	}
	s.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), syncBackgroundBudget)
	defer cancel()
	s.autoSync(ctx)
}

// watchChanges syncs the open profile shortly after local changes, and
// when another profile is opened.
func (s *SyncService) watchChanges() {
	for ch := range s.a.changes(context.Background()) {
		switch ch.Kind {
		case "sync", "syncFix", "slots", "renumber", "prune", "environment":
			continue // made by sync itself, or no change
		case "replaced":
			go s.trigger()
			continue
		}
		s.mu.Lock()
		if s.node != nil {
			if s.debounce != nil {
				s.debounce.Stop()
			}
			s.debounce = time.AfterFunc(syncDebounce, s.trigger)
		}
		s.mu.Unlock()
	}
}

// trigger runs an automatic sync of the open profile in the foreground; a
// trigger while one runs runs another after it.
func (s *SyncService) trigger() {
	s.mu.Lock()
	if s.node == nil || !s.foreground || !s.trustedLocked() {
		s.mu.Unlock()
		return
	}
	if s.running {
		s.again = true
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()
	for {
		s.autoSync(context.Background())
		s.mu.Lock()
		again := s.again
		s.again = false
		if !again {
			s.running = false
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
	}
}

// autoSync syncs the open profile with each reachable linked device,
// leaving out links waiting for the first-sync choice. Failures show in
// the status only.
func (s *SyncService) autoSync(ctx context.Context) {
	profile, links, err := s.openLinks()
	if err != nil {
		return
	}
	for _, l := range links {
		if l.NeedsChoice && l.Pending == cfsync.ChoiceNone {
			continue
		}
		if !s.reachable(l.Peer) {
			continue
		}
		_, _ = s.syncLink(ctx, profile, l.ID, cfsync.ChoiceNone, true)
	}
}

// reachable reports whether a paired device can be tried: it was seen on
// the network, or has a last known address.
func (s *SyncService) reachable(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addrLocked(id) != ""
}

func (s *SyncService) addrLocked(id string) string {
	if d, ok := s.seen[id]; ok {
		return d.Addr
	}
	if s.node != nil {
		d, _ := s.node.Device(id)
		return d.Addr
	}
	return ""
}

// SyncNow syncs the open profile with every reachable linked device and
// reports what it received and sent (device-sync spec: Sync now). A link
// whose first sync needs the user's choice asks for it.
func (s *SyncService) SyncNow(ctx context.Context) (SyncSummary, error) {
	s.mu.Lock()
	_, err := s.readyLocked()
	trusted := s.trustedLocked()
	s.mu.Unlock()
	if err != nil {
		return SyncSummary{}, err
	}
	if !trusted {
		return SyncSummary{Error: core.NewError("syncUntrusted", "Trust this network first.").Error()}, nil
	}
	profile, links, err := s.openLinks()
	if err != nil {
		return SyncSummary{}, err
	}
	epoch := ""
	if c, err := s.a.collection(); err == nil {
		epoch = c.Epoch()
	}
	var sum SyncSummary
	for _, l := range links {
		if !s.reachable(l.Peer) {
			continue
		}
		sum.Reachable = true
		// A collection replaced since the last sync (restore, import of a
		// collection package) asks again (device-sync spec).
		choice, ok := cfsync.ChoiceNone, true
		if l.Epoch == "" || l.NeedsChoice || l.Epoch != epoch {
			if choice, ok = s.askFirstSync(ctx, profile, l); !ok {
				continue
			}
		}
		got, err := s.syncLink(ctx, profile, l.ID, choice, false)
		var r *cfsync.Refusal
		if errors.As(err, &r) && r.Code == cfsync.RefuseNeedsChoice && choice == cfsync.ChoiceNone {
			// The other side's collection was replaced.
			if choice, ok = s.askFirstSync(ctx, profile, l); !ok {
				continue
			}
			got, err = s.syncLink(ctx, profile, l.ID, choice, false)
		}
		sum.Received += got.Received
		sum.Sent += got.Sent
		sum.Media += got.Media
		if err != nil && sum.Error == "" {
			sum.Error = err.Error()
		}
	}
	return sum, nil
}

// askFirstSync asks the user how to combine the open profile with the one
// of link l; without an answer, the link waits for one.
func (s *SyncService) askFirstSync(ctx context.Context, profile string, l linkRecord) (cfsync.Choice, bool) {
	answer, err := s.ui.ChooseFirstSync(ctx, s.linkInfo(l))
	choice := cfsync.Choice(answer)
	if err != nil || choice == cfsync.ChoiceNone ||
		choice != cfsync.ChoiceMerge && choice != cfsync.ChoiceReplaceOther && choice != cfsync.ChoiceReplaceThis {
		s.markNeedsChoice(profile, l.ID)
		return cfsync.ChoiceNone, false
	}
	return choice, true
}

func (s *SyncService) markNeedsChoice(profile, linkID string) {
	_ = s.updateLink(profile, linkID, func(r *linkRecord) { r.NeedsChoice = true })
}

// syncLink runs a session of link linkID of profile, started by this
// device. An automatic sync (auto) records no error when the peer cannot
// be reached and was not seen on the network, as it is away, nor when the
// peer refuses because it is starting a session itself.
func (s *SyncService) syncLink(ctx context.Context, profile, linkID string, choice cfsync.Choice, auto bool) (cfsync.Summary, error) {
	rec, err := s.findLink(profile, linkID)
	if err != nil {
		return cfsync.Summary{}, err
	}
	s.mu.Lock()
	n := s.node
	peer, paired := n.Device(rec.Peer)
	addr := s.addrLocked(rec.Peer)
	_, announced := s.seen[rec.Peer]
	if !paired || addr == "" || s.outgoing[rec.Peer] || !s.trustedLocked() {
		s.mu.Unlock()
		return cfsync.Summary{}, nil
	}
	s.outgoing[rec.Peer] = true
	open := profile == s.openProfile()
	if open {
		s.syncing++
	}
	s.notifyLocked()
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.outgoing, rec.Peer)
		if open {
			s.syncing--
			if s.syncing == 0 {
				s.progress = nil
			}
		}
		s.notifyLocked()
		s.mu.Unlock()
	}()

	col, backend, release, err := s.a.acquireCollection(profile)
	if err != nil {
		return cfsync.Summary{}, err
	}
	defer release()
	away := false
	sum, err := func() (cfsync.Summary, error) {
		c, err := n.Connect(ctx, peer, addr)
		if err != nil {
			away = !announced
			return cfsync.Summary{}, connectError(peer.Name, addr, err)
		}
		if !s.track(c) {
			c.Close()
			return cfsync.Summary{}, core.NewError("syncUntrusted", "Trust this network first.")
		}
		defer s.untrack(c)
		defer c.Close()
		link := rec.Link
		return cfsync.Initiate(ctx, c, col, &link, choice, s.sessionOptions(profile, col, backend, open))
	}()
	var refusal *cfsync.Refusal
	if auto && (away || errors.As(err, &refusal) && refusal.Code == cfsync.RefuseBusy) {
		return sum, err
	}
	s.finished(profile, linkID, col, err)
	return sum, err
}

// connectError explains why the device named name could not be reached at
// addr. A device that does not answer is most likely behind a firewall
// that drops sync connections.
func connectError(name, addr string, err error) error {
	if syncnet.IsTimeout(err) {
		return core.NewError("syncConnectTimeout", "%s did not answer at %s. If it is on, a firewall on it may block sync: allow incoming TCP connections on port %s there.", name, addr, syncnet.Port(addr))
	}
	return core.NewError("syncConnectFailed", "%s could not be reached: %v", name, err)
}

// openProfile returns the open profile's ID.
func (s *SyncService) openProfile() string {
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	if !s.a.opened {
		return ""
	}
	return s.a.profile
}

// sessionOptions configures a session on profile's collection.
func (s *SyncService) sessionOptions(profile string, col *core.Collection, backend storage.Backend, open bool) cfsync.Options {
	return cfsync.Options{
		Device: s.node.ID(),
		Name:   s.node.Name(),
		Now:    s.a.now,
		Backup: func() error {
			_, err := backup.WriteAuto(backend.Backups(), col, appVersion, s.a.now(), limitsOf(col.Settings().Backup), yield.Yield)
			return err
		},
		SaveLink: func(l cfsync.Link) error {
			return s.updateLink(profile, l.ID, func(r *linkRecord) { r.Link = l })
		},
		Progress: func(p cfsync.Progress) {
			if open {
				s.mu.Lock()
				s.progress = &p
				s.notifyLocked()
				s.mu.Unlock()
			}
		},
	}
}

// finished records the end of a session: the error in the link and the
// status, and tombstones every link of the profile has received are
// dropped.
func (s *SyncService) finished(profile, linkID string, col *core.Collection, err error) {
	msg := ""
	var refusal *cfsync.Refusal
	if err != nil {
		msg = err.Error()
		if errors.As(err, &refusal) && refusal.Code == cfsync.RefuseUnknownLink {
			// The other side removed the link.
			_ = s.removeLink(profile, linkID)
		}
	}
	_ = s.updateLink(profile, linkID, func(r *linkRecord) {
		r.Error = msg
		if refusal != nil && refusal.Code == cfsync.RefuseNeedsChoice {
			r.NeedsChoice = true
		}
	})
	if err == nil {
		s.prune(profile, col)
	}
	if profile == s.openProfile() {
		s.mu.Lock()
		s.lastErr = msg
		s.notifyLocked()
		s.mu.Unlock()
	}
}

// prune drops the tombstones of profile's collection that every link has
// received (decision 10).
func (s *SyncService) prune(profile string, col *core.Collection) {
	links, err := s.links(profile)
	if err != nil || len(links) == 0 {
		return
	}
	acked := links[0].PeerReceived
	for _, l := range links[1:] {
		acked = min(acked, l.PeerReceived)
	}
	_, _ = col.PruneTombstones(acked, core.TombstoneRetention)
}

// serve handles a connection of a paired device.
func (s *SyncService) serve(peer syncnet.Device, first *cfsync.Message, c syncnet.SessionConn) {
	defer c.Close()
	switch first.Type {
	case "hello":
		s.respond(peer, c)
		return
	case syncnet.MsgListProfiles:
		list, err := (&AppService{a: s.a}).Profiles()
		if err != nil {
			_ = c.Send(&cfsync.Message{Type: syncnet.MsgFailed, Error: err.Error()})
			return
		}
		data, _ := json.Marshal(list.Profiles)
		_ = c.Send(&cfsync.Message{Type: syncnet.MsgProfiles, Data: data})
	case syncnet.MsgCreateLink:
		var req syncnet.CreateLinkRequest
		if err := json.Unmarshal(first.Data, &req); err != nil {
			return
		}
		p, err := s.acceptLink(peer, req)
		if err != nil {
			_ = c.Send(&cfsync.Message{Type: syncnet.MsgFailed, Error: err.Error()})
			return
		}
		data, _ := json.Marshal(p)
		_ = c.Send(&cfsync.Message{Type: syncnet.MsgLinked, Data: data})
	case syncnet.MsgUnlink:
		for _, p := range s.profileIDs() {
			if r, err := s.findLink(p, first.Name); err == nil && r.Peer == peer.ID {
				_ = s.removeLink(p, first.Name)
			}
		}
	}
}

// respond runs a session a paired device started. When both devices start
// one at the same time, the one the smaller device ID started goes on.
func (s *SyncService) respond(peer syncnet.Device, c syncnet.SessionConn) {
	s.mu.Lock()
	if s.outgoing[peer.ID] && s.node.ID() < peer.ID {
		s.mu.Unlock()
		_ = c.Send(&cfsync.Message{Type: "refuse", Refusal: &cfsync.Refusal{Code: cfsync.RefuseBusy, Device: s.node.Name()}})
		return
	}
	s.mu.Unlock()
	var profile, link string
	var col *core.Collection
	var backend storage.Backend
	lookup := func(peerID, linkID string) (cfsync.Collection, *cfsync.Link, func(), error) {
		for _, p := range s.profileIDs() {
			r, err := s.findLink(p, linkID)
			if err != nil || r.Peer != peerID {
				continue
			}
			cl, b, release, err := s.a.acquireCollection(p)
			if err != nil {
				return nil, nil, nil, err
			}
			profile, link, col, backend = p, linkID, cl, b
			if s.serving != nil {
				s.serving(p)
			}
			l := r.Link
			return cl, &l, release, nil
		}
		return nil, nil, nil, nil
	}
	opts := cfsync.Options{Device: s.node.ID(), Name: s.node.Name(), Now: s.a.now}
	opts.Backup = func() error {
		_, err := backup.WriteAuto(backend.Backups(), col, appVersion, s.a.now(), limitsOf(col.Settings().Backup), yield.Yield)
		return err
	}
	opts.SaveLink = func(l cfsync.Link) error {
		return s.updateLink(profile, l.ID, func(r *linkRecord) { r.Link = l })
	}
	_, err := cfsync.Respond(context.Background(), c, peer.ID, lookup, opts)
	if col == nil {
		return
	}
	if err == nil {
		// The link works, even if this device could not reach the other
		// one before (a firewall there lets only this direction through).
		_ = s.updateLink(profile, link, func(r *linkRecord) { r.Error = "" })
		s.prune(profile, col)
	}
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	if profile == s.openProfile() {
		s.mu.Lock()
		s.lastErr = msg
		s.notifyLocked()
		s.mu.Unlock()
	}
}

// acceptLink creates this side of a link a paired device asks for.
func (s *SyncService) acceptLink(peer syncnet.Device, req syncnet.CreateLinkRequest) (Profile, error) {
	app := &AppService{a: s.a}
	var p Profile
	if req.Profile == "" {
		list, err := app.Profiles()
		if err != nil {
			return Profile{}, err
		}
		name := req.NewName
		for i := 2; slices.ContainsFunc(list.Profiles, func(x Profile) bool { return strings.EqualFold(x.Name, name) }); i++ {
			name = fmt.Sprintf("%s (%d)", req.NewName, i)
		}
		if p, err = app.AddProfile(name); err != nil {
			return Profile{}, err
		}
	} else {
		list, err := app.Profiles()
		if err != nil {
			return Profile{}, err
		}
		i := slices.IndexFunc(list.Profiles, func(x Profile) bool { return x.ID == req.Profile })
		if i < 0 {
			return Profile{}, errProfileGone()
		}
		p = list.Profiles[i]
	}
	links, err := s.links(p.ID)
	if err != nil {
		return Profile{}, err
	}
	if slices.ContainsFunc(links, func(l linkRecord) bool { return l.Peer == peer.ID }) {
		return Profile{}, core.NewError("syncAlreadyLinked", "Profile %q is already linked with a profile on %s.", p.Name, peer.Name)
	}
	rec := linkRecord{Link: cfsync.NewLink(req.LinkID, peer.ID, req.FromProfile), RemoteName: req.FromName}
	if err := s.saveLinks(p.ID, append(links, rec)); err != nil {
		return Profile{}, err
	}
	s.notify()
	return p, nil
}

// request sends a request to a paired device and returns its answer.
func (s *SyncService) request(ctx context.Context, deviceID string, m *cfsync.Message, want string) (*cfsync.Message, error) {
	s.mu.Lock()
	n, err := s.readyLocked()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	peer, ok := n.Device(deviceID)
	addr := s.addrLocked(deviceID)
	trusted := s.trustedLocked()
	s.mu.Unlock()
	if !ok {
		return nil, core.NewError("syncNotPaired", "That device is not paired.")
	}
	if !trusted {
		return nil, core.NewError("syncUntrusted", "Trust this network first.")
	}
	if addr == "" {
		return nil, core.NewError("syncNotReachable", "%s is not reachable on this network.", peer.Name)
	}
	c, err := n.Connect(ctx, peer, addr)
	if err != nil {
		return nil, connectError(peer.Name, addr, err)
	}
	if !s.track(c) {
		c.Close()
		return nil, core.NewError("syncUntrusted", "Trust this network first.")
	}
	defer s.untrack(c)
	defer c.Close()
	if err := c.Send(m); err != nil || want == "" {
		return nil, err
	}
	reply, err := c.Recv()
	if err != nil {
		return nil, err
	}
	if reply.Type == syncnet.MsgFailed {
		return nil, errors.New(reply.Error)
	}
	if reply.Type != want {
		return nil, fmt.Errorf("unexpected answer %q", reply.Type)
	}
	return reply, nil
}

// RemoteProfiles lists the profiles of a paired device, for linking.
func (s *SyncService) RemoteProfiles(ctx context.Context, deviceID string) ([]Profile, error) {
	reply, err := s.request(ctx, deviceID, &cfsync.Message{Type: syncnet.MsgListProfiles}, syncnet.MsgProfiles)
	if err != nil {
		return nil, err
	}
	var out []Profile
	if err := json.Unmarshal(reply.Data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Link links the open profile with profile remoteProfile of a paired
// device, or with a new profile there named like the open one when
// remoteProfile is empty (device-sync spec: Profile links). The first sync
// then asks how to combine them.
func (s *SyncService) Link(ctx context.Context, deviceID, remoteProfile string) (LinkInfo, error) {
	profile := s.openProfile()
	if profile == "" {
		return LinkInfo{}, errNotOpened
	}
	links, err := s.links(profile)
	if err != nil {
		return LinkInfo{}, err
	}
	if slices.ContainsFunc(links, func(l linkRecord) bool { return l.Peer == deviceID }) {
		return LinkInfo{}, core.NewError("syncAlreadyLinkedHere", "This profile is already linked with a profile on that device.")
	}
	name := s.profileName(profile)
	b := make([]byte, 12)
	rand.Read(b)
	req := syncnet.CreateLinkRequest{LinkID: hex.EncodeToString(b), Profile: remoteProfile, NewName: name, FromProfile: profile, FromName: name}
	data, _ := json.Marshal(req)
	reply, err := s.request(ctx, deviceID, &cfsync.Message{Type: syncnet.MsgCreateLink, Data: data}, syncnet.MsgLinked)
	if err != nil {
		return LinkInfo{}, err
	}
	var remote Profile
	if err := json.Unmarshal(reply.Data, &remote); err != nil {
		return LinkInfo{}, err
	}
	rec := linkRecord{Link: cfsync.NewLink(req.LinkID, deviceID, remote.ID), RemoteName: remote.Name}
	rec.NeedsChoice = true
	if err := s.saveLinks(profile, append(links, rec)); err != nil {
		return LinkInfo{}, err
	}
	s.notify()
	return s.linkInfo(rec), nil
}

// Unlink removes a link of the open profile; the other device removes its
// side too, now if it is reachable, otherwise at its next sync.
func (s *SyncService) Unlink(ctx context.Context, linkID string) error {
	profile := s.openProfile()
	rec, err := s.findLink(profile, linkID)
	if err != nil {
		return err
	}
	if err := s.removeLink(profile, linkID); err != nil {
		return err
	}
	_, _ = s.request(ctx, rec.Peer, &cfsync.Message{Type: syncnet.MsgUnlink, Name: linkID}, "")
	return nil
}

func (s *SyncService) profileName(id string) string {
	list, err := (&AppService{a: s.a}).Profiles()
	if err != nil {
		return ""
	}
	for _, p := range list.Profiles {
		if p.ID == id {
			return p.Name
		}
	}
	return ""
}

func (s *SyncService) profileIDs() []string {
	list, err := (&AppService{a: s.a}).Profiles()
	if err != nil {
		return nil
	}
	out := make([]string, len(list.Profiles))
	for i, p := range list.Profiles {
		out[i] = p.ID
	}
	return out
}

// links returns the links of profile.
func (s *SyncService) links(profile string) ([]linkRecord, error) {
	s.linksMu.Lock()
	defer s.linksMu.Unlock()
	return s.linksLocked(profile)
}

func (s *SyncService) linksLocked(profile string) ([]linkRecord, error) {
	data, err := s.readLinks(s.rootDir(), profile)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []linkRecord
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, core.NewError("syncLinksUnreadable", "The links of this profile could not be read: %v", err)
	}
	return out, nil
}

func (s *SyncService) rootDir() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.root
}

func (s *SyncService) saveLinks(profile string, links []linkRecord) error {
	s.linksMu.Lock()
	defer s.linksMu.Unlock()
	return s.saveLinksLocked(profile, links)
}

func (s *SyncService) saveLinksLocked(profile string, links []linkRecord) error {
	if links == nil {
		links = []linkRecord{}
	}
	data, err := json.Marshal(links)
	if err != nil {
		return err
	}
	return s.writeLinks(s.rootDir(), profile, data)
}

// updateLink changes a link of profile and saves it.
func (s *SyncService) updateLink(profile, linkID string, edit func(*linkRecord)) error {
	s.linksMu.Lock()
	links, err := s.linksLocked(profile)
	if err != nil {
		s.linksMu.Unlock()
		return err
	}
	i := slices.IndexFunc(links, func(l linkRecord) bool { return l.ID == linkID })
	if i < 0 {
		s.linksMu.Unlock()
		return nil
	}
	edit(&links[i])
	err = s.saveLinksLocked(profile, links)
	s.linksMu.Unlock()
	s.notify()
	return err
}

func (s *SyncService) findLink(profile, linkID string) (linkRecord, error) {
	links, err := s.links(profile)
	if err != nil {
		return linkRecord{}, err
	}
	for _, l := range links {
		if l.ID == linkID {
			return l, nil
		}
	}
	return linkRecord{}, core.NewError("syncLinkGone", "That link no longer exists.")
}

func (s *SyncService) removeLink(profile, linkID string) error {
	s.linksMu.Lock()
	links, err := s.linksLocked(profile)
	if err == nil {
		err = s.saveLinksLocked(profile, slices.DeleteFunc(links, func(l linkRecord) bool { return l.ID == linkID }))
	}
	s.linksMu.Unlock()
	s.notify()
	return err
}

// removeLinksOf removes every link to an unpaired device.
func (s *SyncService) removeLinksOf(deviceID string) {
	for _, p := range s.profileIDs() {
		s.linksMu.Lock()
		links, err := s.linksLocked(p)
		if err == nil && slices.ContainsFunc(links, func(l linkRecord) bool { return l.Peer == deviceID }) {
			_ = s.saveLinksLocked(p, slices.DeleteFunc(links, func(l linkRecord) bool { return l.Peer == deviceID }))
		}
		s.linksMu.Unlock()
	}
	s.notify()
}

// openLinks returns the open profile and its links.
func (s *SyncService) openLinks() (string, []linkRecord, error) {
	profile := s.openProfile()
	if profile == "" {
		return "", nil, errNotOpened
	}
	links, err := s.links(profile)
	return profile, links, err
}

func (s *SyncService) linkInfo(l linkRecord) LinkInfo {
	info := LinkInfo{ID: l.ID, Device: l.Peer, RemoteProfile: l.RemoteProfile, RemoteName: l.RemoteName,
		LastSynced: l.LastSynced, NeedsChoice: l.NeedsChoice || l.Epoch == "", Error: l.Error}
	s.mu.Lock()
	if s.node != nil {
		d, _ := s.node.Device(l.Peer)
		info.DeviceName = d.Name
	}
	s.mu.Unlock()
	return info
}

// Status returns the state of sync.
func (s *SyncService) Status() SyncStatus {
	links := s.openLinksQuiet()
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.statusLocked()
	s.addLinksLocked(&st, links)
	return st
}

func (s *SyncService) openLinksQuiet() []linkRecord {
	if !s.supported {
		return nil
	}
	_, links, _ := s.openLinks()
	return links
}

func (s *SyncService) statusLocked() SyncStatus {
	st := SyncStatus{Supported: s.supported, Trusted: []TrustedNetwork{}, Paired: []PairedDeviceInfo{},
		Discovered: []DiscoveredDevice{}, Links: []LinkInfo{}, Error: s.lastErr}
	st.Network = NetworkInfo{Kind: s.network.Kind, WiFi: s.network.WiFi, Subnet: s.network.Subnet, ID: s.network.id(),
		Trusted: s.trustedLocked()}
	if s.node == nil {
		return st
	}
	st.Ready, st.DeviceID, st.DeviceName = true, s.node.ID(), s.node.Name()
	st.Trusted = append(st.Trusted, s.state.Networks...)
	if s.listener != nil {
		st.Listening, st.Port = true, s.listener.Port()
		if _, addr := syncnet.LocalNetwork(); addr != "" {
			st.Address = addr
		}
	}
	st.Pairing = st.Listening && s.node.Pairing()
	paired := map[string]bool{}
	for _, p := range s.node.Devices() {
		_, seen := s.seen[p.ID]
		paired[p.ID] = true
		st.Paired = append(st.Paired, PairedDeviceInfo{ID: p.ID, Name: p.Name, Reachable: seen})
	}
	for id, d := range s.seen {
		if paired[id] || !d.Pairing || id == s.node.ID() {
			continue
		}
		host, port, _ := net.SplitHostPort(d.Addr)
		pn, _ := strconv.Atoi(port)
		st.Discovered = append(st.Discovered, DiscoveredDevice{ID: id, Name: d.Name, Host: host, Port: pn})
	}
	slices.SortFunc(st.Discovered, func(a, b DiscoveredDevice) int { return strings.Compare(a.Name, b.Name) })
	st.Syncing = s.syncing > 0
	st.Progress = s.progress
	return st
}

// addLinksLocked adds the open profile's links to the status.
func (s *SyncService) addLinksLocked(st *SyncStatus, links []linkRecord) {
	if s.node == nil {
		return
	}
	for _, l := range links {
		d, _ := s.node.Device(l.Peer)
		name := d.Name
		st.Links = append(st.Links, LinkInfo{ID: l.ID, Device: l.Peer, DeviceName: name, RemoteProfile: l.RemoteProfile,
			RemoteName: l.RemoteName, LastSynced: l.LastSynced, NeedsChoice: l.NeedsChoice || l.Epoch == "", Error: l.Error})
		if l.LastSynced.After(st.LastSynced) {
			st.LastSynced = l.LastSynced
		}
		for i := range st.Paired {
			if st.Paired[i].ID == l.Peer {
				st.Paired[i].LastSynced = l.LastSynced
			}
		}
	}
}

func (s *SyncService) notify() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifyLocked()
}

func (s *SyncService) notifyLocked() {
	close(s.changed)
	s.changed = make(chan struct{})
}

// StatusChanges streams the status, first the current one and then again
// whenever it changes, until ctx ends.
func (s *SyncService) StatusChanges(ctx context.Context) <-chan SyncStatus {
	out := make(chan SyncStatus)
	go func() {
		defer close(out)
		for {
			s.mu.Lock()
			changed := s.changed
			s.mu.Unlock()
			st := s.Status()
			select {
			case out <- st:
			case <-ctx.Done():
				return
			}
			select {
			case <-changed:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// LinkInfo describes a link of the open profile, for the first-sync
// question and the Sync section.
type LinkInfo struct {
	ID            string `json:"id"`
	Device        string `json:"device"`
	DeviceName    string `json:"deviceName"`
	RemoteProfile string `json:"remoteProfile"`
	// RemoteName is the name of the linked profile on the other device.
	RemoteName string    `json:"remoteName"`
	LastSynced time.Time `json:"lastSynced,omitzero"`
	// NeedsChoice is set while the first-sync question waits for an
	// answer; Error is why the last sync failed (a tagged text).
	NeedsChoice bool   `json:"needsChoice"`
	Error       string `json:"error,omitempty"`
}

// SyncSummary is the result of Sync now: how many changes were received
// and sent and media files fetched, or why it failed (a tagged text).
type SyncSummary struct {
	Received int    `json:"received"`
	Sent     int    `json:"sent"`
	Media    int    `json:"media"`
	Error    string `json:"error,omitempty"`
	// Reachable is false when no linked device was reachable.
	Reachable bool `json:"reachable"`
}
