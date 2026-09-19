//go:build e2e

package e2e

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// freePort asks the OS for an unused port, then releases it. A server started
// on that port immediately afterwards is overwhelmingly likely to get it, and
// this avoids colliding with whatever else is running on the machine.
func freePort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find a free port: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

// webServer is a `taskmd web start` process under test.
type webServer struct {
	cmd    *exec.Cmd
	out    *syncBuffer
	port   int
	cancel func()
}

// syncBuffer is an io.Writer safe for concurrent use: the child's stdout is
// written by one goroutine while the test reads it from another.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// startWebServer launches `taskmd web start` and waits for its banner.
func startWebServer(t *testing.T, dir string, args ...string) *webServer {
	t.Helper()

	port := freePort(t)
	full := append([]string{"web", "start", "--port", fmt.Sprint(port)}, args...)

	cmd := exec.Command(binaryPath, full...)
	cmd.Dir = dir
	out := &syncBuffer{}
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start web server: %v", err)
	}

	srv := &webServer{cmd: cmd, out: out, port: port}
	t.Cleanup(srv.stop)

	// Wait for the banner rather than sleeping a fixed interval.
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), "taskmd web server running at") {
			return srv
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("web server did not print a banner within 15s; output:\n%s", out.String())
	return nil
}

func (s *webServer) stop() {
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}
}

func (s *webServer) output() string { return s.out.String() }

// get performs a GET against the given host and this server's port.
func (s *webServer) get(t *testing.T, host string) (int, error) {
	t.Helper()

	url := fmt.Sprintf("http://%s/api/tasks", net.JoinHostPort(host, fmt.Sprint(s.port)))
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func TestWebStart_DefaultBindsLoopbackAndSaysSo(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "001-alpha.md", "001", "Alpha Task", "pending", nil)

	srv := startWebServer(t, root)
	out := srv.output()

	// The banner must state the real bind address, not an unqualified
	// "localhost" that implies loopback without guaranteeing it (issue #23).
	if !strings.Contains(out, "http://127.0.0.1:") {
		t.Errorf("expected banner to show the loopback bind address, got:\n%s", out)
	}
	if strings.Contains(out, "WARNING") {
		t.Errorf("unexpected exposure warning for the default bind, got:\n%s", out)
	}

	status, err := srv.get(t, "127.0.0.1")
	if err != nil {
		t.Fatalf("expected the dashboard to serve on loopback: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("expected 200 from /api/tasks, got %d", status)
	}
}

// The core of issue #23: the default must not be reachable from a
// non-loopback address on this host.
func TestWebStart_DefaultIsNotReachableOffLoopback(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "001-alpha.md", "001", "Alpha Task", "pending", nil)

	ip := nonLoopbackIPv4(t)
	srv := startWebServer(t, root)

	if _, err := srv.get(t, ip); err == nil {
		t.Errorf("default bind answered on %s; it must be loopback-only", ip)
	}
}

func TestWebStart_WildcardWarnsAndIsReachable(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "001-alpha.md", "001", "Alpha Task", "pending", nil)

	ip := nonLoopbackIPv4(t)
	srv := startWebServer(t, root, "--host", "0.0.0.0")
	out := srv.output()

	if !strings.Contains(out, "Bound to 0.0.0.0") {
		t.Errorf("expected the banner to name the wildcard bind, got:\n%s", out)
	}
	if !strings.Contains(out, "WARNING") {
		t.Errorf("expected an exposure warning, got:\n%s", out)
	}
	if !strings.Contains(out, "PUT /api/tasks/{id}") {
		t.Errorf("expected the warning to name the write endpoint, got:\n%s", out)
	}

	status, err := srv.get(t, ip)
	if err != nil {
		t.Fatalf("expected --host 0.0.0.0 to serve on %s: %v", ip, err)
	}
	if status != http.StatusOK {
		t.Errorf("expected 200 from /api/tasks on %s, got %d", ip, status)
	}
}

func TestWebStart_ReadOnlyWildcardOmitsWriteWarning(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "001-alpha.md", "001", "Alpha Task", "pending", nil)

	srv := startWebServer(t, root, "--host", "0.0.0.0", "--readonly")
	out := srv.output()

	if !strings.Contains(out, "WARNING") {
		t.Errorf("expected an exposure warning even in read-only mode, got:\n%s", out)
	}
	if strings.Contains(out, "PUT /api/tasks/{id}") {
		t.Errorf("read-only mode must not warn about writes, got:\n%s", out)
	}
}

func TestWebStart_HostFromConfigFile(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "001-alpha.md", "001", "Alpha Task", "pending", nil)
	writeConfig(t, root, "web:\n  host: 0.0.0.0\n")

	srv := startWebServer(t, root)

	if !strings.Contains(srv.output(), "Bound to 0.0.0.0") {
		t.Errorf("expected web.host from .taskmd.yaml to take effect, got:\n%s", srv.output())
	}
}

func TestWebStart_UnassignableHostFailsWithAddress(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "001-alpha.md", "001", "Alpha Task", "pending", nil)

	// TEST-NET-1 (RFC 5737) is never assigned to a local interface.
	res := run(t, root, "web", "start", "--host", "192.0.2.1", "--port", "8080")
	if res.ExitCode == 0 {
		t.Fatal("expected a non-zero exit for an unassignable bind address")
	}

	combined := res.Stdout + res.Stderr
	if !strings.Contains(combined, "192.0.2.1:8080") {
		t.Errorf("expected the error to name host and port, got:\n%s", combined)
	}
}

// nonLoopbackIPv4 returns a routable IPv4 address of this host, skipping the
// test when there is none (an isolated CI container).
func nonLoopbackIPv4(t *testing.T) string {
	t.Helper()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Skipf("cannot enumerate interfaces: %v", err)
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if v4 := ipNet.IP.To4(); v4 != nil {
				return v4.String()
			}
		}
	}
	t.Skip("no non-loopback IPv4 address on this host")
	return ""
}
