package web

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCorsMiddleware_SetsHeaders(t *testing.T) {
	var innerCalled bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := corsMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !innerCalled {
		t.Error("expected inner handler to be called for GET request")
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:5173" {
		t.Errorf("expected CORS origin 'http://localhost:5173', got %q", origin)
	}

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if methods != "GET, PUT, OPTIONS" {
		t.Errorf("expected CORS methods 'GET, PUT, OPTIONS', got %q", methods)
	}

	headers := rec.Header().Get("Access-Control-Allow-Headers")
	if headers != "Content-Type" {
		t.Errorf("expected CORS headers 'Content-Type', got %q", headers)
	}
}

func TestCorsMiddleware_OptionsRequest(t *testing.T) {
	var innerCalled bool
	inner := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		innerCalled = true
	})

	handler := corsMiddleware(inner)

	req := httptest.NewRequest(http.MethodOptions, "/api/tasks", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if innerCalled {
		t.Error("expected inner handler NOT to be called for OPTIONS request")
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", rec.Code)
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:5173" {
		t.Errorf("expected CORS origin header on OPTIONS response, got %q", origin)
	}
}

func TestNewServer_CreatesInstance(t *testing.T) {
	dir := createTestTaskDir(t)
	cfg := Config{
		Port:    0,
		ScanDir: dir,
		Dev:     false,
		Verbose: false,
	}

	s := NewServer(cfg)

	if s == nil {
		t.Fatal("expected non-nil server")
		return
	}
	if s.dp == nil {
		t.Error("expected non-nil data provider")
	}
	if s.broker == nil {
		t.Error("expected non-nil SSE broker")
	}
	if s.watcher == nil {
		t.Error("expected non-nil watcher")
	}
	if s.config.ScanDir != dir {
		t.Errorf("expected scan dir %q, got %q", dir, s.config.ScanDir)
	}
}

func TestNewServer_DevMode(t *testing.T) {
	dir := createTestTaskDir(t)
	cfg := Config{
		Port:    0,
		ScanDir: dir,
		Dev:     true,
	}

	s := NewServer(cfg)

	if s == nil {
		t.Fatal("expected non-nil server")
		return
	}
	if !s.config.Dev {
		t.Error("expected dev mode to be enabled")
	}
}

func TestNewServer_ReadOnlyMode(t *testing.T) {
	dir := createTestTaskDir(t)
	cfg := Config{
		Port:     0,
		ScanDir:  dir,
		ReadOnly: true,
		Version:  "1.0.0-test",
	}

	s := NewServer(cfg)

	if !s.config.ReadOnly {
		t.Error("expected read-only mode")
	}
	if s.config.Version != "1.0.0-test" {
		t.Errorf("expected version '1.0.0-test', got %q", s.config.Version)
	}
}

func TestMountFallback_ServesHTML(t *testing.T) {
	dir := createTestTaskDir(t)
	s := NewServer(Config{ScanDir: dir})

	mux := http.NewServeMux()
	s.mountFallback(mux)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %q", ct)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "taskmd") {
		t.Error("expected fallback HTML to contain 'taskmd'")
	}
	if !strings.Contains(body, "No web UI embedded") {
		t.Error("expected fallback HTML to mention no embedded UI")
	}
}

func TestMountFallback_NonRootPath(t *testing.T) {
	dir := createTestTaskDir(t)
	s := NewServer(Config{ScanDir: dir})

	mux := http.NewServeMux()
	s.mountFallback(mux)

	req := httptest.NewRequest(http.MethodGet, "/board", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for non-root path, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "taskmd") {
		t.Error("expected fallback HTML for non-root path")
	}
}

func TestMountFallback_APIReturns404(t *testing.T) {
	dir := createTestTaskDir(t)
	s := NewServer(Config{ScanDir: dir})

	mux := http.NewServeMux()
	s.mountFallback(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /api/ path, got %d", rec.Code)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	data, err2 := io.ReadAll(r)
	if err2 != nil {
		t.Fatalf("failed to read pipe: %v", err2)
	}
	return string(data)
}

func TestPrintBanner_Default(t *testing.T) {
	dir := createTestTaskDir(t)
	s := NewServer(Config{Port: 8080, ScanDir: dir})

	output := captureStdout(t, func() {
		s.printBanner()
	})

	if !strings.Contains(output, "http://127.0.0.1:8080") {
		t.Errorf("expected banner to show the loopback bind address, got %q", output)
	}
	if strings.Contains(output, "WARNING") {
		t.Errorf("unexpected exposure warning for a loopback bind, got %q", output)
	}
	if !strings.Contains(output, fmt.Sprintf("Watching %s", dir)) {
		t.Errorf("expected banner to show scan dir, got %q", output)
	}
	if strings.Contains(output, "Read-only") {
		t.Error("unexpected read-only message in default mode")
	}
	if strings.Contains(output, "Dev mode") {
		t.Error("unexpected dev mode message in default mode")
	}
}

func TestPrintBanner_ReadOnlyAndDev(t *testing.T) {
	dir := createTestTaskDir(t)
	s := NewServer(Config{Port: 3000, ScanDir: dir, ReadOnly: true, Dev: true})

	output := captureStdout(t, func() {
		s.printBanner()
	})

	if !strings.Contains(output, "Read-only mode") {
		t.Error("expected read-only message in banner")
	}
	if !strings.Contains(output, "Dev mode") {
		t.Error("expected dev mode message in banner")
	}
}

func TestPrintBanner_WildcardWarnsAndNamesBind(t *testing.T) {
	dir := createTestTaskDir(t)
	s := NewServer(Config{Port: 8080, Host: "0.0.0.0", ScanDir: dir})

	output := captureStdout(t, func() {
		s.printBanner()
	})

	// The bug in issue #23: the banner said "localhost" while bound to every
	// interface, and nothing corrected the reader.
	if !strings.Contains(output, "Bound to 0.0.0.0") {
		t.Errorf("expected banner to name the wildcard bind, got %q", output)
	}
	if !strings.Contains(output, "WARNING") {
		t.Errorf("expected an exposure warning for a wildcard bind, got %q", output)
	}
	if !strings.Contains(output, "PUT /api/tasks/{id}") {
		t.Errorf("expected the warning to name the write endpoint, got %q", output)
	}
}

func TestPrintBanner_SpecificHostShowsRealAddress(t *testing.T) {
	dir := createTestTaskDir(t)
	s := NewServer(Config{Port: 8380, Host: "10.0.0.1", ScanDir: dir})

	output := captureStdout(t, func() {
		s.printBanner()
	})

	if !strings.Contains(output, "http://10.0.0.1:8380") {
		t.Errorf("expected banner to show the configured address, got %q", output)
	}
	if strings.Contains(output, "localhost") {
		t.Errorf("banner should not claim localhost for a specific bind, got %q", output)
	}
	if strings.Contains(output, "Bound to") {
		t.Errorf("a specific bind is not a wildcard, got %q", output)
	}
	if !strings.Contains(output, "WARNING") {
		t.Errorf("expected an exposure warning for a non-loopback bind, got %q", output)
	}
}

func TestPrintBanner_WildcardReadOnlyWarnsWithoutWriteClaim(t *testing.T) {
	dir := createTestTaskDir(t)
	s := NewServer(Config{Port: 8080, Host: "0.0.0.0", ScanDir: dir, ReadOnly: true})

	output := captureStdout(t, func() {
		s.printBanner()
	})

	if !strings.Contains(output, "WARNING") {
		t.Errorf("expected an exposure warning even in read-only mode, got %q", output)
	}
	if strings.Contains(output, "PUT /api/tasks/{id}") {
		t.Errorf("read-only mode should not warn about writes, got %q", output)
	}
}

// Proof of the default against a real socket: the OS must report a loopback
// bind, not a wildcard one.
func TestDefaultBind_RealListenerIsLoopbackOnly(t *testing.T) {
	// Port 0 lets the OS choose, so this cannot collide with a real server.
	ln, err := net.Listen("tcp", listenAddr("", 0))
	if err != nil {
		t.Fatalf("failed to listen on the default address: %v", err)
	}
	defer ln.Close()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected addr type %T", ln.Addr())
	}
	if addr.IP.IsUnspecified() {
		t.Fatalf("default config bound a wildcard address %q", addr)
	}
	if !addr.IP.IsLoopback() {
		t.Fatalf("default config bound non-loopback address %q", addr)
	}
}

// A bind failure on a specific host is a different diagnosis from a port
// clash, so the error has to name the whole address.
func TestStart_ListenErrorNamesAddress(t *testing.T) {
	dir := createTestTaskDir(t)
	// 192.0.2.0/24 is TEST-NET-1 (RFC 5737): never assigned to a local
	// interface, so binding it fails everywhere.
	s := NewServer(Config{Port: 8080, Host: "192.0.2.1", ScanDir: dir})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.Start(ctx)
	if err == nil {
		t.Fatal("expected a listen error for an unassignable address")
	}
	if !strings.Contains(err.Error(), "192.0.2.1:8080") {
		t.Errorf("expected the error to name host and port, got %v", err)
	}
}
