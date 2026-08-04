package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
	"github.com/driangle/taskmd/sdk/go/validator"
)

var (
	validateFormat string
	validateStrict bool
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:        "validate",
	SuggestFor: []string{"check", "verify", "lint"},
	Short:      "Lint and validate tasks",
	Long: `Validate checks task files and the .taskmd.yaml config file for errors.

Task validation checks:
  - Required fields (id, title)
  - Invalid field values (status, priority, effort)
  - Duplicate task IDs
  - Missing dependencies (references to non-existent tasks)
  - Circular dependencies (cycles in dependency graph)

Config validation checks (when .taskmd.yaml is present):
  - Scope definitions have non-empty paths arrays
  - No unknown top-level config keys
  - Task touches reference defined scopes

Use --strict to enable additional warnings for missing optional fields.

Output formats: text (default), table, json

Exit codes:
  0 - Valid (no errors)
  1 - Invalid (errors found)
  2 - Valid but with warnings (only in strict mode)

Examples:
  taskmd validate
  taskmd validate ./tasks
  taskmd validate --strict
  taskmd validate --format json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVar(&validateFormat, "format", "text", "output format (text, table, json)")
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "enable strict validation with additional warnings")
}

func runValidate(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	// Create scanner and scan for tasks
	taskScanner := newTaskScanner(scanDir, flags)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	tasks := result.Tasks

	// Report scan errors if any
	if len(result.Errors) > 0 {
		if !flags.Quiet {
			fmt.Fprintf(os.Stderr, "Warning: encountered %d errors during scan:\n", len(result.Errors))
			for _, scanErr := range result.Errors {
				fmt.Fprintf(os.Stderr, "  %s: %v\n", scanErr.FilePath, scanErr.Error)
			}
			fmt.Fprintln(os.Stderr)
		}
	}

	// Scan archive directories for task IDs to avoid false-positive dependency errors
	v := validator.NewValidator(validateStrict)
	if externalIDs := collectArchivedIDs(taskScanner); len(externalIDs) > 0 {
		v.SetExternalIDs(externalIDs)
	}

	// Run validation
	validationResult := v.Validate(tasks)
	validateConfig(v, validationResult, tasks)

	// Output results
	switch validateFormat {
	case "json":
		if err := outputValidationJSON(validationResult); err != nil {
			return err
		}
	case "text", "table":
		outputValidationText(validationResult, flags.Quiet)
	default:
		return ValidateFormat(validateFormat, []string{"text", "table", "json"})
	}

	// Determine exit code
	if !validationResult.IsValid() {
		os.Exit(ExitError)
	} else if validateStrict && validationResult.HasWarnings() {
		os.Exit(ExitValidationWarning)
	}

	return nil
}

// outputValidationText outputs validation results in human-readable text format
func outputValidationText(result *validator.ValidationResult, quiet bool) {
	r := getRenderer()

	if len(result.Issues) == 0 {
		if !quiet {
			fmt.Printf("%s All %d task(s) are valid\n", formatSuccess("✓", r), result.TaskCount)
		}
		return
	}

	// Group issues by level
	errors := []validator.ValidationIssue{}
	warnings := []validator.ValidationIssue{}

	for _, issue := range result.Issues {
		if issue.Level == validator.LevelError {
			errors = append(errors, issue)
		} else if issue.Level == validator.LevelWarning {
			warnings = append(warnings, issue)
		}
	}

	// Print errors
	if len(errors) > 0 {
		fmt.Printf("\n%s Found %d error(s):\n\n", formatError("❌", r), len(errors))
		for _, issue := range errors {
			printIssue(issue, r)
		}
	}

	// Print warnings
	if len(warnings) > 0 {
		fmt.Printf("\n%s  Found %d warning(s):\n\n", formatWarning("⚠️", r), len(warnings))
		for _, issue := range warnings {
			printIssue(issue, r)
		}
	}

	// Print summary
	fmt.Println()
	if result.Errors > 0 {
		fmt.Printf("Validated %d task(s): %s", result.TaskCount,
			formatError(fmt.Sprintf("%d error(s)", result.Errors), r))
		if result.Warnings > 0 {
			fmt.Printf(", %s", formatWarning(fmt.Sprintf("%d warning(s)", result.Warnings), r))
		}
		fmt.Println()
	} else if result.Warnings > 0 {
		fmt.Printf("Validated %d task(s) with %s\n", result.TaskCount,
			formatWarning(fmt.Sprintf("%d warning(s)", result.Warnings), r))
	}
}

// printIssue prints a single validation issue
func printIssue(issue validator.ValidationIssue, r *lipgloss.Renderer) {
	if issue.TaskID != "" {
		fmt.Printf("  [%s] %s\n", formatTaskID(issue.TaskID, r), issue.Message)
	} else {
		fmt.Printf("  %s\n", issue.Message)
	}

	if issue.FilePath != "" {
		fmt.Printf("    %s %s\n", formatLabel("File:", r), formatDim(issue.FilePath, r))
	}
}

// outputValidationJSON outputs validation results as JSON
func outputValidationJSON(result *validator.ValidationResult) error {
	return WriteJSON(os.Stdout, result)
}

// validateConfig runs config and cross-validation checks, merging results into validationResult.
func validateConfig(v *validator.Validator, validationResult *validator.ValidationResult, tasks []*model.Task) {
	configData := loadConfigForValidation()
	mergeValidationResults(validationResult, v.ValidateConfig(configData))

	if configData != nil && len(configData.Scopes) > 0 {
		knownScopes := make(map[string]bool, len(configData.Scopes))
		for name := range configData.Scopes {
			knownScopes[name] = true
		}
		mergeValidationResults(validationResult, v.ValidateTouchesAgainstScopes(tasks, knownScopes))
	}

	if configData != nil && len(configData.Phases) > 0 {
		knownPhases := make(map[string]bool, len(configData.Phases))
		for _, m := range configData.Phases {
			key := m.ID
			if key == "" {
				key = m.Name // backwards compat: fall back to name if no id
			}
			if key != "" {
				knownPhases[key] = true
			}
		}
		mergeValidationResults(validationResult, v.ValidatePhasesAgainstConfig(tasks, knownPhases))
	}
}

// loadConfigForValidation extracts config data from viper for validation.
// Returns nil if no config file was loaded.
func loadConfigForValidation() *validator.ConfigData {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		return nil
	}

	// Only include keys actually present in the config file, not pflag-bound defaults
	allSettings := viper.AllSettings()
	topKeys := make([]string, 0, len(allSettings))
	for k := range allSettings {
		if viper.InConfig(k) {
			topKeys = append(topKeys, k)
		}
	}
	sort.Strings(topKeys)

	config := &validator.ConfigData{
		TopKeys:    topKeys,
		ConfigPath: configPath,
		Workflow:   viper.GetString("workflow"),
	}

	raw := viper.Get("scopes")
	if raw != nil {
		if scopeMap, ok := raw.(map[string]any); ok {
			config.Scopes = parseScopeEntries(scopeMap)
		}
	}

	if viper.InConfig("id") {
		config.ID = parseIDConfig(viper.Get("id"))
	}

	if viper.InConfig("phases") {
		config.Phases = parsePhasesConfig(viper.Get("phases"))
	}

	return config
}

// parseIDConfig converts raw viper id data into a typed IDConfig.
// Returns nil if raw is nil or not a map.
func parseIDConfig(raw any) *validator.IDConfig {
	if raw == nil {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}

	cfg := &validator.IDConfig{}

	if s, ok := m["strategy"].(string); ok {
		cfg.Strategy = s
	}
	if s, ok := m["prefix"].(string); ok {
		cfg.Prefix = s
	}
	cfg.Length = toInt(m["length"])
	cfg.Padding = toInt(m["padding"])

	return cfg
}

// parsePhasesConfig converts raw viper phases data into typed PhaseConfig entries.
func parsePhasesConfig(raw any) []validator.PhaseConfig {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}

	phases := make([]validator.PhaseConfig, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		mc, valid := parsePhaseEntry(m)
		if valid {
			phases = append(phases, mc)
		}
	}
	return phases
}

func parsePhaseEntry(m map[string]any) (validator.PhaseConfig, bool) {
	mc := validator.PhaseConfig{}
	if id, ok := m["id"].(string); ok {
		mc.ID = id
	}
	if name, ok := m["name"].(string); ok {
		mc.Name = name
	}
	if desc, ok := m["description"].(string); ok {
		mc.Description = desc
	}
	if due, ok := m["due"].(string); ok {
		ft := model.FlexibleTime{}
		node := yaml.Node{Kind: yaml.ScalarNode, Value: due, Tag: "!!str"}
		if err := ft.UnmarshalYAML(&node); err == nil {
			mc.Due = ft
		}
	}
	return mc, true
}

// toInt converts a viper numeric value (int, int64, float64) to int.
func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// parseScopeEntries converts raw viper scope data into typed ScopeConfig entries.
func parseScopeEntries(scopeMap map[string]any) map[string]validator.ScopeConfig {
	scopes := make(map[string]validator.ScopeConfig, len(scopeMap))
	for name, val := range scopeMap {
		sc := validator.ScopeConfig{} // Paths stays nil if not found

		entryMap, ok := val.(map[string]any)
		if !ok {
			scopes[name] = sc
			continue
		}

		if desc, ok := entryMap["description"].(string); ok {
			sc.Description = desc
		}

		pathsRaw, exists := entryMap["paths"]
		if !exists {
			scopes[name] = sc
			continue
		}

		pathsSlice, ok := pathsRaw.([]any)
		if !ok {
			scopes[name] = sc
			continue
		}

		paths := make([]string, 0, len(pathsSlice))
		for _, p := range pathsSlice {
			if s, ok := p.(string); ok {
				paths = append(paths, s)
			}
		}
		sc.Paths = paths
		scopes[name] = sc
	}
	return scopes
}

// collectArchivedIDs scans archive directories and returns a set of task IDs found there.
func collectArchivedIDs(s *scanner.Scanner) map[string]bool {
	archivedTasks, err := s.ScanArchive()
	if err != nil || len(archivedTasks) == 0 {
		return nil
	}
	ids := make(map[string]bool, len(archivedTasks))
	for _, t := range archivedTasks {
		if t.ID != "" {
			ids[t.ID] = true
		}
	}
	return ids
}

// mergeValidationResults appends issues from source into target and updates counters.
func mergeValidationResults(target, source *validator.ValidationResult) {
	target.Issues = append(target.Issues, source.Issues...)
	target.Errors += source.Errors
	target.Warnings += source.Warnings
}
