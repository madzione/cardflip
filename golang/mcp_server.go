//go:build !js

package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cardflip/golang/core"
)

// mcpServerBuilt is false in the web build, which has no MCP server.
const mcpServerBuilt = true

// mcpServer is the running HTTP server of McpService.
type mcpServer struct {
	http *http.Server
	done chan struct{}
}

// startMcpServer listens on the loopback interface only and serves MCP over
// Streamable HTTP at /mcp. A port in use is reported as such.
func startMcpServer(s *McpService, port int) (*mcpServer, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		if isAddrInUse(err) {
			return nil, core.NewError("mcpPortInUse", "Port %d is already in use.", port)
		}
		return nil, core.NewError("mcpListenFailed", "The server could not start: %v", err)
	}
	m := &mcpServer{http: &http.Server{Handler: mcpHandler(s), ReadHeaderTimeout: 10 * time.Second}, done: make(chan struct{})}
	go func() {
		defer close(m.done)
		m.http.Serve(ln)
	}()
	return m, nil
}

// stop ends the server and every open agent session.
func (m *mcpServer) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := m.http.Shutdown(ctx); err != nil {
		m.http.Close() // open event streams
	}
	<-m.done
}

func isAddrInUse(err error) bool {
	var errno syscall.Errno
	// 10048 is WSAEADDRINUSE on Windows.
	return errors.As(err, &errno) && (errno == syscall.EADDRINUSE || errno == 10048)
}

// mcpHandler checks a request's origin and token before the MCP handler
// sees it (design decision 4). A new MCP server per start ends the
// sessions of the previous one.
func mcpHandler(s *McpService) http.Handler {
	server := newMcpServer(s)
	// Requests may carry a 20 MB media file as base64 (store_media).
	streamable := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{MaxRequestBodyBytes: 32 << 20})
	mux := http.NewServeMux()
	mux.Handle("/mcp", streamable)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if o := r.Header.Get("Origin"); o != "" && !loopbackOrigin(o) {
			http.Error(w, "Forbidden: requests from web pages are not allowed.", http.StatusForbidden)
			return
		}
		if !validBearer(r.Header.Get("Authorization"), s.token()) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "Unauthorized: a valid bearer token is required.", http.StatusUnauthorized)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

// loopbackOrigin reports whether origin is a page served from this
// computer: http(s) on 127.0.0.1, localhost or [::1], any port.
func loopbackOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	switch u.Hostname() {
	case "127.0.0.1", "localhost", "::1":
		return true
	}
	return false
}

func validBearer(header, token string) bool {
	got, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(got)), []byte(token)) == 1
}

// newMcpServer returns the MCP server with the agent tools.
func newMcpServer(s *McpService) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "cardflip", Title: "Cardflip", Version: appVersion}, &mcp.ServerOptions{
		Instructions: mcpInstructions,
	})
	addMcpTools(server, s)
	return server
}
