package web

import (
	"fmt"
	"net"
	"strconv"
)

// defaultHost is the bind address used when Config.Host is empty.
//
// The dashboard's API is unauthenticated, and unless Config.ReadOnly is set it
// serves PUT /api/tasks/{id}, which rewrites task markdown on disk. Binding
// every interface by default would put that on whatever networks the host
// happens to sit on, so the default is loopback and widening it is a
// deliberate act (--host).
const defaultHost = "127.0.0.1"

// resolveHost returns the bind host, substituting the loopback default for an
// empty value.
func resolveHost(host string) string {
	if host == "" {
		return defaultHost
	}
	return host
}

// listenAddr builds the address passed to net.Listen. JoinHostPort is what
// makes IPv6 literals work: "::1" has to reach net.Listen as "[::1]:8080".
func listenAddr(host string, port int) string {
	return net.JoinHostPort(resolveHost(host), strconv.Itoa(port))
}

// isWildcardHost reports whether host binds every interface on the machine
// ("0.0.0.0", "::", or an empty host in the pre-Host-field spelling ":8080").
func isWildcardHost(host string) bool {
	ip := net.ParseIP(resolveHost(host))
	return ip != nil && ip.IsUnspecified()
}

// isLoopbackHost reports whether host is reachable only from this machine.
// Names other than "localhost" are not resolved: a hostname could point
// anywhere, so it is treated as exposed and warned about accordingly.
func isLoopbackHost(host string) bool {
	h := resolveHost(host)
	if h == "localhost" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

// BrowseURL returns a URL that opens the dashboard from this machine.
//
// A wildcard bind has no usable address of its own — "http://0.0.0.0:8080" is
// not something a browser can follow — so it maps to localhost. Every other
// host is used as given, which is what makes the URL honest for a specific
// bind like --host 10.0.0.1.
func BrowseURL(host string, port int) string {
	h := resolveHost(host)
	if isWildcardHost(h) {
		h = "localhost"
	}
	return fmt.Sprintf("http://%s", net.JoinHostPort(h, strconv.Itoa(port)))
}

// exposureWarning returns the warning shown when the dashboard is bound
// somewhere other than loopback, or "" when there is nothing to warn about.
func exposureWarning(readOnly bool) string {
	if readOnly {
		return "WARNING: this address is reachable from outside this machine and the API is unauthenticated. " +
			"Read-only mode is on, so task files cannot be modified, but anyone who can reach it can read every task."
	}
	return "WARNING: this address is reachable from outside this machine, the API is unauthenticated, " +
		"and PUT /api/tasks/{id} rewrites task files on disk. " +
		"Use --host 127.0.0.1 to restrict it to this machine, or --readonly to disable editing."
}
