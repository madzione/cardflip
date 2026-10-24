//go:build !js

package storage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func mcpFile(dataDir string) string { return filepath.Join(dataDir, "mcp.json") }

// ReadMcp returns the stored MCP server settings of the device, or
// ErrNotFound when none were written yet. Like the profile registry they
// live in the data directory, outside every collection.
func ReadMcp(dataDir string) ([]byte, error) {
	if dataDir == "" {
		return nil, errors.New("storage: no data directory")
	}
	data, err := os.ReadFile(mcpFile(dataDir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

// WriteMcp replaces the stored MCP server settings atomically. The file is
// readable only by its owner, as it holds the server's token.
func WriteMcp(dataDir string, data []byte) error {
	if dataDir == "" {
		return errors.New("storage: no data directory")
	}
	// writeAtomic renames a temporary file created with mode 0600.
	return writeAtomic(mcpFile(dataDir), data)
}
