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
	"gopkg.in/yaml.v3"

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
	viper.BindPFlag("web.readonly", webStartCmd.Flags().Lookup("readonly"))
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
		Port:           port,
		ScanDir:        absDir,
		Dev:            webDev,
		Verbose:        flags.Verbose,
		ReadOnly:       viper.GetBool("web.readonly"),
		Version:        FullVersion(),
		Phases:         parsePhasesForWeb(),
		Efforts:        resolveEffortScale(),
		ListProjects:   buildListProjects(),
		ResolveProject: buildResolveProject(),
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

// parsePhasesForWeb reads phases from viper config and converts to web.PhaseInfo.
func parsePhasesForWeb() []web.PhaseInfo {
	raw := viper.Get("phases")
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	phases := make([]web.PhaseInfo, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		p := web.PhaseInfo{}
		if id, ok := m["id"].(string); ok {
			p.ID = id
		}
		if name, ok := m["name"].(string); ok {
			p.Name = name
		}
		if desc, ok := m["description"].(string); ok {
			p.Description = desc
		}
		if p.ID != "" {
			phases = append(phases, p)
		}
	}
	return phases
}

// buildListProjects returns a function that loads the global project registry
// and converts entries to web.ProjectEntry values.
func buildListProjects() func() ([]web.ProjectEntry, error) {
	return func() ([]web.ProjectEntry, error) {
		entries, err := LoadGlobalRegistry()
		if err != nil {
			return nil, err
		}
		projects := make([]web.ProjectEntry, len(entries))
		for i, e := range entries {
			projects[i] = web.ProjectEntry{ID: e.ID, Name: e.Name, Path: e.Path}
		}
		return projects, nil
	}
}

// buildResolveProject returns a function that resolves a project ID to its
// scan directory and phases, reading directly from the project's .taskmd.yaml.
func buildResolveProject() web.ProjectResolverFunc {
	return func(id string) (string, []web.PhaseInfo, error) {
		entries, err := LoadGlobalRegistry()
		if err != nil {
			return "", nil, fmt.Errorf("load global registry: %w", err)
		}

		entry, found := findProjectEntry(entries, id)
		if !found {
			return "", nil, web.ErrProjectNotFound
		}

		info, statErr := os.Stat(entry.Path)
		if statErr != nil || !info.IsDir() {
			return "", nil, fmt.Errorf("project path does not exist: %s", entry.Path)
		}

		scanDir, err := resolveProjectTaskDirStandalone(entry.Path)
		if err != nil {
			return "", nil, err
		}

		phases := loadProjectPhases(entry.Path)
		return scanDir, phases, nil
	}
}

// resolveProjectTaskDirStandalone reads a project's .taskmd.yaml without viper
// to determine the task directory.
func resolveProjectTaskDirStandalone(projectPath string) (string, error) {
	cfgPath := filepath.Join(projectPath, ".taskmd.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return projectPath, nil
		}
		return "", fmt.Errorf("read project config: %w", err)
	}

	var cfg struct {
		TaskDir string `yaml:"task-dir"`
		Dir     string `yaml:"dir"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse project config: %w", err)
	}

	if cfg.TaskDir != "" {
		return resolveRelativeTo(cfg.TaskDir, projectPath), nil
	}
	if cfg.Dir != "" {
		return resolveRelativeTo(cfg.Dir, projectPath), nil
	}
	return projectPath, nil
}

// resolveRelativeTo makes a path absolute relative to a base directory.
func resolveRelativeTo(path, base string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(base, path))
}

// loadProjectPhases reads phases from a project's .taskmd.yaml.
func loadProjectPhases(projectPath string) []web.PhaseInfo {
	cfgPath := filepath.Join(projectPath, ".taskmd.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil
	}

	var cfg struct {
		Phases []struct {
			ID          string `yaml:"id"`
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		} `yaml:"phases"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	phases := make([]web.PhaseInfo, 0, len(cfg.Phases))
	for _, p := range cfg.Phases {
		if p.ID != "" {
			phases = append(phases, web.PhaseInfo{ID: p.ID, Name: p.Name, Description: p.Description})
		}
	}
	return phases
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
