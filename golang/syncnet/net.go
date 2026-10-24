//go:build !js

// Package syncnet is the network side of device sync (sync design decisions
// 8 and 9): device identities, TLS connections pinned to paired keys,
// framed messages, the listener, and pairing and unpairing.
package syncnet

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"strconv"
	gosync "sync"
	"time"

	"github.com/klauspost/compress/zstd"

	cfsync "cardflip/golang/sync"
)

// Built is false in the web build, which cannot sync.
const Built = true

// DefaultPort is the port sync listens on, one above the MCP server's,
// when it is free; otherwise any free port (sync design decision 8).
const DefaultPort = 47812

// Sync connection limits: a frame holds one compressed message of at most
// maxFrame bytes, which decompresses to at most maxMessage.
const (
	maxFrame   = 16 << 20
	maxMessage = 64 << 20
)

// IdleTimeout is how long a connection waits for the next message:
// longer than the 2 minutes a user has to confirm a pairing code.
var IdleTimeout = 3 * time.Minute

// Identity is this device's sync key, the self-signed certificate it
// presents and its ID.
type Identity struct {
	key  ed25519.PrivateKey
	cert tls.Certificate
	id   string
}

// NewKey returns the seed of a new device key.
func NewKey() []byte {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err) // crypto/rand does not fail
	}
	return key.Seed()
}

// DeviceID returns the ID of the device with public key pub: the base32 of
// the first 10 bytes of its SHA-256, 16 characters.
func DeviceID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return base32.StdEncoding.EncodeToString(sum[:10])
}

// ID returns the device's ID.
func (id *Identity) ID() string { return id.id }

// PublicKey returns the device's public key.
func (id *Identity) PublicKey() ed25519.PublicKey { return id.key.Public().(ed25519.PublicKey) }

// IdentityFromSeed returns the identity of the device key with seed.
func IdentityFromSeed(seed []byte) (*Identity, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, errors.New("invalid device key")
	}
	key := ed25519.NewKeyFromSeed(seed)
	pub := key.Public().(ed25519.PublicKey)
	id := DeviceID(pub)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "cardflip " + id},
		NotBefore:    time.Unix(0, 0),
		NotAfter:     time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, key)
	if err != nil {
		return nil, err
	}
	return &Identity{key: key, id: id, cert: tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}}, nil
}

// peerKey returns the Ed25519 key of a peer's certificate.
func peerKey(raw [][]byte) (ed25519.PublicKey, error) {
	if len(raw) == 0 {
		return nil, errors.New("the other device presented no certificate")
	}
	cert, err := x509.ParseCertificate(raw[0])
	if err != nil {
		return nil, err
	}
	pub, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("the other device's key is not a device key")
	}
	return pub, nil
}

// tlsConfig returns the configuration of sync connections (decision 8):
// TLS 1.3 only, both sides present their device certificate, and the
// handshake fails unless verify accepts the peer's key, so nothing is sent
// to or accepted from a device it refuses.
func (id *Identity) tlsConfig(server bool, verify func(ed25519.PublicKey) error) *tls.Config {
	cfg := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{id.cert},
		// The certificates are self-signed; verify checks the pinned keys.
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(raw [][]byte, _ [][]*x509.Certificate) error {
			pub, err := peerKey(raw)
			if err != nil {
				return err
			}
			return verify(pub)
		},
	}
	if server {
		cfg.ClientAuth = tls.RequireAnyClientCert
	}
	return cfg
}

// Conn is an established sync connection: framed, compressed messages
// over TLS. It implements sync.Conn.
type Conn struct {
	tls  *tls.Conn
	r    *bufio.Reader
	peer ed25519.PublicKey
	enc  *zstd.Encoder
	dec  *zstd.Decoder
	// sendMu and recvMu serialise Send and Recv, and keep Close from
	// freeing the codecs while they are used.
	sendMu, recvMu gosync.Mutex
	closeOnce      gosync.Once
	closed         bool
}

var errClosed = errors.New("sync connection closed")

func newConn(c *tls.Conn) (*Conn, error) {
	var raw [][]byte
	for _, cert := range c.ConnectionState().PeerCertificates {
		raw = append(raw, cert.Raw)
	}
	pub, err := peerKey(raw)
	if err != nil {
		return nil, err
	}
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	dec, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxMemory(maxMessage))
	if err != nil {
		return nil, err
	}
	return &Conn{tls: c, r: bufio.NewReader(c), peer: pub, enc: enc, dec: dec}, nil
}

// PeerID returns the peer's device ID.
func (c *Conn) PeerID() string { return DeviceID(c.peer) }

// PeerKey returns the peer's device key.
func (c *Conn) PeerKey() ed25519.PublicKey { return c.peer }

// Send sends one message.
func (c *Conn) Send(m *cfsync.Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	c.sendMu.Lock()
	if c.closed {
		c.sendMu.Unlock()
		return errClosed
	}
	frame := c.enc.EncodeAll(data, make([]byte, 4, 4+len(data)/2))
	c.sendMu.Unlock()
	if len(frame)-4 > maxFrame {
		return fmt.Errorf("sync message of %d bytes is too large", len(frame)-4)
	}
	binary.BigEndian.PutUint32(frame, uint32(len(frame)-4))
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.closed {
		return errClosed
	}
	c.tls.SetWriteDeadline(time.Now().Add(IdleTimeout))
	_, err = c.tls.Write(frame)
	return err
}

// Recv receives one message. A frame above the limit ends the connection
// unread.
func (c *Conn) Recv() (*cfsync.Message, error) {
	c.recvMu.Lock()
	defer c.recvMu.Unlock()
	if c.closed {
		return nil, errClosed
	}
	c.tls.SetReadDeadline(time.Now().Add(IdleTimeout))
	var hdr [4]byte
	if _, err := io.ReadFull(c.r, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n > maxFrame {
		c.tls.Close()
		return nil, fmt.Errorf("sync frame of %d bytes is too large", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(c.r, buf); err != nil {
		return nil, err
	}
	data, err := c.dec.DecodeAll(buf, nil)
	if err != nil {
		c.tls.Close()
		return nil, fmt.Errorf("sync frame: %w", err)
	}
	var m cfsync.Message
	if err := json.Unmarshal(data, &m); err != nil {
		c.tls.Close()
		return nil, fmt.Errorf("sync message: %w", err)
	}
	return &m, nil
}

// ExportCode returns the pairing code of the connection (decision 9): six
// digits derived from the TLS session, which differ between the two
// sessions a man in the middle would run.
func (c *Conn) ExportCode() (string, error) {
	cs := c.tls.ConnectionState()
	b, err := cs.ExportKeyingMaterial("cardflip pairing", nil, 8)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", binary.BigEndian.Uint64(b)%1000000), nil
}

// Close closes the connection; a Recv or Send in progress fails.
func (c *Conn) Close() error {
	err := c.tls.Close()
	c.closeOnce.Do(func() {
		c.recvMu.Lock()
		c.sendMu.Lock()
		c.closed = true
		c.enc.Close()
		c.dec.Close()
		c.sendMu.Unlock()
		c.recvMu.Unlock()
	})
	return err
}

// Dial connects to a device at addr, which must present a key verify
// accepts.
func Dial(ctx context.Context, id *Identity, addr string, verify func(ed25519.PublicKey) error) (*Conn, error) {
	d := tls.Dialer{NetDialer: &net.Dialer{Timeout: 10 * time.Second}, Config: id.tlsConfig(false, verify)}
	c, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	sc, err := newConn(c.(*tls.Conn))
	if err != nil {
		c.Close()
		return nil, err
	}
	return sc, nil
}

// Listener accepts sync connections.
type Listener struct {
	ln   net.Listener
	cfg  *tls.Config
	done chan struct{}
}

// Listen listens on every interface on DefaultPort, or on any free
// port when that one is taken. handle gets every connection whose key
// verify accepts, after the handshake; it closes it.
func Listen(id *Identity, port int, verify func(ed25519.PublicKey) error, handle func(*Conn)) (*Listener, error) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil && port != 0 {
		ln, err = net.Listen("tcp", ":0")
	}
	if err != nil {
		return nil, err
	}
	l := &Listener{ln: ln, cfg: id.tlsConfig(true, verify), done: make(chan struct{})}
	go l.serve(handle)
	return l, nil
}

// Port returns the port the listener is on.
func (l *Listener) Port() int { return l.ln.Addr().(*net.TCPAddr).Port }

func (l *Listener) serve(handle func(*Conn)) {
	defer close(l.done)
	for {
		c, err := l.ln.Accept()
		if err != nil {
			return
		}
		go func() {
			tc := tls.Server(c, l.cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := tc.HandshakeContext(ctx); err != nil {
				tc.Close()
				return
			}
			sc, err := newConn(tc)
			if err != nil {
				tc.Close()
				return
			}
			handle(sc)
		}()
	}
}

// Close stops accepting connections. Open connections are closed by their
// handlers.
func (l *Listener) Close() {
	l.ln.Close()
	<-l.done
}

// LocalNetwork returns the subnet and address of this device on the local
// network: those of the first interface that is up with a private IPv4
// address. Wired desktops are identified by the subnet (decision 11).
func LocalNetwork() (subnet, addr string) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipn.IP.To4()
			if ip == nil || !ip.IsPrivate() {
				continue
			}
			ones, _ := ipn.Mask.Size()
			return fmt.Sprintf("%s/%d", ip.Mask(ipn.Mask), ones), ip.String()
		}
	}
	return "", ""
}
