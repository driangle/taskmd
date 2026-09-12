package parser

import "regexp"

// taskSignaturePattern matches line-anchored task-signature keys in raw
// frontmatter text. Matching is textual on purpose: it is used when the
// frontmatter failed to parse as YAML, so parsing it is not an option.
var taskSignaturePattern = regexp.MustCompile(`(?m)^\s*(id|status|priority|dependencies)\s*:`)

// HasTaskSignature reports whether raw frontmatter text looks like it was
// intended to be taskmd frontmatter, based on the presence of task-signature
// keys (id, status, priority, dependencies).
//
// This is a behavioral contract for error reporting: a file whose frontmatter
// is present but not valid YAML is reported as a broken task file only when it
// carries at least one of these keys. Foreign markdown with frontmatter (docs
// pages, blog posts) that lacks them is silently skipped, as before.
func HasTaskSignature(frontmatter []byte) bool {
	return taskSignaturePattern.Match(frontmatter)
}
