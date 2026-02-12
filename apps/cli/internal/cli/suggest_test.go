package cli

import (
	"testing"
)

// commandSuggestion maps an alias to the expected command name.
type commandSuggestion struct {
	alias   string
	command string
}

func TestSuggestFor_AllAliasesRegistered(t *testing.T) {
	tests := []commandSuggestion{
		// update
		{alias: "set", command: "update"},
		{alias: "edit", command: "update"},
		{alias: "modify", command: "update"},
		{alias: "change", command: "update"},
		// show
		{alias: "view", command: "show"},
		{alias: "info", command: "show"},
		{alias: "detail", command: "show"},
		{alias: "details", command: "show"},
		{alias: "describe", command: "show"},
		{alias: "get", command: "show"},
		// list
		{alias: "ls", command: "list"},
		{alias: "tasks", command: "list"},
		{alias: "all", command: "list"},
		// graph
		{alias: "deps", command: "graph"},
		{alias: "dependencies", command: "graph"},
		{alias: "tree", command: "graph"},
		// stats
		{alias: "summary", command: "stats"},
		{alias: "status", command: "stats"},
		{alias: "overview", command: "stats"},
		// validate
		{alias: "check", command: "validate"},
		{alias: "verify", command: "validate"},
		{alias: "lint", command: "validate"},
		// board
		{alias: "kanban", command: "board"},
		{alias: "columns", command: "board"},
		// next
		{alias: "pick", command: "next"},
		{alias: "suggest", command: "next"},
		{alias: "what", command: "next"},
		// snapshot
		{alias: "save", command: "snapshot"},
		{alias: "backup", command: "snapshot"},
		{alias: "export", command: "snapshot"},
		// init
		{alias: "setup", command: "init"},
		{alias: "create", command: "init"},
		{alias: "new", command: "init"},
		// tui
		{alias: "ui", command: "tui"},
		{alias: "interactive", command: "tui"},
		{alias: "dashboard", command: "tui"},
		// web
		{alias: "serve", command: "web"},
		{alias: "server", command: "web"},
		{alias: "http", command: "web"},
	}

	// Build a lookup: alias -> expected command name, from the actual registered commands.
	aliasMap := make(map[string]string)
	for _, cmd := range rootCmd.Commands() {
		for _, alias := range cmd.SuggestFor {
			aliasMap[alias] = cmd.Name()
		}
	}

	for _, tt := range tests {
		t.Run(tt.alias+"->"+tt.command, func(t *testing.T) {
			got, ok := aliasMap[tt.alias]
			if !ok {
				t.Fatalf("alias %q is not registered in any command's SuggestFor", tt.alias)
			}
			if got != tt.command {
				t.Fatalf("alias %q: expected command %q, got %q", tt.alias, tt.command, got)
			}
		})
	}
}

func TestSuggestFor_CobraSuggestsCorrectCommand(t *testing.T) {
	// Verify that cobra actually suggests the right command when given an alias.
	tests := []commandSuggestion{
		{alias: "set", command: "update"},
		{alias: "view", command: "show"},
		{alias: "ls", command: "list"},
		{alias: "deps", command: "graph"},
		{alias: "summary", command: "stats"},
		{alias: "check", command: "validate"},
		{alias: "kanban", command: "board"},
		{alias: "pick", command: "next"},
		{alias: "backup", command: "snapshot"},
		{alias: "setup", command: "init"},
		{alias: "ui", command: "tui"},
		{alias: "serve", command: "web"},
	}

	for _, tt := range tests {
		t.Run(tt.alias+"->"+tt.command, func(t *testing.T) {
			suggestions := rootCmd.SuggestionsFor(tt.alias)
			if len(suggestions) == 0 {
				t.Fatalf("rootCmd.SuggestionsFor(%q) returned no suggestions, expected %q", tt.alias, tt.command)
			}
			found := false
			for _, s := range suggestions {
				if s == tt.command {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("rootCmd.SuggestionsFor(%q) = %v, expected to contain %q", tt.alias, suggestions, tt.command)
			}
		})
	}
}

func TestSuggestFor_NoDuplicateAliases(t *testing.T) {
	// Ensure no alias is claimed by more than one command.
	seen := make(map[string]string) // alias -> command name
	for _, cmd := range rootCmd.Commands() {
		for _, alias := range cmd.SuggestFor {
			if prev, ok := seen[alias]; ok {
				t.Errorf("alias %q registered by both %q and %q", alias, prev, cmd.Name())
			}
			seen[alias] = cmd.Name()
		}
	}
}
