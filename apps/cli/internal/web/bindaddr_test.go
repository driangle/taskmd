package web

import (
	"net"
	"strings"
	"testing"
)

func TestResolveHost_EmptyDefaultsToLoopback(t *testing.T) {
	if got := resolveHost(""); got != "127.0.0.1" {
		t.Errorf("expected empty host to default to 127.0.0.1, got %q", got)
	}
	if got := resolveHost("0.0.0.0"); got != "0.0.0.0" {
		t.Errorf("expected explicit host to be preserved, got %q", got)
	}
}

func TestListenAddr_Formats(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{"default", "", 8080, "127.0.0.1:8080"},
		{"loopback", "127.0.0.1", 3000, "127.0.0.1:3000"},
		{"wildcard", "0.0.0.0", 8080, "0.0.0.0:8080"},
		{"specific", "10.0.0.1", 8380, "10.0.0.1:8380"},
		{"ipv6 brackets", "::1", 8080, "[::1]:8080"},
		{"ipv6 wildcard brackets", "::", 8080, "[::]:8080"},
		{"hostname", "myhost", 8080, "myhost:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listenAddr(tt.host, tt.port); got != tt.want {
				t.Errorf("listenAddr(%q, %d) = %q, want %q", tt.host, tt.port, got, tt.want)
			}
		})
	}
}

func TestIsWildcardHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"", false}, // empty resolves to the loopback default
		{"0.0.0.0", true},
		{"::", true},
		{"127.0.0.1", false},
		{"::1", false},
		{"10.0.0.1", false},
		{"localhost", false},
	}
	for _, tt := range tests {
		if got := isWildcardHost(tt.host); got != tt.want {
			t.Errorf("isWildcardHost(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}

func TestIsLoopbackHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"", true},
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"::1", true},
		{"localhost", true},
		{"0.0.0.0", false},
		{"::", false},
		{"10.0.0.1", false},
		// A hostname could resolve anywhere, so it is treated as exposed.
		{"somebox.internal", false},
	}
	for _, tt := range tests {
		if got := isLoopbackHost(tt.host); got != tt.want {
			t.Errorf("isLoopbackHost(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}

func TestBrowseURL(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{"default", "", 8080, "http://127.0.0.1:8080"},
		{"wildcard maps to localhost", "0.0.0.0", 8080, "http://localhost:8080"},
		{"ipv6 wildcard maps to localhost", "::", 8080, "http://localhost:8080"},
		{"specific host kept", "10.0.0.1", 8380, "http://10.0.0.1:8380"},
		{"ipv6 literal bracketed", "::1", 8080, "http://[::1]:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BrowseURL(tt.host, tt.port); got != tt.want {
				t.Errorf("BrowseURL(%q, %d) = %q, want %q", tt.host, tt.port, got, tt.want)
			}
		})
	}
}

func TestExposureWarning_MentionsWriteAPIOnlyWhenWritable(t *testing.T) {
	writable := exposureWarning(false)
	if !strings.Contains(writable, "PUT /api/tasks/{id}") {
		t.Errorf("expected writable warning to name the write endpoint, got %q", writable)
	}
	if !strings.Contains(writable, "--readonly") {
		t.Errorf("expected writable warning to suggest --readonly, got %q", writable)
	}

	readOnly := exposureWarning(true)
	if strings.Contains(readOnly, "PUT /api/tasks/{id}") {
		t.Errorf("read-only warning should not claim files can be rewritten, got %q", readOnly)
	}
	if !strings.Contains(readOnly, "unauthenticated") {
		t.Errorf("expected read-only warning to still note the lack of auth, got %q", readOnly)
	}
}

// The regression this whole change exists to prevent: the default bind must
// not be reachable from a non-loopback address.
func TestListenAddr_DefaultIsNotWildcard(t *testing.T) {
	addr := listenAddr("", 8080)
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("default addr %q is not a valid host:port: %v", addr, err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		t.Fatalf("default host %q is not an IP literal", host)
	}
	if ip.IsUnspecified() {
		t.Errorf("default bind %q is a wildcard address; it must be loopback", addr)
	}
	if !ip.IsLoopback() {
		t.Errorf("default bind %q is not loopback", addr)
	}
}
