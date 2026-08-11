package cli

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestParseEffortConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		raw        any
		wantValues []string
		wantErrSub string
	}{
		{
			name:       "absent key",
			raw:        nil,
			wantValues: nil,
		},
		{
			name:       "ordered list",
			raw:        []any{"xs", "s", "m", "l", "xl"},
			wantValues: []string{"xs", "s", "m", "l", "xl"},
		},
		{
			name:       "empty list is passed through for the validator to reject",
			raw:        []any{},
			wantValues: []string{},
		},
		{
			// The map form is the shape a user reaches for after reading about
			// `id:` or `scopes:`, so the error has to point at the list form.
			name:       "mapping instead of list",
			raw:        map[string]any{"values": []any{"xs", "s"}},
			wantErrSub: "must be a list of effort values",
		},
		{
			name:       "scalar instead of list",
			raw:        "small",
			wantErrSub: "must be a list of effort values",
		},
		{
			name:       "non-string item",
			raw:        []any{"xs", 3},
			wantValues: []string{"xs"},
			wantErrSub: "index 1 must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			values, errs := parseEffortConfig(tt.raw)

			assertEffortErrors(t, errs, tt.wantErrSub)
			assertEffortValues(t, values, tt.wantValues)
		})
	}
}

// assertEffortErrors checks the structural errors against an expected substring;
// an empty want means no errors are expected.
func assertEffortErrors(t *testing.T, errs []string, want string) {
	t.Helper()

	if want == "" {
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		return
	}
	if len(errs) == 0 {
		t.Fatalf("expected an error containing %q, got none", want)
	}
	if !strings.Contains(strings.Join(errs, "; "), want) {
		t.Errorf("errors = %v, want one containing %q", errs, want)
	}
}

// assertEffortValues compares parsed values against the expected slice, where a
// nil want means the key was absent and values must be nil too.
func assertEffortValues(t *testing.T, got, want []string) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Errorf("values = %v, want nil", got)
		}
		return
	}
	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("values[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolveEffortScale(t *testing.T) {
	tests := []struct {
		name string
		set  func()
		want string
	}{
		{
			name: "unset falls back to the default",
			set:  func() {},
			want: "small, medium, large",
		},
		{
			name: "configured vocabulary is used in order",
			set:  func() { viper.Set("effort", []any{"xs", "s", "m"}) },
			want: "xs, s, m",
		},
		{
			// Invalid config must not break every command; validate reports it.
			name: "duplicate values degrade to the default",
			set:  func() { viper.Set("effort", []any{"a", "a"}) },
			want: "small, medium, large",
		},
		{
			name: "empty list degrades to the default",
			set:  func() { viper.Set("effort", []any{}) },
			want: "small, medium, large",
		},
		{
			name: "wrong container degrades to the default",
			set:  func() { viper.Set("effort", map[string]any{"values": []any{"a"}}) },
			want: "small, medium, large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCLIState()
			defer resetCLIState()

			tt.set()

			if got := resolveEffortScale().String(); got != tt.want {
				t.Errorf("resolveEffortScale() = %q, want %q", got, tt.want)
			}
		})
	}
}
