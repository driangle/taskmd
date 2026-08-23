package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	projectRegisterID   string
	projectRegisterPath string
	projectRegisterName string

	projectUnregisterID string
)

var projectRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register current directory as a project",
	Long: `Register a project in the global registry (~/.taskmd.yaml).

By default, the current directory is registered with an ID derived from the
directory basename. The directory must contain a .taskmd.yaml file.

Examples:
  taskmd projects register
  taskmd projects register --id my-project
  taskmd projects register --path /path/to/project
  taskmd projects register --id my-project --name "My Project"`,
	Args: cobra.NoArgs,
	RunE: runProjectRegister,
}

var projectUnregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Remove a project from the global registry",
	Long: `Remove a project from the global registry (~/.taskmd.yaml).

By default, matches by the current directory path. Use --id to match by
project ID instead.

Examples:
  taskmd projects unregister
  taskmd projects unregister --id my-project`,
	Args: cobra.NoArgs,
	RunE: runProjectUnregister,
}

func init() {
	projectsCmd.AddCommand(projectRegisterCmd)
	projectsCmd.AddCommand(projectUnregisterCmd)

	projectRegisterCmd.Flags().StringVar(&projectRegisterID, "id", "", "project ID (default: directory basename)")
	projectRegisterCmd.Flags().StringVar(&projectRegisterPath, "path", "", "path to register (default: current directory)")
	projectRegisterCmd.Flags().StringVar(&projectRegisterName, "name", "", "display name (default: same as ID)")

	projectUnregisterCmd.Flags().StringVar(&projectUnregisterID, "id", "", "project ID to remove (default: match by current directory)")
}

func runProjectRegister(_ *cobra.Command, _ []string) error {
	targetPath, err := resolveRegisterPath()
	if err != nil {
		return err
	}
	// Project identity is repository identity (ADR 0005 §1): registering from
	// a linked worktree stores the primary worktree's path.
	targetPath = canonicalRegisterPath(targetPath)
	if err := validateProjectDir(targetPath); err != nil {
		return err
	}
	id, name := resolveRegisterIDAndName(targetPath)
	return registerProject(id, name, targetPath)
}

func resolveRegisterPath() (string, error) {
	p := projectRegisterPath
	if p == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot determine current directory: %w", err)
		}
		p = cwd
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("cannot resolve symlinks: %w", err)
	}
	return resolved, nil
}

func validateProjectDir(path string) error {
	configFile := filepath.Join(path, configFilename)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("no %s found in %s — initialize a project first with 'taskmd init'", configFilename, path)
	}
	return nil
}

func resolveRegisterIDAndName(targetPath string) (string, string) {
	id := projectRegisterID
	if id == "" {
		id = filepath.Base(targetPath)
	}
	name := projectRegisterName
	if name == "" {
		name = id
	}
	return id, name
}

func registerProject(id, name, targetPath string) error {
	configPath, err := globalConfigPath()
	if err != nil {
		return err
	}
	doc, err := readGlobalConfigNode(configPath)
	if err != nil {
		return fmt.Errorf("failed to read global config: %w", err)
	}
	if existing, ok := findRegisteredRepo(doc, targetPath, filepath.Dir(configPath)); ok {
		fmt.Printf("Already registered as %q (path: %s)\n", existing.ID, existing.Path)
		return nil
	}
	if err := checkDuplicateID(doc, id); err != nil {
		return err
	}
	appendProjectNode(doc, id, name, targetPath)
	if err := writeGlobalConfigNode(configPath, doc); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}
	fmt.Printf("Registered project %q at %s\n", id, targetPath)
	return nil
}

func runProjectUnregister(_ *cobra.Command, _ []string) error {
	configPath, err := globalConfigPath()
	if err != nil {
		return err
	}
	doc, err := readGlobalConfigNode(configPath)
	if err != nil {
		return fmt.Errorf("failed to read global config: %w", err)
	}
	matchKey, matchVal, err := resolveUnregisterMatch()
	if err != nil {
		return err
	}
	removedID, ok := removeProjectNode(doc, matchKey, matchVal)
	if !ok {
		return fmt.Errorf("no project found matching %s %q", matchKey, matchVal)
	}
	if err := writeGlobalConfigNode(configPath, doc); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}
	fmt.Printf("Unregistered project %q\n", removedID)
	return nil
}

func resolveUnregisterMatch() (string, string, error) {
	if projectUnregisterID != "" {
		return "id", projectUnregisterID, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("cannot determine current directory: %w", err)
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", "", fmt.Errorf("cannot resolve path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", "", fmt.Errorf("cannot resolve symlinks: %w", err)
	}
	return "path", resolved, nil
}

// readGlobalConfigNode reads the YAML document from the given path.
// Returns an empty document if the file does not exist.
func readGlobalConfigNode(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return emptyDocNode(), nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return emptyDocNode(), nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return &doc, nil
}

func emptyDocNode() *yaml.Node {
	return &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{{Kind: yaml.MappingNode}},
	}
}

// writeGlobalConfigNode writes the YAML document to the given path.
func writeGlobalConfigNode(path string, doc *yaml.Node) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return err
	}
	return enc.Close()
}

// findProjectsSequence locates the "projects" sequence node in the mapping.
// Returns nil if not found.
func findProjectsSequence(mapping *yaml.Node) *yaml.Node {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == "projects" {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// extractProjectEntries reads all project entries from the document.
func extractProjectEntries(doc *yaml.Node) []GlobalProjectEntry {
	if len(doc.Content) == 0 {
		return nil
	}
	seq := findProjectsSequence(doc.Content[0])
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return nil
	}
	var entries []GlobalProjectEntry
	for _, item := range seq.Content {
		entries = append(entries, nodeToProjectEntry(item))
	}
	return entries
}

func nodeToProjectEntry(node *yaml.Node) GlobalProjectEntry {
	var entry GlobalProjectEntry
	if node.Kind != yaml.MappingNode {
		return entry
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		switch node.Content[i].Value {
		case "id":
			entry.ID = node.Content[i+1].Value
		case "name":
			entry.Name = node.Content[i+1].Value
		case "path":
			entry.Path = node.Content[i+1].Value
		}
	}
	return entry
}

// findRegisteredRepo reports whether targetPath's repository is already
// registered, matching by repo identity so any worktree of a registered repo
// is recognized (ADR 0005 §1). Non-git targets never match — non-git projects
// keep today's ID-only duplicate handling.
func findRegisteredRepo(doc *yaml.Node, targetPath, cfgDir string) (GlobalProjectEntry, bool) {
	target := resolveRepoIdentity(targetPath)
	if target == nil {
		return GlobalProjectEntry{}, false
	}
	for _, e := range extractProjectEntries(doc) {
		entryPath, err := resolvePath(e.Path, cfgDir)
		if err != nil {
			continue
		}
		if entryPath == targetPath || sameRepo(target, resolveRepoIdentity(entryPath)) {
			return e, true
		}
	}
	return GlobalProjectEntry{}, false
}

func checkDuplicateID(doc *yaml.Node, id string) error {
	entries := extractProjectEntries(doc)
	for _, e := range entries {
		if e.ID == id {
			return fmt.Errorf("project with ID %q already exists (path: %s)", id, e.Path)
		}
	}
	return nil
}

// appendProjectNode adds a new project entry to the document's projects list.
func appendProjectNode(doc *yaml.Node, id, name, path string) {
	mapping := doc.Content[0]
	seq := findProjectsSequence(mapping)
	if seq == nil {
		// Create the "projects" key and sequence
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "projects", Tag: "!!str"}
		seq = &yaml.Node{Kind: yaml.SequenceNode}
		mapping.Content = append(mapping.Content, keyNode, seq)
	}
	seq.Content = append(seq.Content, buildProjectNode(id, name, path))
}

func buildProjectNode(id, name, path string) *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "id", Tag: "!!str"},
			{Kind: yaml.ScalarNode, Value: id, Tag: "!!str"},
			{Kind: yaml.ScalarNode, Value: "name", Tag: "!!str"},
			{Kind: yaml.ScalarNode, Value: name, Tag: "!!str"},
			{Kind: yaml.ScalarNode, Value: "path", Tag: "!!str"},
			{Kind: yaml.ScalarNode, Value: path, Tag: "!!str"},
		},
	}
}

// removeProjectNode removes a project entry matching key=val from the projects sequence.
// Returns the removed project's ID and true, or "" and false if not found.
func removeProjectNode(doc *yaml.Node, key, val string) (string, bool) {
	if len(doc.Content) == 0 {
		return "", false
	}
	seq := findProjectsSequence(doc.Content[0])
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return "", false
	}
	for i, item := range seq.Content {
		entry := nodeToProjectEntry(item)
		if matchesEntry(entry, key, val) {
			seq.Content = append(seq.Content[:i], seq.Content[i+1:]...)
			return entry.ID, true
		}
	}
	return "", false
}

func matchesEntry(entry GlobalProjectEntry, key, val string) bool {
	switch key {
	case "id":
		return entry.ID == val
	case "path":
		return entry.Path == val
	}
	return false
}
