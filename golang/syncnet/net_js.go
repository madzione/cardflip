//go:build js

package syncnet

import (
	"context"
	"crypto/ed25519"
	"errors"

	cfsync "cardflip/golang/sync"
)

// Built is false in the web build: browsers can neither find
// devices nor accept connections (app-shell spec: no device sync on the
// web).
const Built = false

const DefaultPort = 47812

var errUnsupported = errors.New("device sync is not available on the web")

type Identity struct{ id string }

func (id *Identity) ID() string                   { return id.id }
func (id *Identity) PublicKey() ed25519.PublicKey { return nil }

func NewKey() []byte { return nil }

func DeviceID(pub ed25519.PublicKey) string { return "" }

func IdentityFromSeed(seed []byte) (*Identity, error) { return nil, errUnsupported }

type Conn struct{}

func (c *Conn) PeerKey() ed25519.PublicKey     { return nil }
func (c *Conn) PeerID() string                 { return "" }
func (c *Conn) Send(m *cfsync.Message) error   { return errUnsupported }
func (c *Conn) Recv() (*cfsync.Message, error) { return nil, errUnsupported }
func (c *Conn) ExportCode() (string, error)    { return "", errUnsupported }
func (c *Conn) Close() error                   { return nil }

func Dial(ctx context.Context, id *Identity, addr string, verify func(ed25519.PublicKey) error) (*Conn, error) {
	return nil, errUnsupported
}

type Listener struct{}

func Listen(id *Identity, port int, verify func(ed25519.PublicKey) error, handle func(*Conn)) (*Listener, error) {
	return nil, errUnsupported
}

func (l *Listener) Port() int { return 0 }
func (l *Listener) Close()    {}

func LocalNetwork() (subnet, addr string) { return "", "" }
