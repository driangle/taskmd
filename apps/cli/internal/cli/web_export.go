package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/web"
)

var (
	webExportOutput   string
	webExportBasePath string
)

var webExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the dashboard as a static site",
	Long: `Export the taskmd dashboard as a self-contained static site.

The exported site can be deployed to GitHub Pages, Netlify, S3, or any
static file host. All task data is pre-rendered as JSON files and the
frontend is patched to read from local files instead of an API server.

Examples:
  taskmd web export
  taskmd web export -o ./public
  taskmd web export --base-path /demo/
  taskmd web export --task-dir ./tasks -o ./site`,
	Args: cobra.NoArgs,
	RunE: runWebExport,
}

func init() {
	webCmd.AddCommand(webExportCmd)

	webExportCmd.Flags().StringVarP(&webExportOutput, "output", "o", "./taskmd-export", "output directory")
	webExportCmd.Flags().StringVar(&webExportBasePath, "base-path", "/", "base path for URLs (e.g. /demo/)")
}

func runWebExport(_ *cobra.Command, _ []string) error {
	flags := GetGlobalFlags()

	absDir, err := filepath.Abs(flags.TaskDir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("not a valid directory: %s", absDir)
	}

	absOutput, err := filepath.Abs(webExportOutput)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	return web.Export(web.ExportConfig{
		OutputDir: absOutput,
		ScanDir:   absDir,
		BasePath:  webExportBasePath,
		Verbose:   flags.Verbose,
		Version:   FullVersion(),
		Efforts:   resolveEffortScale(),
	})
}
