//go:build js

package main

// mcpServerBuilt is false in the web build, which has no MCP server
// (agent-access spec: not offered on other platforms).
const mcpServerBuilt = false

type mcpServer struct{}

func startMcpServer(s *McpService, port int) (*mcpServer, error) { return nil, errMcpUnsupported }

func (m *mcpServer) stop() {}
