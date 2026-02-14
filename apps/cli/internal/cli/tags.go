package cli

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var (
	tagsFormat  string
	tagsFilters []string
)

// TagInfo holds a tag name and the number of tasks using it.
type TagInfo struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

var tagsCmd = &cobra.Command{
	Use:        "tags",
	SuggestFor: []string{"labels", "categories"},
	Short:      "List all tags with task counts",
	Long: `Tags displays all tags used across task files along with the number
of tasks per tag, sorted from most to least used.

By default, scans the current directory and all subdirectories for markdown files
with task frontmatter. You can specify a different directory to scan.

Output formats: table (default), json, yaml

Multiple --filter flags are combined with AND logic.

Examples:
  taskmd tags
  taskmd tags ./tasks
  taskmd tags --filter status=pending
  taskmd tags --format json
  taskmd tags --format yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTags,
}

func init() {
	rootCmd.AddCommand(tagsCmd)

	tagsCmd.Flags().StringVar(&tagsFormat, "format", "table", "output format (table, json, yaml)")
	tagsCmd.Flags().StringArrayVar(&tagsFilters, "filter", []string{}, "filter tasks before aggregating tags (e.g., --filter status=pending)")
}

func runTags(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	tasks := result.Tasks

	if flags.Verbose && len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarning: encountered %d errors during scan:\n", len(result.Errors))
		for _, scanErr := range result.Errors {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", scanErr.FilePath, scanErr.Error)
		}
		fmt.Fprintln(os.Stderr)
	}

	if len(tagsFilters) > 0 {
		tasks, err = applyFilters(tasks, tagsFilters)
		if err != nil {
			return fmt.Errorf("filter error: %w", err)
		}
	}

	tagInfos := aggregateTags(tasks)

	switch tagsFormat {
	case "json":
		return outputTagsJSON(tagInfos)
	case "yaml":
		return WriteYAML(os.Stdout, tagInfos)
	case "table":
		return outputTagsTable(tagInfos)
	default:
		return ValidateFormat(tagsFormat, []string{"table", "json", "yaml"})
	}
}

func aggregateTags(tasks []*model.Task) []TagInfo {
	counts := make(map[string]int)
	for _, task := range tasks {
		for _, tag := range task.Tags {
			counts[tag]++
		}
	}

	tagInfos := make([]TagInfo, 0, len(counts))
	for tag, count := range counts {
		tagInfos = append(tagInfos, TagInfo{Tag: tag, Count: count})
	}

	sort.Slice(tagInfos, func(i, j int) bool {
		if tagInfos[i].Count != tagInfos[j].Count {
			return tagInfos[i].Count > tagInfos[j].Count
		}
		return tagInfos[i].Tag < tagInfos[j].Tag
	})

	return tagInfos
}

func outputTagsJSON(tagInfos []TagInfo) error {
	return WriteJSON(os.Stdout, tagInfos)
}

func outputTagsTable(tagInfos []TagInfo) error {
	if len(tagInfos) == 0 {
		fmt.Println("No tags found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "TAG\tCOUNT")
	fmt.Fprintf(w, "%s\t%s\n", "----------", "----------")

	for _, ti := range tagInfos {
		fmt.Fprintf(w, "%s\t%d\n", ti.Tag, ti.Count)
	}

	return nil
}
