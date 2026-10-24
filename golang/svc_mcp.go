package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/storage"
)

// DefaultMcpPort is the port the MCP server listens on until the user picks
// another. It is fixed, so agent configurations survive restarts.
const DefaultMcpPort = 47811

// The ports the user can pick.
const (
	MinMcpPort = 1024
	MaxMcpPort = 65535
)

// maxMcpActivity is how many agent changes Settings lists.
const maxMcpActivity = 50

var errMcpUnsupported = errors.New("the MCP server is only available in the desktop app")

// mcpSettings is mcp.json in the data root: device settings, the same for
// every profile and never part of a collection or backup.
type mcpSettings struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Token   string `json:"token"`
}

// McpStatus is the state of the MCP server for agents, for Settings.
type McpStatus struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
	// Token is the bearer token agents must send; empty until the server
	// is first turned on.
	Token string `json:"token"`
	URL   string `json:"url"`
	// Running is set while the server accepts connections.
	Running bool `json:"running"`
	// Error says why the server is not running although it is enabled, such
	// as a port in use (a tagged text).
	Error string `json:"error,omitempty"`
}

// McpActivity is a change an agent made: the name of its undo step (a
// tagged text) and when.
type McpActivity struct {
	Time time.Time `json:"time"`
	Step string    `json:"step"`
}

// McpService runs the MCP server through which agents work on the open
// collection (agent-access spec). The server runs while the app is open and
// the user turned it on; it follows profile switches, as its tools ask the
// app for the open collection on every call.
type McpService struct {
	a *app

	mu       sync.Mutex
	root     string // the data root, once the app was opened
	settings mcpSettings
	server   *mcpServer // nil when not running
	err      string
	activity []McpActivity // oldest first
	// changed is closed and replaced when the status or activity changes.
	changed chan struct{}

	// The storage functions of mcp.json and the platform check; tests
	// replace them.
	read      func(root string) ([]byte, error)
	write     func(root string, data []byte) error
	supported bool
}

func newMcpService(a *app) *McpService {
	s := &McpService{a: a, changed: make(chan struct{}), read: storage.ReadMcp, write: storage.WriteMcp,
		settings:  mcpSettings{Port: DefaultMcpPort},
		supported: mcpServerBuilt && (runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "windows")}
	a.onOpen = append(a.onOpen, s.opened)
	return s
}

// Supported reports whether this platform has the MCP server: only the
// desktop app on Linux, macOS and Windows.
func (s *McpService) Supported() bool { return s.supported }

// Status returns the server's state.
func (s *McpService) Status() McpStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statusLocked()
}

func (s *McpService) statusLocked() McpStatus {
	return McpStatus{Enabled: s.settings.Enabled, Port: s.settings.Port, Token: s.settings.Token,
		URL: mcpURL(s.settings.Port), Running: s.server != nil, Error: s.err}
}

func mcpURL(port int) string { return fmt.Sprintf("http://127.0.0.1:%d/mcp", port) }

// opened loads the device settings once the app knows its data root, and
// starts the server when it is on. The app calls it once per process, so a
// hot restart leaves a running server alone.
func (s *McpService) opened(root string) {
	if !s.supported {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.root = root
	data, err := s.read(root)
	switch {
	case errors.Is(err, storage.ErrNotFound):
	case err != nil:
		s.err = core.NewError("mcpSettingsUnreadable", "The agent settings could not be read: %v", err).Error()
	default:
		var st mcpSettings
		if err := json.Unmarshal(data, &st); err != nil {
			s.err = core.NewError("mcpSettingsUnreadable", "The agent settings could not be read: %v", err).Error()
			break
		}
		if st.Port < MinMcpPort || st.Port > MaxMcpPort {
			st.Port = DefaultMcpPort
		}
		s.settings = st
	}
	if s.settings.Enabled {
		s.startLocked()
	}
	s.notifyLocked()
}

// SetEnabled turns the server on or off. The token is created the first
// time it is turned on and kept afterwards.
func (s *McpService) SetEnabled(on bool) (McpStatus, error) {
	return s.update(func(st *mcpSettings) error {
		st.Enabled = on
		if on && st.Token == "" {
			st.Token = newMcpToken()
		}
		return nil
	})
}

// SetPort moves the server to another port, 1024 to 65535.
func (s *McpService) SetPort(port int) (McpStatus, error) {
	if port < MinMcpPort || port > MaxMcpPort {
		return s.Status(), core.NewError("mcpPortRange", "The port must be between %d and %d.", MinMcpPort, MaxMcpPort)
	}
	return s.update(func(st *mcpSettings) error {
		st.Port = port
		return nil
	})
}

// RegenerateToken replaces the token; the old one stops working at once.
func (s *McpService) RegenerateToken() (McpStatus, error) {
	return s.update(func(st *mcpSettings) error {
		st.Token = newMcpToken()
		return nil
	})
}

// update changes the settings, saves them and restarts the server when
// enabled or port changed.
func (s *McpService) update(edit func(*mcpSettings) error) (McpStatus, error) {
	if !s.supported {
		return McpStatus{}, errMcpUnsupported
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.root == "" {
		return s.statusLocked(), errNotOpened
	}
	st := s.settings
	if err := edit(&st); err != nil {
		return s.statusLocked(), err
	}
	data, err := json.Marshal(st)
	if err != nil {
		return s.statusLocked(), err
	}
	if err := s.write(s.root, data); err != nil {
		return s.statusLocked(), core.NewError("mcpSettingsNotSaved", "The agent settings could not be saved: %v", err)
	}
	old := s.settings
	s.settings = st
	if old.Enabled != st.Enabled || old.Port != st.Port || (st.Enabled && s.server == nil) {
		s.stopLocked()
		s.err = ""
		if st.Enabled {
			s.startLocked()
		}
	}
	s.notifyLocked()
	return s.statusLocked(), nil
}

func (s *McpService) startLocked() {
	srv, err := startMcpServer(s, s.settings.Port)
	if err != nil {
		s.err = err.Error()
		return
	}
	s.server, s.err = srv, ""
}

func (s *McpService) stopLocked() {
	if s.server != nil {
		s.server.stop()
		s.server = nil
	}
}

// token returns the token requests must carry.
func (s *McpService) token() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.settings.Token
}

func newMcpToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Activity returns the changes agents made in this app session, newest
// first.
func (s *McpService) Activity() []McpActivity {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]McpActivity, len(s.activity))
	for i, e := range s.activity {
		out[len(out)-1-i] = e
	}
	return out
}

// recordActivity adds an agent's change, named by its undo step.
func (s *McpService) recordActivity(step string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activity = append(s.activity, McpActivity{Time: s.a.now(), Step: step})
	if n := len(s.activity); n > maxMcpActivity {
		s.activity = append([]McpActivity(nil), s.activity[n-maxMcpActivity:]...)
	}
	s.notifyLocked()
}

func (s *McpService) notifyLocked() {
	close(s.changed)
	s.changed = make(chan struct{})
}

// StatusChanges streams the status, first the current one and then again
// whenever it or the activity list changes, until ctx ends.
func (s *McpService) StatusChanges(ctx context.Context) <-chan McpStatus {
	out := make(chan McpStatus)
	go func() {
		defer close(out)
		for {
			s.mu.Lock()
			st, changed := s.statusLocked(), s.changed
			s.mu.Unlock()
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
