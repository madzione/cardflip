//go:build js

package storage

import "errors"

var errMcpUnsupported = errors.New("storage: the MCP server is not available on the web")

// ReadMcp is not available on the web, which has no MCP server.
func ReadMcp(dataDir string) ([]byte, error) { return nil, errMcpUnsupported }

// WriteMcp is not available on the web, which has no MCP server.
func WriteMcp(dataDir string, data []byte) error { return errMcpUnsupported }
