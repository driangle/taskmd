package web

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSEBroker_Broadcast(t *testing.T) {
	broker := NewSSEBroker()

	// Register a client channel
	ch := make(chan struct{}, 1)
	broker.mu.Lock()
	broker.clients[ch] = struct{}{}
	broker.mu.Unlock()

	// Broadcast
	broker.Broadcast()

	select {
	case <-ch:
		// OK
	case <-time.After(time.Second):
		t.Fatal("expected broadcast to reach client")
	}
}

func TestSSEBroker_ServeHTTP(t *testing.T) {
	broker := NewSSEBroker()

	server := httptest.NewServer(broker)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "text/event-stream" {
		t.Fatalf("expected text/event-stream, got %s", ct)
	}

	// Read initial "connected" event
	scanner := bufio.NewScanner(resp.Body)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if line == "" && len(lines) > 1 {
			break
		}
	}

	found := false
	for _, line := range lines {
		if strings.Contains(line, "event: connected") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'event: connected', got: %v", lines)
	}

	// Now broadcast and read reload event
	broker.Broadcast()

	lines = nil
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if line == "" && len(lines) > 1 {
			break
		}
	}

	found = false
	for _, line := range lines {
		if strings.Contains(line, "event: reload") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'event: reload', got: %v", lines)
	}
}
