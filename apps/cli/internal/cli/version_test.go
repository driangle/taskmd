package cli

import "testing"

func TestFullVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		commit   string
		dirty    string
		expected string
	}{
		{
			name:     "release build",
			version:  "1.0.0",
			commit:   "unknown",
			dirty:    "",
			expected: "1.0.0",
		},
		{
			name:     "dev build with commit",
			version:  "0.0.3",
			commit:   "abc1234",
			dirty:    "",
			expected: "0.0.3-abc1234",
		},
		{
			name:     "dev build with dirty tree",
			version:  "0.0.3",
			commit:   "abc1234",
			dirty:    "true",
			expected: "0.0.3-abc1234*",
		},
		{
			name:     "dev build with long commit hash",
			version:  "0.0.3",
			commit:   "abc1234567890",
			dirty:    "false",
			expected: "0.0.3-abc1234",
		},
		{
			name:     "empty commit",
			version:  "0.0.3",
			commit:   "",
			dirty:    "",
			expected: "0.0.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origVersion := Version
			origCommit := GitCommit
			origDirty := GitDirty
			defer func() {
				Version = origVersion
				GitCommit = origCommit
				GitDirty = origDirty
			}()

			Version = tt.version
			GitCommit = tt.commit
			GitDirty = tt.dirty

			got := FullVersion()
			if got != tt.expected {
				t.Errorf("FullVersion() = %q, want %q", got, tt.expected)
			}
		})
	}
}
