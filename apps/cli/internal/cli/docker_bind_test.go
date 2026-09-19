package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// `taskmd web start` binds 127.0.0.1 by default, which is correct on a host but
// unreachable through Docker's published-port DNAT. Both images must opt back in
// to a wildcard bind, and the CI smoke test must be able to notice when they do
// not. These guards are here because the regression they cover shipped once.

func readRepoFile(t *testing.T, parts ...string) string {
	t.Helper()
	repoRoot := filepath.Join("..", "..", "..", "..")
	path := filepath.Join(append([]string{repoRoot}, parts...)...)
	data, err := os.ReadFile(path) //nolint:gosec // test-only read of a repo-relative path
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

func TestDockerfiles_BindWebServerToWildcard(t *testing.T) {
	for _, name := range []string{"Dockerfile", "Dockerfile.release"} {
		t.Run(name, func(t *testing.T) {
			content := readRepoFile(t, name)
			if !strings.Contains(content, "ENV TASKMD_WEB_HOST=0.0.0.0") {
				t.Errorf("%s must set ENV TASKMD_WEB_HOST=0.0.0.0; otherwise "+
					"`docker run -p 8080:8080 ...` cannot reach the web server", name)
			}
		})
	}
}

func TestCIWorkflow_DockerSmokeTestUsesPipefail(t *testing.T) {
	content := readRepoFile(t, ".github", "workflows", "ci.yml")

	idx := strings.Index(content, "- name: Test Docker image")
	if idx < 0 {
		t.Fatal(`step "Test Docker image" not found in ci.yml; this guard needs updating`)
	}

	// Look at the step header only, up to its `run:` block.
	rest := content[idx:]
	runIdx := strings.Index(rest, "run: |")
	if runIdx < 0 {
		t.Fatal(`step "Test Docker image" has no run block`)
	}
	// Ignore comment lines so the explanatory comment above the directive
	// cannot satisfy this guard on its own.
	var directives []string
	for _, line := range strings.Split(rest[:runIdx], "\n") {
		if trimmed := strings.TrimSpace(line); !strings.HasPrefix(trimmed, "#") {
			directives = append(directives, trimmed)
		}
	}
	header := strings.Join(directives, "\n")

	if !strings.Contains(header, "shell: bash") {
		t.Error(`the "Test Docker image" step must declare "shell: bash". ` +
			`Actions' default (bash -e {0}) has no pipefail, so ` +
			`"curl -sf ... | head" takes its exit status from head and the ` +
			`assertion can never fail.`)
	}
}
