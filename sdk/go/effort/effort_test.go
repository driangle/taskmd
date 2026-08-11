package effort

import (
	"strings"
	"testing"
)

func TestNewScale_Valid(t *testing.T) {
	s, err := NewScale([]string{"xs", "s", "m", "l", "xl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := s.String(); got != "xs, s, m, l, xl" {
		t.Errorf("String() = %q, want %q", got, "xs, s, m, l, xl")
	}
	if got := s.Len(); got != 5 {
		t.Errorf("Len() = %d, want 5", got)
	}
	if got := s.Lowest(); got != "xs" {
		t.Errorf("Lowest() = %q, want %q", got, "xs")
	}
	if got := s.Highest(); got != "xl" {
		t.Errorf("Highest() = %q, want %q", got, "xl")
	}
}

func TestNewScale_TrimsWhitespace(t *testing.T) {
	s, err := NewScale([]string{" xs ", "s"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Contains("xs") {
		t.Errorf("Contains(%q) = false, want true (value should be trimmed)", "xs")
	}
}

func TestNewScale_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		wantErr string
	}{
		{"empty list", []string{}, "at least one value"},
		{"nil list", nil, "at least one value"},
		{"blank entry", []string{"small", ""}, "index 1 is empty"},
		{"whitespace entry", []string{"small", "   "}, "index 1 is empty"},
		{"duplicate", []string{"small", "large", "small"}, `duplicate effort value: "small"`},
		{"duplicate after trim", []string{"small", " small "}, `duplicate effort value: "small"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewScale(tt.values)
			if err == nil {
				t.Fatalf("expected an error for %v", tt.values)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// The zero value must behave as the default vocabulary so callers with no
// configuration can pass Scale{} safely.
func TestScale_ZeroValueIsDefault(t *testing.T) {
	var zero Scale

	if got, want := zero.String(), "small, medium, large"; got != want {
		t.Errorf("zero Scale String() = %q, want %q", got, want)
	}
	if got, want := zero.String(), Default().String(); got != want {
		t.Errorf("zero Scale = %q, Default() = %q; want them equal", got, want)
	}
	if got := zero.Lowest(); got != "small" {
		t.Errorf("Lowest() = %q, want %q", got, "small")
	}
}

func TestScale_Rank(t *testing.T) {
	s, err := NewScale([]string{"xs", "s", "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		value string
		want  int
	}{
		{"xs", 0},
		{"s", 1},
		{"m", 2},
		{"small", -1}, // default vocabulary must not leak through
		{"", -1},
	}

	for _, tt := range tests {
		if got := s.Rank(tt.value); got != tt.want {
			t.Errorf("Rank(%q) = %d, want %d", tt.value, got, tt.want)
		}
		if got, want := s.Contains(tt.value), tt.want >= 0; got != want {
			t.Errorf("Contains(%q) = %v, want %v", tt.value, got, want)
		}
	}
}

// Points spreads maxPoints across the vocabulary by position. The default
// three-value case must reproduce the historical 5 / 2 / 0 scoring exactly.
func TestScale_Points(t *testing.T) {
	tests := []struct {
		name       string
		values     []string
		value      string
		wantPoints int
		wantOK     bool
	}{
		{"default lowest", nil, "small", 5, true},
		{"default middle", nil, "medium", 2, true},
		{"default highest", nil, "large", 0, false},
		{"default unset", nil, "", 0, false},
		{"default unknown", nil, "xs", 0, false},

		{"five-value rank 0", []string{"xs", "s", "m", "l", "xl"}, "xs", 5, true},
		{"five-value rank 1", []string{"xs", "s", "m", "l", "xl"}, "s", 3, true},
		{"five-value rank 2", []string{"xs", "s", "m", "l", "xl"}, "m", 2, true},
		{"five-value rank 3", []string{"xs", "s", "m", "l", "xl"}, "l", 1, true},
		{"five-value rank 4", []string{"xs", "s", "m", "l", "xl"}, "xl", 0, false},

		{"two-value lowest", []string{"s", "l"}, "s", 5, true},
		{"two-value highest", []string{"s", "l"}, "l", 0, false},

		{"single value scores full", []string{"only"}, "only", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Default()
			if tt.values != nil {
				var err error
				s, err = NewScale(tt.values)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			points, ok := s.Points(tt.value, 5)
			if points != tt.wantPoints || ok != tt.wantOK {
				t.Errorf("Points(%q, 5) = (%d, %v), want (%d, %v)",
					tt.value, points, ok, tt.wantPoints, tt.wantOK)
			}
		})
	}
}

// Points must never exceed maxPoints or go negative for any vocabulary size.
func TestScale_PointsWithinBounds(t *testing.T) {
	for size := 1; size <= 12; size++ {
		values := make([]string, size)
		for i := range values {
			values[i] = string(rune('a' + i))
		}
		s, err := NewScale(values)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, v := range values {
			points, _ := s.Points(v, 5)
			if points < 0 || points > 5 {
				t.Errorf("size %d: Points(%q, 5) = %d, want within [0, 5]", size, v, points)
			}
		}
		if points, _ := s.Points(s.Lowest(), 5); points != 5 {
			t.Errorf("size %d: lowest value scored %d, want 5", size, points)
		}
	}
}

func TestScale_ValuesIsOrdered(t *testing.T) {
	want := []string{"xs", "s", "m"}
	s, err := NewScale(want)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := s.Values()
	if len(got) != len(want) {
		t.Fatalf("Values() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Values()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
