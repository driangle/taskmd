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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/driangle/taskmd/apps/cli/internal/web"
)

var (
	webPort     int
	webDev      bool
	webOpen     bool
	webReadOnly bool
)

var webCmd = &cobra.Command{
	Use:        "web",
	SuggestFor: []string{"serve", "server", "http"},
	Short:      "Web dashboard commands",
	Long:       `Commands for the taskmd web dashboard.`,
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
  taskmd web start --task-dir ./tasks
  taskmd web start --port 3000
  taskmd web start --dev --port 8080 --task-dir ./tasks`,
	Args: cobra.NoArgs,
	RunE: runWebStart,
}

func init() {
	rootCmd.AddCommand(webCmd)
	webCmd.AddCommand(webStartCmd)

	webStartCmd.Flags().IntVar(&webPort, "port", 8080, "server port")
	webStartCmd.Flags().BoolVar(&webDev, "dev", false, "enable dev mode (CORS for Vite dev server)")
	webStartCmd.Flags().BoolVar(&webOpen, "open", false, "open browser on start")
	webStartCmd.Flags().BoolVar(&webReadOnly, "readonly", false, "start in read-only mode (disables editing)")

	// Bind flags to viper for config file support
	viper.BindPFlag("web.port", webStartCmd.Flags().Lookup("port"))
	viper.BindPFlag("web.auto_open_browser", webStartCmd.Flags().Lookup("open"))
}

func runWebStart(cmd *cobra.Command, _ []string) error {
	absDir, err := filepath.Abs(GetGlobalFlags().TaskDir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("not a valid directory: %s", absDir)
	}

	// Read from viper to support config file values
	port := viper.GetInt("web.port")
	open := viper.GetBool("web.auto_open_browser")
	flags := GetGlobalFlags()

	srv := web.NewServer(web.Config{
		Port:     port,
		ScanDir:  absDir,
		Dev:      webDev,
		Verbose:  flags.Verbose,
		ReadOnly: webReadOnly,
		Version:  FullVersion(),
	})

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	if open {
		go openBrowser(fmt.Sprintf("http://localhost:%d", port))
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
	cmd.Run() //nolint:errcheck
}
