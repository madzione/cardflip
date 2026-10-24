//go:build !js

package syncnet

import (
	"context"
	"crypto/ed25519"
	"errors"
	"net"
	"strconv"
	gosync "sync"
	"testing"
	"time"

	"cardflip/golang/internal/usermsg"
	cfsync "cardflip/golang/sync"
)

// fakeSyncUI answers like a user: it accepts pairing requests if accept is
// set, and confirms the code if confirm says so.
type fakeSyncUI struct {
	mu       gosync.Mutex
	accept   bool
	confirm  func(ctx context.Context, code string) bool
	requests []string
	codes    []string
	finished []bool
}

func (u *fakeSyncUI) PairingRequest(ctx context.Context, name string) (bool, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.requests = append(u.requests, name)
	return u.accept, nil
}

func (u *fakeSyncUI) ConfirmPairing(ctx context.Context, name, code string) (bool, error) {
	u.mu.Lock()
	u.codes = append(u.codes, code)
	confirm := u.confirm
	u.mu.Unlock()
	return confirm(ctx, code), nil
}

func (u *fakeSyncUI) PairingFinished(name string, paired bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.finished = append(u.finished, paired)
}

// comparingUsers returns confirm functions for two users who compare the
// codes their screens show.
func comparingUsers() (func(context.Context, string) bool, func(context.Context, string) bool) {
	a, b := make(chan string, 1), make(chan string, 1)
	user := func(mine, other chan string) func(context.Context, string) bool {
		return func(ctx context.Context, code string) bool {
			mine <- code
			select {
			case theirs := <-other:
				other <- theirs
				return theirs == code
			case <-ctx.Done():
				return false
			}
		}
	}
	return user(a, b), user(b, a)
}

type testNode struct {
	*Node
	ui *fakeSyncUI
	l  *Listener
}

func (n *testNode) addr() string { return net.JoinHostPort("127.0.0.1", strconv.Itoa(n.l.Port())) }

func startNode(t *testing.T, name string) *testNode {
	t.Helper()
	ui := &fakeSyncUI{accept: true, confirm: func(context.Context, string) bool { return true }}
	n := NewNode(testIdentity(t), name, ui, time.Now, nil, nil)
	l, err := Listen(n.id, 0, n.Verify, n.Handle)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(l.Close)
	return &testNode{Node: n, ui: ui, l: l}
}

func isPaired(n *testNode, other *testNode) bool {
	for _, p := range n.Devices() {
		if p.ID == other.id.ID() && ed25519.PublicKey(p.Key).Equal(other.id.PublicKey()) {
			return true
		}
	}
	return false
}

// Matching codes confirmed on both devices pair them.
func TestPairingMatchingCodes(t *testing.T) {
	laptop, phone := startNode(t, "Laptop"), startNode(t, "Phone")
	laptop.ui.confirm, phone.ui.confirm = comparingUsers()
	laptop.OpenPairing(PairingWindow)
	p, err := phone.Pair(context.Background(), laptop.addr())
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Laptop" || p.ID != laptop.id.ID() {
		t.Fatalf("paired %+v", p)
	}
	if !isPaired(phone, laptop) || !waitFor(func() bool { return isPaired(laptop, phone) }) {
		t.Fatal("not paired on both devices")
	}
	if laptop.ui.requests[0] != "Phone" || len(laptop.ui.codes) != 1 || laptop.ui.codes[0] != phone.ui.codes[0] {
		t.Fatalf("requests %v, codes %v and %v", laptop.ui.requests, laptop.ui.codes, phone.ui.codes)
	}
	// Now the phone may connect outside the pairing window.
	laptop.OpenPairing(0)
	sessions := make(chan string, 1)
	laptop.OnSession = func(peer Device, _ *cfsync.Message, c SessionConn) {
		defer c.Close()
		if m, err := c.Recv(); err == nil {
			sessions <- peer.Name + ":" + m.Type
		}
	}
	c, err := phone.Connect(context.Background(), p, laptop.addr())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.Send(&cfsync.Message{Type: "hello"}); err != nil {
		t.Fatal(err)
	}
	if got := <-sessions; got != "Phone:hello" {
		t.Fatalf("session %q", got)
	}
}

func waitFor(cond func() bool) bool {
	for range 200 {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// A code declined on one device stores nothing on either.
func TestPairingDeclinedCode(t *testing.T) {
	laptop, phone := startNode(t, "Laptop"), startNode(t, "Phone")
	laptop.ui.confirm = func(context.Context, string) bool { return false }
	laptop.OpenPairing(PairingWindow)
	if _, err := phone.Pair(context.Background(), laptop.addr()); err == nil {
		t.Fatal("pairing succeeded")
	}
	time.Sleep(50 * time.Millisecond)
	if len(phone.Devices())+len(laptop.Devices()) != 0 {
		t.Fatal("a device stored the other")
	}
	if !waitFor(func() bool {
		laptop.ui.mu.Lock()
		defer laptop.ui.mu.Unlock()
		return len(laptop.ui.finished) == 1 && !laptop.ui.finished[0]
	}) || len(phone.ui.finished) != 1 || phone.ui.finished[0] {
		t.Fatalf("finished: laptop %v, phone %v", laptop.ui.finished, phone.ui.finished)
	}
}

// A declined request stores nothing.
func TestPairingDeclinedRequest(t *testing.T) {
	laptop, phone := startNode(t, "Laptop"), startNode(t, "Phone")
	laptop.ui.accept = false
	laptop.OpenPairing(PairingWindow)
	if _, err := phone.Pair(context.Background(), laptop.addr()); err == nil {
		t.Fatal("pairing succeeded")
	}
	if len(phone.Devices())+len(laptop.Devices()) != 0 || len(phone.ui.codes) != 0 {
		t.Fatal("pairing went on after the request was declined")
	}
}

// Confirming too late stores nothing.
func TestPairingTimeout(t *testing.T) {
	old := PairingTimeout
	PairingTimeout = 300 * time.Millisecond
	defer func() { PairingTimeout = old }()
	laptop, phone := startNode(t, "Laptop"), startNode(t, "Phone")
	laptop.ui.confirm = func(ctx context.Context, _ string) bool {
		<-ctx.Done() // the user does not answer
		return true
	}
	laptop.OpenPairing(PairingWindow)
	if _, err := phone.Pair(context.Background(), laptop.addr()); err == nil {
		t.Fatal("pairing succeeded")
	}
	time.Sleep(100 * time.Millisecond)
	if len(phone.Devices())+len(laptop.Devices()) != 0 {
		t.Fatal("a device stored the other")
	}
}

// A device that does not answer, as behind a firewall that drops sync
// connections, is reported as such, with the port to allow; one that
// refuses the connection is not.
func TestPairingNoAnswer(t *testing.T) {
	phone := startNode(t, "Phone")
	// The kernel accepts the connection, but nothing answers the TLS
	// handshake.
	silent, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer silent.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, err = phone.Pair(ctx, silent.Addr().String())
	_, port, _ := net.SplitHostPort(silent.Addr().String())
	if usermsg.Code(err.Error()) != "syncPairTimeout" || usermsg.Arguments(err.Error())[1] != port {
		t.Fatalf("Pair = %v", err)
	}
	addr := silent.Addr().String()
	silent.Close()
	if _, err = phone.Pair(context.Background(), addr); usermsg.Code(err.Error()) != "syncPairConnectFailed" {
		t.Fatalf("Pair = %v", err)
	}
}

// A man in the middle runs two TLS sessions, so the two devices show
// different codes, and users who compare them pair nothing.
func TestPairingManInTheMiddle(t *testing.T) {
	laptop, phone := startNode(t, "Laptop"), startNode(t, "Phone")
	laptop.ui.confirm, phone.ui.confirm = comparingUsers()
	laptop.OpenPairing(PairingWindow)
	mitm := testIdentity(t)
	any := func(ed25519.PublicKey) error { return nil }
	proxy, err := Listen(mitm, 0, any, func(toPhone *Conn) {
		defer toPhone.Close()
		toLaptop, err := Dial(context.Background(), mitm, laptop.addr(), any)
		if err != nil {
			return
		}
		defer toLaptop.Close()
		relay := func(from, to *Conn) {
			for {
				m, err := from.Recv()
				if err != nil || to.Send(m) != nil {
					from.Close()
					to.Close()
					return
				}
			}
		}
		go relay(toPhone, toLaptop)
		relay(toLaptop, toPhone)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()
	_, err = phone.Pair(context.Background(), net.JoinHostPort("127.0.0.1", strconv.Itoa(proxy.Port())))
	if err == nil {
		t.Fatal("pairing through a man in the middle succeeded")
	}
	if len(laptop.ui.codes) != 1 || len(phone.ui.codes) != 1 || laptop.ui.codes[0] == phone.ui.codes[0] {
		t.Fatalf("codes %v and %v", laptop.ui.codes, phone.ui.codes)
	}
	time.Sleep(50 * time.Millisecond)
	if len(phone.Devices())+len(laptop.Devices()) != 0 {
		t.Fatal("a device stored the man in the middle")
	}
}

// Outside the pairing window, an unknown device cannot even connect.
func TestPairingOutsideWindow(t *testing.T) {
	laptop, phone := startNode(t, "Laptop"), startNode(t, "Phone")
	if _, err := phone.Pair(context.Background(), laptop.addr()); err == nil {
		t.Fatal("pairing succeeded")
	}
	if len(laptop.ui.requests) != 0 || len(laptop.Devices()) != 0 {
		t.Fatal("request shown outside the pairing window")
	}
	// Once the window closes, it is refused again.
	laptop.OpenPairing(PairingWindow)
	laptop.OpenPairing(0)
	if _, err := phone.Pair(context.Background(), laptop.addr()); err == nil {
		t.Fatal("pairing succeeded after the window closed")
	}
}

// pairBoth pairs a and b.
func pairBoth(t *testing.T, a, b *testNode) Device {
	t.Helper()
	a.OpenPairing(PairingWindow)
	p, err := b.Pair(context.Background(), a.addr())
	if err != nil {
		t.Fatal(err)
	}
	a.OpenPairing(0)
	if !waitFor(func() bool { return isPaired(a, b) }) {
		t.Fatal("not paired")
	}
	return p
}

// A device unpaired while it was away learns it at its next connection
// and unpairs too; one that can be reached learns it at once.
func TestUnpair(t *testing.T) {
	laptop, phone := startNode(t, "Laptop"), startNode(t, "Phone")
	toLaptop := pairBoth(t, laptop, phone)
	var unpaired []string
	var mu gosync.Mutex
	phone.OnUnpaired = func(id string) {
		mu.Lock()
		defer mu.Unlock()
		unpaired = append(unpaired, id)
	}
	// The phone is away.
	phone.l.Close()
	if err := laptop.Unpair(context.Background(), phone.id.ID(), ""); err != nil {
		t.Fatal(err)
	}
	if isPaired(laptop, phone) {
		t.Fatal("laptop still lists the phone")
	}
	if !isPaired(phone, laptop) {
		t.Fatal("setup: the phone learned it too early")
	}
	if _, err := phone.Connect(context.Background(), toLaptop, laptop.addr()); !errors.Is(err, ErrUnpaired) {
		t.Fatalf("connect after unpairing = %v", err)
	}
	if isPaired(phone, laptop) || len(unpaired) != 1 || unpaired[0] != laptop.id.ID() {
		t.Fatalf("phone still paired (unpaired %v)", unpaired)
	}

	// Reachable: told at once.
	tablet := startNode(t, "Tablet")
	pairBoth(t, laptop, tablet)
	if err := laptop.Unpair(context.Background(), tablet.id.ID(), tablet.addr()); err != nil {
		t.Fatal(err)
	}
	if !waitFor(func() bool { return !isPaired(tablet, laptop) }) {
		t.Fatal("the reachable tablet still lists the laptop")
	}
}
