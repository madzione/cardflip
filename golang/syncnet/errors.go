package syncnet

import (
	"context"
	"errors"
	"net"
	"os"
)

// IsTimeout reports whether a connection failed because the other side
// did not answer in time. On a local network that usually means a
// firewall on the other device drops the connection, where a device that
// is on but does not listen refuses it at once.
func IsTimeout(err error) bool {
	if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// Port returns the port of addr ("host:port"), or addr if it has none.
func Port(addr string) string {
	if _, port, err := net.SplitHostPort(addr); err == nil {
		return port
	}
	return addr
}
