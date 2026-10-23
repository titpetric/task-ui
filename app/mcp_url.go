package app

import (
	"net"
)

// MCPURL is the loopback MCP endpoint Claude runs reach for addr, 127.0.0.1 for one on every address.
// MCPURL returns "" before the server listens.
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
