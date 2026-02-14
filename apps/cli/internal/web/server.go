package web

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/watcher"
)

// Config holds server configuration.
type Config struct {
	Port     int
	ScanDir  string
	Dev      bool
	Verbose  bool
	ReadOnly bool
}

// Server is the taskmd web server.
type Server struct {
	config  Config
	dp      *DataProvider
	broker  *SSEBroker
	watcher *watcher.Watcher
}

// NewServer creates a new web server.
func NewServer(cfg Config) *Server {
	dp := NewDataProvider(cfg.ScanDir, cfg.Verbose)
	broker := NewSSEBroker()

	w := watcher.New(cfg.ScanDir, func() {
		dp.Invalidate()
		broker.Broadcast()
	}, 200*time.Millisecond)

	return &Server{
		config:  cfg,
		dp:      dp,
		broker:  broker,
		watcher: w,
	}
}

// Start starts the HTTP server. It blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/config", handleConfig(s.config))
	mux.HandleFunc("GET /api/tasks", handleTasks(s.dp))
	mux.HandleFunc("GET /api/tasks/{id}", handleTaskByID(s.dp))
	mux.HandleFunc("PUT /api/tasks/{id}", handleUpdateTask(s.dp, s.config.ReadOnly))
	mux.HandleFunc("GET /api/board", handleBoard(s.dp))
	mux.HandleFunc("GET /api/graph", handleGraph(s.dp))
	mux.HandleFunc("GET /api/graph/mermaid", handleGraphMermaid(s.dp))
	mux.HandleFunc("GET /api/stats", handleStats(s.dp))
	mux.HandleFunc("GET /api/validate", handleValidate(s.dp))
	mux.Handle("GET /api/events", s.broker)

	// Static file serving
	s.mountStatic(mux)

	var handler http.Handler = mux
	if s.config.Dev {
		handler = corsMiddleware(handler)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: handler,
	}

	// Start file watcher in background
	go func() {
		if err := s.watcher.Start(); err != nil && s.config.Verbose {
			fmt.Printf("watcher error: %v\n", err)
		}
	}()

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		s.watcher.Stop()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.Port, err)
	}

	s.printBanner()

	if err := srv.Serve(listener); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) printBanner() {
	fmt.Printf("taskmd web server running at http://localhost:%d\n", s.config.Port)
	fmt.Printf("Watching %s for changes\n", s.config.ScanDir)
	if s.config.ReadOnly {
		fmt.Println("Read-only mode: editing is disabled")
	}
	if s.config.Dev {
		fmt.Println("Dev mode: CORS enabled for http://localhost:5173")
	}
}

func (s *Server) mountStatic(mux *http.ServeMux) {
	if s.config.Dev {
		return
	}

	staticFS, err := fs.Sub(StaticFiles(), "static/dist")
	if err != nil {
		s.mountFallback(mux)
		return
	}

	if _, err := staticFS.Open("index.html"); err != nil {
		s.mountFallback(mux)
		return
	}

	fileServer := http.FileServer(http.FS(staticFS))

	mux.HandleFunc("/{path...}", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			http.NotFound(w, r)
			return
		}

		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		f, err := staticFS.Open(path[1:])
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) mountFallback(mux *http.ServeMux) {
	mux.HandleFunc("/{path...}", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html>
<html><body style="font-family:system-ui;max-width:480px;margin:80px auto;text-align:center">
<h2>taskmd</h2>
<p>No web UI embedded in this build.</p>
<p>Rebuild with <code>make build-full</code> or use <code>--dev</code> mode with the Vite dev server.</p>
</body></html>`)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
