//go:build !js

package main

import (
	"context"
	"net"
	"strconv"
	gosync "sync"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/internal/usermsg"
	"cardflip/golang/storage"
)

// fakeSyncUI answers like a user: it accepts pairing requests if accept is
// set, and confirms the code if confirm says so.
type fakeSyncUI struct {
	mu      gosync.Mutex
	accept  bool
	confirm func(ctx context.Context, code string) bool
}

func (u *fakeSyncUI) PairingRequest(ctx context.Context, name string) (bool, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.accept, nil
}

func (u *fakeSyncUI) ConfirmPairing(ctx context.Context, name, code string) (bool, error) {
	return u.confirm(ctx, code), nil
}

func (u *fakeSyncUI) PairingFinished(name string, paired bool) {}

func waitFor(cond func() bool) bool {
	for range 300 {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// syncPeer is an app with its sync service, on its own in-memory data.
type syncPeer struct {
	*services
	sync *SyncService
	ui   *fakeSyncUI
	mu   gosync.Mutex
	file []byte
	// links are the stored sync-links.json files by profile.
	links map[string][]byte
	// choose answers the first-sync question.
	choose func(ctx context.Context, l LinkInfo) (string, error)
}

// syncTestUI is fakeSyncUI with a configurable first-sync answer.
type syncTestUI struct {
	*fakeSyncUI
	p *syncPeer
}

func (u syncTestUI) ChooseFirstSync(ctx context.Context, l LinkInfo) (string, error) {
	if u.p.choose == nil {
		return "merge", nil
	}
	return u.p.choose(ctx, l)
}

// newSyncPeer opens an app named name whose sync is set up on the trusted
// Wi-Fi "Home".
func newSyncPeer(t *testing.T, name string, trust bool) *syncPeer {
	t.Helper()
	p := &syncPeer{services: newServices(t, storage.NewMemoryStore()), links: map[string][]byte{}}
	p.ui = &fakeSyncUI{accept: true, confirm: func(context.Context, string) bool { return true }}
	p.sync = newSyncService(p.app.a, syncTestUI{p.ui, p})
	p.sync.read = func(string) ([]byte, error) {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.file == nil {
			return nil, storage.ErrNotFound
		}
		return p.file, nil
	}
	p.sync.write = func(_ string, data []byte) error {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.file = data
		return nil
	}
	p.sync.readLinks = func(_, profile string) ([]byte, error) {
		p.mu.Lock()
		defer p.mu.Unlock()
		if d, ok := p.links[profile]; ok {
			return d, nil
		}
		return nil, storage.ErrNotFound
	}
	p.sync.writeLinks = func(_, profile string, data []byte) error {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.links[profile] = data
		return nil
	}
	if _, err := p.app.Open("/data"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.sync.Init(name); err != nil {
		t.Fatal(err)
	}
	p.sync.SetNetwork(NetworkWiFi, "Home", "192.168.1.0/24")
	if trust {
		check(t, p.sync.TrustNetwork())
	}
	t.Cleanup(func() {
		p.sync.mu.Lock()
		defer p.sync.mu.Unlock()
		p.sync.state.Networks = nil
		p.sync.applyNetworkLocked()
	})
	return p
}

func (p *syncPeer) id() string { return p.sync.Status().DeviceID }
func (p *syncPeer) port() int  { return p.sync.Status().Port }

// see makes each peer see the other on the network.
func see(a, b *syncPeer) {
	a.sync.Seen(b.id(), b.sync.Status().DeviceName, "127.0.0.1", b.port(), false)
	b.sync.Seen(a.id(), a.sync.Status().DeviceName, "127.0.0.1", a.port(), false)
}

// pairPeers pairs a and b and lets them see each other.
func pairPeers(t *testing.T, a, b *syncPeer) {
	t.Helper()
	check(t, b.sync.StartPairing())
	check(t, a.sync.PairWith(context.Background(), "127.0.0.1", b.port()))
	b.sync.StopPairing()
	if !waitFor(func() bool { return len(b.sync.Status().Paired) == 1 }) || len(a.sync.Status().Paired) != 1 {
		t.Fatal("not paired")
	}
	see(a, b)
}

// linkedPeers returns two paired peers whose open profiles are linked and
// merged; b has a note of its own, a has one too.
func linkedPeers(t *testing.T) (a, b *syncPeer) {
	t.Helper()
	a, b = newSyncPeer(t, "Laptop", true), newSyncPeer(t, "Phone", true)
	*b.clock = (*a.clock).Add(time.Second)
	pairPeers(t, a, b)
	basicNote(t, a.services, "from the laptop", "")
	basicNote(t, b.services, "from the phone", "")
	if _, err := a.sync.Link(context.Background(), b.id(), storage.DefaultProfile); err != nil {
		t.Fatal(err)
	}
	sum, err := a.sync.SyncNow(context.Background())
	if err != nil || sum.Error != "" || !sum.Reachable {
		t.Fatalf("first sync: %+v, %v", sum, err)
	}
	return a, b
}

func noteCount(t *testing.T, p *syncPeer) int {
	t.Helper()
	c, err := p.app.a.collection()
	if err != nil {
		t.Fatal(err)
	}
	return len(c.AllNoteIDs())
}

// Nothing listens before the network is trusted, and cellular networks
// cannot be trusted.
func TestSyncListensOnlyOnTrustedNetworks(t *testing.T) {
	p := newSyncPeer(t, "Laptop", false)
	if st := p.sync.Status(); st.Listening || st.Port != 0 || st.Network.Trusted || !st.Ready {
		t.Fatalf("status before trusting: %+v", st)
	}
	p.sync.SetNetwork(NetworkCellular, "", "")
	if err := p.sync.TrustNetwork(); err == nil {
		t.Fatal("cellular network trusted")
	}
	if p.sync.Status().Listening {
		t.Fatal("listening on a cellular network")
	}
	p.sync.SetNetwork(NetworkWiFi, "Home", "192.168.1.0/24")
	check(t, p.sync.TrustNetwork())
	st := p.sync.Status()
	if !st.Listening || st.Port == 0 || !st.Network.Trusted || len(st.Trusted) != 1 || st.Trusted[0].WiFi != "Home" {
		t.Fatalf("status after trusting: %+v", st)
	}
	c, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(st.Port)))
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	// Another Wi-Fi network is not trusted.
	p.sync.SetNetwork(NetworkWiFi, "Café", "10.0.0.0/24")
	if p.sync.Status().Listening {
		t.Fatal("listening on an untrusted network")
	}
	// Back home.
	p.sync.SetNetwork(NetworkWiFi, "Home", "192.168.1.0/24")
	if !p.sync.Status().Listening {
		t.Fatal("not listening back home")
	}
}

// Removing trust in the current network stops sync at once and closes open
// connections.
func TestSyncUntrustClosesConnections(t *testing.T) {
	a, b := newSyncPeer(t, "Laptop", true), newSyncPeer(t, "Phone", true)
	pairPeers(t, a, b)
	peer := a.sync.node.Devices()[0]
	c, err := a.sync.node.Connect(context.Background(), peer, net.JoinHostPort("127.0.0.1", strconv.Itoa(b.port())))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	port := b.port()
	check(t, b.sync.UntrustNetwork(b.sync.Status().Network.ID))
	if b.sync.Status().Listening {
		t.Fatal("still listening")
	}
	done := make(chan error, 1)
	go func() {
		_, err := c.Recv()
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("connection still open")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("connection not closed")
	}
	if c, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port))); err == nil {
		c.Close()
		t.Fatal("still accepting connections")
	}
}

// The first sync asks the user and changes nothing before the answer; then
// both profiles have both notes.
func TestSyncFirstSyncAsks(t *testing.T) {
	a, b := newSyncPeer(t, "Laptop", true), newSyncPeer(t, "Phone", true)
	pairPeers(t, a, b)
	basicNote(t, a.services, "from the laptop", "")
	if _, err := a.sync.Link(context.Background(), b.id(), storage.DefaultProfile); err != nil {
		t.Fatal(err)
	}
	if links := a.sync.Status().Links; len(links) != 1 || !links[0].NeedsChoice || links[0].DeviceName != "Phone" {
		t.Fatalf("links %+v", links)
	}
	asked := make(chan LinkInfo, 1)
	answer := make(chan string)
	a.choose = func(ctx context.Context, l LinkInfo) (string, error) {
		asked <- l
		return <-answer, nil
	}
	revB := must(b.app.a.collection()).Revision()
	type result struct {
		sum SyncSummary
		err error
	}
	done := make(chan result, 1)
	go func() {
		sum, err := a.sync.SyncNow(context.Background())
		done <- result{sum, err}
	}()
	l := <-asked
	if l.DeviceName != "Phone" || l.RemoteName != "Default" {
		t.Fatalf("asked about %+v", l)
	}
	time.Sleep(50 * time.Millisecond)
	if must(b.app.a.collection()).Revision() != revB || noteCount(t, b) != 0 {
		t.Fatal("the phone changed before the answer")
	}
	answer <- "merge"
	r := <-done
	if r.err != nil || r.sum.Error != "" || r.sum.Sent == 0 {
		t.Fatalf("SyncNow = %+v, %v", r.sum, r.err)
	}
	if noteCount(t, b) != 1 {
		t.Fatal("the note did not arrive")
	}
	if links := a.sync.Status().Links; links[0].NeedsChoice || links[0].LastSynced.IsZero() {
		t.Fatalf("link after the first sync: %+v", links[0])
	}
	// Deciding later leaves the link waiting.
	c, d := newSyncPeer(t, "Tablet", true), newSyncPeer(t, "Desk", true)
	pairPeers(t, c, d)
	if _, err := c.sync.Link(context.Background(), d.id(), ""); err != nil {
		t.Fatal(err)
	}
	c.choose = func(context.Context, LinkInfo) (string, error) { return "", nil }
	if sum, err := c.sync.SyncNow(context.Background()); err != nil || sum.Sent+sum.Received != 0 {
		t.Fatalf("SyncNow = %+v, %v", sum, err)
	}
	if links := c.sync.Status().Links; !links[0].NeedsChoice {
		t.Fatal("link no longer waits for the choice")
	}
}

// A change syncs to the linked device within the debounce.
func TestSyncAfterChange(t *testing.T) {
	old := syncDebounce
	syncDebounce = 100 * time.Millisecond
	defer func() { syncDebounce = old }()
	a, b := linkedPeers(t)
	if noteCount(t, a) != 2 || noteCount(t, b) != 2 {
		t.Fatalf("after the first sync: %d and %d notes", noteCount(t, a), noteCount(t, b))
	}
	basicNote(t, a.services, "later", "")
	if !waitFor(func() bool { return noteCount(t, b) == 3 }) {
		t.Fatal("the change did not sync")
	}
	// And the other way.
	basicNote(t, b.services, "from the phone, later", "")
	if !waitFor(func() bool { return noteCount(t, a) == 4 }) {
		t.Fatal("the phone's change did not sync")
	}
	if st := a.sync.Status(); st.LastSynced.IsZero() || st.Error != "" {
		t.Fatalf("status %+v", st)
	}
}

// Sync now without a reachable linked device says so.
func TestSyncNowNothingReachable(t *testing.T) {
	a, b := linkedPeers(t)
	a.sync.Lost(b.id())
	a.sync.node.ForgetAddr()
	sum, err := a.sync.SyncNow(context.Background())
	if err != nil || sum.Reachable {
		t.Fatalf("SyncNow = %+v, %v", sum, err)
	}
}

// A device that cannot reach the other, as when a firewall there drops its
// connections, shows the error only until the other device syncs with it.
func TestSyncByPeerClearsLinkError(t *testing.T) {
	a, b := linkedPeers(t)
	link := b.sync.Status().Links[0].ID
	check(t, b.sync.updateLink(storage.DefaultProfile, link, func(r *linkRecord) { r.Error = "no answer" }))
	if b.sync.Status().Links[0].Error == "" {
		t.Fatal("the error is not shown")
	}
	if sum, err := a.sync.SyncNow(context.Background()); err != nil || sum.Error != "" {
		t.Fatalf("SyncNow = %+v, %v", sum, err)
	}
	if !waitFor(func() bool { return b.sync.Status().Links[0].Error == "" }) {
		t.Fatalf("the error stayed: %+v", b.sync.Status().Links[0])
	}
}

// An automatic sync with a device that is away, known only by its last
// address, fails without an error; one that is seen on the network but
// cannot be reached shows the error, and so does Sync now.
func TestAutoSyncDeviceAway(t *testing.T) {
	a, b := linkedPeers(t)
	if sum, err := b.sync.SyncNow(context.Background()); err != nil || sum.Error != "" {
		t.Fatalf("SyncNow = %+v, %v", sum, err) // b learns a's address
	}
	port := a.port()
	check(t, a.sync.UntrustNetwork("wifi:Home"))
	b.sync.Lost(a.id())
	b.sync.autoSync(context.Background())
	if st := b.sync.Status(); st.Error != "" || st.Links[0].Error != "" {
		t.Fatalf("an away device is an error: %q, %q", st.Error, st.Links[0].Error)
	}
	sum, err := b.sync.SyncNow(context.Background())
	if err != nil || usermsg.Code(sum.Error) != "syncConnectFailed" {
		t.Fatalf("SyncNow = %+v, %v", sum, err)
	}
	check(t, b.sync.updateLink(storage.DefaultProfile, b.sync.Status().Links[0].ID, func(r *linkRecord) { r.Error = "" }))
	b.sync.Seen(a.id(), "Laptop", "127.0.0.1", port, false) // syncs
	if !waitFor(func() bool { return usermsg.Code(b.sync.Status().Links[0].Error) == "syncConnectFailed" }) {
		t.Fatalf("no error for a device seen: %+v", b.sync.Status().Links[0])
	}
}

// A peer syncs a profile that is not open; the user switching to it
// meanwhile gets the same collection.
func TestSyncServesProfileNotOpen(t *testing.T) {
	a, b := newSyncPeer(t, "Laptop", true), newSyncPeer(t, "Phone", true)
	pairPeers(t, a, b)
	work := must(b.app.AddProfile("Work"))
	basicNote(t, a.services, "for work", "")
	if _, err := a.sync.Link(context.Background(), b.id(), work.ID); err != nil {
		t.Fatal(err)
	}
	switched := make(chan *core.Collection, 1)
	b.sync.serving = func(profile string) {
		if profile != work.ID {
			return
		}
		b.sync.serving = nil
		// The user opens the profile while the laptop syncs it.
		if _, err := b.app.SwitchProfile(work.ID); err != nil {
			t.Error(err)
		}
		switched <- must(b.app.a.collection())
	}
	sum, err := a.sync.SyncNow(context.Background())
	if err != nil || sum.Error != "" {
		t.Fatalf("SyncNow = %+v, %v", sum, err)
	}
	col := <-switched
	// Once the phone's session ended, only the open profile uses it.
	shared := func() *sharedCollection {
		b.app.a.mu.Lock()
		defer b.app.a.mu.Unlock()
		sc := *b.app.a.shared[work.ID]
		return &sc
	}
	if !waitFor(func() bool { return shared().refs == 1 }) || shared().col != col {
		t.Fatalf("shared collection %+v, open %p", shared(), col)
	}
	if noteCount(t, b) != 1 {
		t.Fatal("the note is not in the open profile")
	}
	// The collection is usable and saved.
	basicNote(t, b.services, "on the phone", "")
	if _, err := b.app.SwitchProfile(storage.DefaultProfile); err != nil {
		t.Fatal(err)
	}
	if _, err := b.app.SwitchProfile(work.ID); err != nil {
		t.Fatal(err)
	}
	if noteCount(t, b) != 2 {
		t.Fatalf("%d notes after reopening", noteCount(t, b))
	}
}

// Unpairing removes the links to the device on both sides.
func TestSyncUnpairRemovesLinks(t *testing.T) {
	a, b := linkedPeers(t)
	if len(b.sync.Status().Links) != 1 {
		t.Fatal("phone has no link")
	}
	check(t, a.sync.Unpair(context.Background(), b.id()))
	if len(a.sync.Status().Links)+len(a.sync.Status().Paired) != 0 {
		t.Fatal("laptop keeps the link or the device")
	}
	if !waitFor(func() bool { st := b.sync.Status(); return len(st.Links)+len(st.Paired) == 0 }) {
		t.Fatal("phone keeps the link or the device")
	}
}

// The device state survives a restart; sync.json holds no collection data.
func TestSyncStateSaved(t *testing.T) {
	a, b := linkedPeers(t)
	st := a.sync.Status()
	// A new service on the same data, as after a restart.
	s2 := newSyncService(a.app.a, a.sync.ui)
	s2.read, s2.write, s2.readLinks, s2.writeLinks = a.sync.read, a.sync.write, a.sync.readLinks, a.sync.writeLinks
	s2.opened("/data")
	got := s2.Status()
	if got.DeviceID != st.DeviceID || got.DeviceName != "Laptop" || len(got.Paired) != 1 || got.Paired[0].ID != b.id() ||
		len(got.Links) != 1 {
		t.Fatalf("after a restart: %+v", got)
	}
	if err := s2.SetDeviceName(""); err == nil {
		t.Fatal("empty device name accepted")
	}
	check(t, s2.SetDeviceName("Kitchen tablet"))
	if s2.Status().DeviceName != "Kitchen tablet" {
		t.Fatal("not renamed")
	}
}
