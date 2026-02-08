package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// outputSnapshotJSON outputs snapshot as JSON
func outputSnapshotJSON(output any, outFile *os.File) error {
	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// outputSnapshotYAML outputs snapshot as YAML
func outputSnapshotYAML(output any, outFile *os.File) error {
	encoder := yaml.NewEncoder(outFile)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(output)
}

// outputSnapshotMarkdown outputs snapshot as Markdown
func outputSnapshotMarkdown(snapshots []TaskSnapshot, outFile *os.File, groupBy string) error {
	if groupBy != "" {
		groups := groupSnapshots(snapshots, groupBy)

		// Sort group keys
		keys := make([]string, 0, len(groups))
		for key := range groups {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Output each group
		for _, key := range keys {
			// Capitalize first letter of key
			title := key
			if len(key) > 0 {
				title = strings.ToUpper(key[:1]) + key[1:]
			}
			fmt.Fprintf(outFile, "## %s\n\n", title)
			for _, snapshot := range groups[key] {
				outputSnapshotMarkdownTask(snapshot, outFile)
			}
			fmt.Fprintln(outFile)
		}
	} else {
		// Output all tasks
		for _, snapshot := range snapshots {
			outputSnapshotMarkdownTask(snapshot, outFile)
		}
	}

	return nil
}

// outputSnapshotMarkdownTask outputs a single task in markdown format
func outputSnapshotMarkdownTask(snapshot TaskSnapshot, outFile *os.File) {
	fmt.Fprintf(outFile, "### [%s] %s\n\n", snapshot.ID, snapshot.Title)

	if snapshot.Status != "" {
		fmt.Fprintf(outFile, "- **Status**: %s\n", snapshot.Status)
	}
	if snapshot.Priority != "" {
		fmt.Fprintf(outFile, "- **Priority**: %s\n", snapshot.Priority)
	}
	if snapshot.Effort != "" {
		fmt.Fprintf(outFile, "- **Effort**: %s\n", snapshot.Effort)
	}
	if len(snapshot.Dependencies) > 0 {
		fmt.Fprintf(outFile, "- **Dependencies**: %s\n", strings.Join(snapshot.Dependencies, ", "))
	}
	if len(snapshot.Tags) > 0 {
		fmt.Fprintf(outFile, "- **Tags**: %s\n", strings.Join(snapshot.Tags, ", "))
	}

	// Derived fields
	if snapshot.IsBlocked != nil {
		fmt.Fprintf(outFile, "- **Blocked**: %v\n", *snapshot.IsBlocked)
	}
	if snapshot.DependencyDepth != nil {
		fmt.Fprintf(outFile, "- **Depth**: %d\n", *snapshot.DependencyDepth)
	}
	if snapshot.TopologicalOrder != nil {
		fmt.Fprintf(outFile, "- **Topo Order**: %d\n", *snapshot.TopologicalOrder)
	}
	if snapshot.OnCriticalPath != nil && *snapshot.OnCriticalPath {
		fmt.Fprintf(outFile, "- **On Critical Path**: yes\n")
	}

	fmt.Fprintln(outFile)
}
