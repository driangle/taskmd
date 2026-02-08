package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestInputResolver_ResolveInput(t *testing.T) {
	// Create a temporary directory and file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	testContent := "# Test Content\n"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		useStdin bool
		args     []string
		wantErr  bool
	}{
		{
			name:     "explicit file",
			useStdin: false,
			args:     []string{testFile},
			wantErr:  false,
		},
		{
			name:     "file not found",
			useStdin: false,
			args:     []string{filepath.Join(tmpDir, "nonexistent.md")},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewInputResolver(tt.useStdin, false)
			reader, cleanup, err := resolver.ResolveInput(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				defer cleanup()

				// Try to read content
				content, err := io.ReadAll(reader)
				if err != nil {
					t.Errorf("Failed to read from resolver: %v", err)
				}

				if !tt.useStdin && string(content) != testContent {
					t.Errorf("Content mismatch: got %q, want %q", string(content), testContent)
				}
			}
		})
	}
}

func TestInputResolver_ReadAll(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	testContent := []byte("# Test Content\nLine 2\n")

	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resolver := NewInputResolver(false, false)
	content, err := resolver.ReadAll([]string{testFile})

	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("ReadAll() got %q, want %q", string(content), string(testContent))
	}
}
