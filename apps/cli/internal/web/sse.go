package web

import (
	"fmt"
	"net/http"
	"sync"
)

// SSEBroker manages Server-Sent Events connections.
type SSEBroker struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

// NewSSEBroker creates a new SSE broker.
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		clients: make(map[chan struct{}]struct{}),
	}
}

// Broadcast sends a reload event to all connected clients.
func (b *SSEBroker) Broadcast() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		select {
		case ch <- struct{}{}:
		default:
			// client is slow, skip
		}
	}
}

// ServeHTTP handles SSE connections at /api/events.
func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan struct{}, 1)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
	}()

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: ok\n\n")
	flusher.Flush()

	for {
		select {
		case <-ch:
			fmt.Fprintf(w, "event: reload\ndata: changed\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
