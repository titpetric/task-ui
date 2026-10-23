package app

import (
	"net"
)

// MCPURL is the MCP endpoint the Claude runs reach the server on, over
// the loopback of the address it listens on: 127.0.0.1 for one that
// listens on every address. Empty before the server listens.
func MCPURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	if ip := net.ParseIP(host); host == "" || (ip != nil && ip.IsUnspecified()) {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port) + "/api/mcp"
}
