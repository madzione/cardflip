package syncnet

import (
	"bytes"
	"cmp"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"slices"
	"strings"
	gosync "sync"
	"time"

	"cardflip/golang/core"
	cfsync "cardflip/golang/sync"
)

// UI is how pairing asks the user.
type UI interface {
	// PairingRequest asks whether to accept a pairing request from the
	// device named deviceName.
	PairingRequest(ctx context.Context, deviceName string) (bool, error)
	// ConfirmPairing shows the pairing code and asks whether the device
	// named deviceName shows the same.
	ConfirmPairing(ctx context.Context, deviceName, code string) (bool, error)
	// PairingFinished tells whether pairing with deviceName succeeded.
	PairingFinished(deviceName string, paired bool)
}

// PairingTimeout is how long both users have to confirm the pairing code
// (device-pairing spec: Confirm the pairing code); a variable for tests.
var PairingTimeout = 2 * time.Minute

// PairingWindow is how long "Pair a device" makes a device visible for
// pairing.
const PairingWindow = 5 * time.Minute

// Messages of the connection itself and of pairing (sync design decision
// 9). After the handshake, the accepting side sends msgWelcome, or
// msgUnpaired to a device it unpaired, which then forgets it too.
const (
	msgWelcome       = "welcome"
	msgUnpaired      = "unpaired"
	msgPairRequest   = "pairRequest"
	msgPairAccepted  = "pairAccepted"
	msgPairRefused   = "pairRefused"
	msgPairDeclined  = "pairDeclined"
	msgPairConfirmed = "pairConfirmed"
)

// Device is a device paired with this one: its pinned key, and the address
// it was last reached at.
type Device struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  []byte `json:"key"`
	Addr string `json:"addr,omitempty"`
}

// Revoked is the key of a device this one unpaired, which is told so when
// it next connects.
type Revoked struct {
	ID  string `json:"id"`
	Key []byte `json:"key"`
}

// ErrUnpaired means the other device unpaired this one; it was removed.
var ErrUnpaired = errors.New("the other device unpaired this one")

// SessionConn is a connection a sync session runs on.
type SessionConn interface {
	cfsync.Conn
	Close() error
}

// replayConn is a connection whose first message was read already.
type replayConn struct {
	*Conn
	first *cfsync.Message
}

func (c *replayConn) Recv() (*cfsync.Message, error) {
	if m := c.first; m != nil {
		c.first = nil
		return m, nil
	}
	return c.Conn.Recv()
}

// Node is a device on the network: its identity, its paired and unpaired
// devices and the pairing window. It accepts and makes connections, and
// pairs and unpairs devices.
type Node struct {
	id  *Identity
	ui  UI
	now func() time.Time

	mu        gosync.Mutex
	name      string
	paired    map[string]Device
	revoked   map[string][]byte // keys of unpaired devices, by device ID
	pairUntil time.Time         // the pairing window ends then
	pending   bool              // a pairing request is being handled

	// Save persists the device state after a change; OnPaired and
	// OnUnpaired tell about changes; OnSession serves another connection
	// of a paired device, whose first message is first (it is also the
	// first that c receives); it closes c.
	Save       func() error
	OnPaired   func(Device)
	OnUnpaired func(id string)
	OnSession  func(peer Device, first *cfsync.Message, c SessionConn)
}

// NewNode returns the node of the device with identity id.
func NewNode(id *Identity, name string, ui UI, now func() time.Time, paired []Device, revoked []Revoked) *Node {
	n := &Node{id: id, name: name, ui: ui, now: now, paired: map[string]Device{}, revoked: map[string][]byte{},
		Save: func() error { return nil }, OnPaired: func(Device) {}, OnUnpaired: func(string) {},
		OnSession: func(_ Device, _ *cfsync.Message, c SessionConn) { c.Close() }}
	for _, p := range paired {
		n.paired[p.ID] = p
	}
	for _, r := range revoked {
		n.revoked[r.ID] = r.Key
	}
	return n
}

// Identity returns the node's identity, and ID its device ID.
func (n *Node) Identity() *Identity { return n.id }
func (n *Node) ID() string          { return n.id.id }

// Name returns the device's name.
func (n *Node) Name() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.name
}

// SetName renames the device; paired devices see the name at their next
// connection.
func (n *Node) SetName(name string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.name = name
}

// Devices returns the paired devices, sorted by name.
func (n *Node) Devices() []Device {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]Device, 0, len(n.paired))
	for _, p := range n.paired {
		out = append(out, p)
	}
	slices.SortFunc(out, func(a, b Device) int { return cmp.Or(strings.Compare(a.Name, b.Name), strings.Compare(a.ID, b.ID)) })
	return out
}

// Device returns a paired device.
func (n *Node) Device(id string) (Device, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	p, ok := n.paired[id]
	return p, ok
}

// RevokedKeys returns the keys of the unpaired devices, sorted by ID.
func (n *Node) RevokedKeys() []Revoked {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]Revoked, 0, len(n.revoked))
	for id, key := range n.revoked {
		out = append(out, Revoked{ID: id, Key: key})
	}
	slices.SortFunc(out, func(a, b Revoked) int { return strings.Compare(a.ID, b.ID) })
	return out
}

// Verify decides whether a device may connect (decision 8): paired devices
// may; unpaired ones only to learn that they were; unknown ones only during
// the pairing window, and then only to ask to pair.
func (n *Node) Verify(pub ed25519.PublicKey) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	id := DeviceID(pub)
	if p, ok := n.paired[id]; ok && bytes.Equal(p.Key, pub) {
		return nil
	}
	if k, ok := n.revoked[id]; ok && bytes.Equal(k, pub) {
		return nil
	}
	if n.now().Before(n.pairUntil) {
		return nil
	}
	return errors.New("not a paired device")
}

// OpenPairing makes the device visible for pairing for d (0 closes it).
func (n *Node) OpenPairing(d time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.pairUntil = n.now().Add(d)
}

// Pairing reports whether the device is visible for pairing.
func (n *Node) Pairing() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.now().Before(n.pairUntil)
}

// Handle serves an accepted connection and closes it.
func (n *Node) Handle(c *Conn) {
	pub := c.PeerKey()
	id := DeviceID(pub)
	n.mu.Lock()
	peer, paired := n.paired[id]
	paired = paired && bytes.Equal(peer.Key, pub)
	_, revoked := n.revoked[id]
	n.mu.Unlock()
	if revoked && !paired {
		_ = c.Send(&cfsync.Message{Type: msgUnpaired})
		c.Close()
		return
	}
	if err := c.Send(&cfsync.Message{Type: msgWelcome, Name: n.Name()}); err != nil {
		c.Close()
		return
	}
	if !paired {
		n.acceptPairing(c)
		return
	}
	m, err := c.Recv()
	if err != nil {
		c.Close()
		return
	}
	if m.Type == msgUnpaired {
		c.Close()
		n.forget(id)
		return
	}
	n.OnSession(peer, m, &replayConn{Conn: c, first: m})
}

// dial connects to addr, which must present a key accept allows, and reads
// the welcome. A device that unpaired this one is removed.
func (n *Node) dial(ctx context.Context, addr string, accept func(ed25519.PublicKey) error) (*Conn, string, error) {
	c, err := Dial(ctx, n.id, addr, accept)
	if err != nil {
		return nil, "", err
	}
	m, err := c.Recv()
	if err != nil {
		c.Close()
		return nil, "", err
	}
	switch m.Type {
	case msgWelcome:
		return c, m.Name, nil
	case msgUnpaired:
		c.Close()
		n.forget(c.PeerID())
		return nil, "", ErrUnpaired
	}
	c.Close()
	return nil, "", fmt.Errorf("unexpected message %q", m.Type)
}

// Connect opens a connection to a paired device at addr.
func (n *Node) Connect(ctx context.Context, peer Device, addr string) (*Conn, error) {
	c, name, err := n.dial(ctx, addr, func(pub ed25519.PublicKey) error {
		if !bytes.Equal(pub, peer.Key) {
			return errors.New("not the paired device")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	n.seen(peer.ID, name, addr)
	return c, nil
}

// seen records a paired device's current name and address.
func (n *Node) seen(id, name, addr string) {
	n.mu.Lock()
	p, ok := n.paired[id]
	changed := ok && (name != "" && p.Name != name || addr != "" && p.Addr != addr)
	if changed {
		if name != "" {
			p.Name = name
		}
		if addr != "" {
			p.Addr = addr
		}
		n.paired[id] = p
	}
	n.mu.Unlock()
	if changed {
		_ = n.Save()
	}
}

// ForgetAddr drops the last known addresses of the paired devices.
func (n *Node) ForgetAddr() {
	n.mu.Lock()
	defer n.mu.Unlock()
	for id, p := range n.paired {
		p.Addr = ""
		n.paired[id] = p
	}
}

// Pair asks the device at addr to pair (device-pairing spec: Pair a
// device): its user accepts the request, both users confirm that their
// codes match, and both devices store each other.
func (n *Node) Pair(ctx context.Context, addr string) (Device, error) {
	c, name, err := n.dial(ctx, addr, func(ed25519.PublicKey) error { return nil })
	if IsTimeout(err) {
		return Device{}, core.NewError("syncPairTimeout", "The device at %s did not answer. A firewall on it may block sync: allow incoming TCP connections on port %s there, or start the pairing on that device.", addr, Port(addr))
	}
	if err != nil {
		return Device{}, core.NewError("syncPairConnectFailed", "Could not reach the device at %s: %v", addr, err)
	}
	defer c.Close()
	if err := c.Send(&cfsync.Message{Type: msgPairRequest, Name: n.Name()}); err != nil {
		return Device{}, err
	}
	m, err := c.Recv()
	if err != nil {
		return Device{}, err
	}
	switch m.Type {
	case msgPairAccepted:
	case msgPairRefused:
		return Device{}, core.NewError("syncPairNotVisible", "%s is not visible for pairing. Open “Pair a device” on it and try again.", name)
	case msgPairDeclined:
		n.ui.PairingFinished(name, false)
		return Device{}, core.NewError("syncPairDeclined", "%s declined the pairing request.", name)
	default:
		return Device{}, fmt.Errorf("unexpected message %q", m.Type)
	}
	return n.confirm(ctx, c, m.Name, addr)
}

// acceptPairing handles a pairing request on an accepted connection.
func (n *Node) acceptPairing(c *Conn) {
	defer c.Close()
	m, err := c.Recv()
	if err != nil || m.Type != msgPairRequest {
		return
	}
	n.mu.Lock()
	ok := n.now().Before(n.pairUntil) && !n.pending
	if ok {
		n.pending = true
	}
	n.mu.Unlock()
	if !ok {
		_ = c.Send(&cfsync.Message{Type: msgPairRefused})
		return
	}
	defer func() {
		n.mu.Lock()
		n.pending = false
		n.mu.Unlock()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), PairingTimeout)
	defer cancel()
	accepted, err := n.ui.PairingRequest(ctx, m.Name)
	if err != nil || !accepted {
		_ = c.Send(&cfsync.Message{Type: msgPairDeclined})
		return
	}
	if err := c.Send(&cfsync.Message{Type: msgPairAccepted, Name: n.Name()}); err != nil {
		return
	}
	_, _ = n.confirm(context.Background(), c, m.Name, "")
}

// confirm shows the code to the user and exchanges the confirmations; when
// both users confirmed in time, the peer is stored.
func (n *Node) confirm(ctx context.Context, c *Conn, name, addr string) (Device, error) {
	ctx, cancel := context.WithTimeout(ctx, PairingTimeout)
	defer cancel()
	code, err := c.ExportCode()
	if err != nil {
		return Device{}, err
	}
	peerAnswer := make(chan bool, 1)
	go func() {
		m, err := c.Recv()
		peerAnswer <- err == nil && m.Type == msgPairConfirmed
		if err != nil || m.Type != msgPairConfirmed {
			cancel() // close the code dialog
		}
	}()
	mine, err := n.ui.ConfirmPairing(ctx, name, code)
	mine = mine && err == nil && ctx.Err() == nil
	reply := msgPairDeclined
	if mine {
		reply = msgPairConfirmed
	}
	_ = c.Send(&cfsync.Message{Type: reply})
	theirs := false
	if mine {
		select {
		case theirs = <-peerAnswer:
		case <-ctx.Done():
		}
	}
	if !mine || !theirs {
		n.ui.PairingFinished(name, false)
		return Device{}, core.NewError("syncPairFailed", "Pairing with %s failed: the codes were not confirmed on both devices.", name)
	}
	p := Device{ID: c.PeerID(), Name: name, Key: slices.Clone(c.PeerKey()), Addr: addr}
	n.mu.Lock()
	n.paired[p.ID] = p
	delete(n.revoked, p.ID)
	n.mu.Unlock()
	if err := n.Save(); err != nil {
		return Device{}, err
	}
	n.OnPaired(p)
	n.ui.PairingFinished(name, true)
	return p, nil
}

// Unpair removes a paired device (device-pairing spec: Unpair a device):
// its key goes to the unpaired list, so that it is told when it next
// connects, and it is told now if it can be reached at addr (or its last
// known address).
func (n *Node) Unpair(ctx context.Context, id, addr string) error {
	n.mu.Lock()
	p, ok := n.paired[id]
	if ok {
		delete(n.paired, id)
		n.revoked[id] = p.Key
	}
	n.mu.Unlock()
	if !ok {
		return nil
	}
	if err := n.Save(); err != nil {
		return err
	}
	n.OnUnpaired(id)
	if addr == "" {
		addr = p.Addr
	}
	if addr == "" {
		return nil
	}
	c, _, err := n.dial(ctx, addr, func(pub ed25519.PublicKey) error {
		if !bytes.Equal(pub, p.Key) {
			return errors.New("not the unpaired device")
		}
		return nil
	})
	if err != nil {
		return nil // told when it next connects
	}
	defer c.Close()
	_ = c.Send(&cfsync.Message{Type: msgUnpaired})
	return nil
}

// forget removes a device that unpaired this one.
func (n *Node) forget(id string) {
	n.mu.Lock()
	_, ok := n.paired[id]
	delete(n.paired, id)
	n.mu.Unlock()
	if ok {
		_ = n.Save()
		n.OnUnpaired(id)
	}
}
