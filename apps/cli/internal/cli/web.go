package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/driangle/md-task-tracker/apps/cli/internal/web"
	"github.com/spf13/cobra"
)

var (
	webPort int
	webDir  string
	webDev  bool
	webOpen bool
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Web dashboard commands",
	Long:  `Commands for the taskmd web dashboard.`,
}

var webStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the web dashboard server",
	Long: `Start a local web server serving the taskmd dashboard.

The server provides:
  - A JSON API backed by the same packages as the CLI
  - A web UI for viewing tasks, boards, graphs, and stats
  - Live reload via Server-Sent Events when task files change

Examples:
  taskmd web start
  taskmd web start --dir ./tasks
  taskmd web start --port 3000
  taskmd web start --dev --port 8080 --dir ./tasks`,
	Args: cobra.NoArgs,
	RunE: runWebStart,
}

func init() {
	rootCmd.AddCommand(webCmd)
	webCmd.AddCommand(webStartCmd)

	webStartCmd.Flags().IntVar(&webPort, "port", 8080, "server port")
	webStartCmd.Flags().StringVar(&webDir, "dir", ".", "task directory to scan")
	webStartCmd.Flags().BoolVar(&webDev, "dev", false, "enable dev mode (CORS for Vite dev server)")
	webStartCmd.Flags().BoolVar(&webOpen, "open", false, "open browser on start")
}

func runWebStart(cmd *cobra.Command, _ []string) error {
	absDir, err := filepath.Abs(webDir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("not a valid directory: %s", absDir)
	}

	srv := web.NewServer(web.Config{
		Port:    webPort,
		ScanDir: absDir,
		Dev:     webDev,
		Verbose: verbose,
	})

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	if webOpen {
		go openBrowser(fmt.Sprintf("http://localhost:%d", webPort))
	}

	return srv.Start(ctx)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	cmd.Run()
}
