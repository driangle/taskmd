package cli

import _ "embed"

//go:embed templates/CLAUDE.md
var claudeTemplate []byte

//go:embed templates/GEMINI.md
var geminiTemplate []byte

//go:embed templates/CODEX.md
var codexTemplate []byte

type agentConfig struct {
	name     string
	filename string
	template []byte
}
