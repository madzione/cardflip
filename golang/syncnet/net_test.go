//go:build !js

package syncnet

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	cfsync "cardflip/golang/sync"
)

func testIdentity(t *testing.T) *Identity {
	t.Helper()
	id, err := IdentityFromSeed(NewKey())
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// pinned accepts only the keys of ids.
func pinned(ids ...*Identity) func(ed25519.PublicKey) error {
	return func(pub ed25519.PublicKey) error {
		for _, id := range ids {
			if pub.Equal(id.key.Public()) {
				return nil
			}
		}
		return errors.New("not paired")
	}
}

// echoServer listens with server's identity, accepting keys verify
// accepts, and echoes every message back; got receives the connections'
// peer IDs.
func echoServer(t *testing.T, server *Identity, verify func(ed25519.PublicKey) error) (*Listener, chan string) {
	t.Helper()
	got := make(chan string, 4)
	l, err := Listen(server, 0, verify, func(c *Conn) {
		defer c.Close()
		got <- c.PeerID()
		for {
			m, err := c.Recv()
			if err != nil {
				return
			}
			if err := c.Send(m); err != nil {
				return
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(l.Close)
	return l, got
}

func addrOf(l *Listener) string { return net.JoinHostPort("127.0.0.1", strconv.Itoa(l.Port())) }

func TestDeviceIdentity(t *testing.T) {
	seed := NewKey()
	a, err := IdentityFromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := IdentityFromSeed(seed)
	if a.id != b.id || len(a.id) != 16 || strings.ToUpper(a.id) != a.id {
		t.Fatalf("device IDs %q and %q", a.id, b.id)
	}
	if c := testIdentity(t); c.id == a.id {
		t.Fatal("two keys with one ID")
	}
	if _, err := IdentityFromSeed([]byte("short")); err == nil {
		t.Fatal("invalid seed accepted")
	}
}

// Paired devices connect and exchange messages; each sees the other's
// device ID.
func TestSyncConnPaired(t *testing.T) {
	laptop, phone := testIdentity(t), testIdentity(t)
	l, got := echoServer(t, laptop, pinned(phone))
	c, err := Dial(context.Background(), phone, addrOf(l), pinned(laptop))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if c.PeerID() != laptop.id {
		t.Fatalf("phone sees %s, want %s", c.PeerID(), laptop.id)
	}
	if peer := <-got; peer != phone.id {
		t.Fatalf("laptop sees %s, want %s", peer, phone.id)
	}
	big := bytes.Repeat([]byte("compressible "), 1<<20)
	if err := c.Send(&cfsync.Message{Type: "ping", Name: "hello", Data: big}); err != nil {
		t.Fatal(err)
	}
	m, err := c.Recv()
	if err != nil || m.Type != "ping" || m.Name != "hello" || !bytes.Equal(m.Data, big) {
		t.Fatalf("echo = %+v, %v", m, err)
	}
	// Both ends derive the same pairing code from the session.
	if code, err := c.ExportCode(); err != nil || len(code) != 6 {
		t.Fatalf("code %q, %v", code, err)
	}
}

// A device whose key is not paired, or whose key changed, is refused in the
// handshake: no connection reaches the handler, nothing is exchanged.
func TestSyncConnRefusesUnknownKey(t *testing.T) {
	laptop, phone, stranger := testIdentity(t), testIdentity(t), testIdentity(t)
	l, got := echoServer(t, laptop, pinned(phone))
	for name, id := range map[string]*Identity{"unknown": stranger, "changed": testIdentity(t)} {
		c, err := Dial(context.Background(), id, addrOf(l), pinned(laptop))
		if err == nil {
			// TLS 1.3 clients finish before the server checks their
			// certificate; the first read fails.
			err = c.Send(&cfsync.Message{Type: "ping"})
			if err == nil {
				_, err = c.Recv()
			}
			c.Close()
		}
		if err == nil {
			t.Fatalf("%s key: exchanged a message", name)
		}
	}
	// The client refuses a server it did not pair with, too.
	if c, err := Dial(context.Background(), phone, addrOf(l), pinned(stranger)); err == nil {
		c.Close()
		t.Fatal("phone connected to an unpaired laptop")
	}
	select {
	case peer := <-got:
		t.Fatalf("handler got a connection from %s", peer)
	case <-time.After(100 * time.Millisecond):
	}
}

// A frame above 16 MB is rejected without being read, and a message that
// does not fit is not sent.
func TestSyncConnFrameLimit(t *testing.T) {
	laptop, phone := testIdentity(t), testIdentity(t)
	recvErr := make(chan error, 1)
	l, err := Listen(laptop, 0, pinned(phone), func(c *Conn) {
		defer c.Close()
		_, err := c.Recv()
		recvErr <- err
	})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	c, err := Dial(context.Background(), phone, addrOf(l), pinned(laptop))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	noise := make([]byte, maxFrame+1024)
	rand.Read(noise)
	if err := c.Send(&cfsync.Message{Type: "big", Data: noise}); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("sending an oversized message: %v", err)
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], maxFrame+1)
	if _, err := c.tls.Write(hdr[:]); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-recvErr:
		if err == nil || !strings.Contains(err.Error(), "too large") {
			t.Fatalf("receiving an oversized frame: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("oversized frame not rejected")
	}
}

// The listener takes the default port when it is free, and any port when
// it is not.
func TestListenSyncPortFallback(t *testing.T) {
	id := testIdentity(t)
	busy, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	port := busy.Addr().(*net.TCPAddr).Port
	l, err := Listen(id, port, pinned(), func(c *Conn) { c.Close() })
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if l.Port() == port || l.Port() == 0 {
		t.Fatalf("listening on %d with %d taken", l.Port(), port)
	}
}
