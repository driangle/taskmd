package nextid

import (
	"testing"
)

func TestParseID(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		wantOK bool
		prefix string
		number int
		numStr string
	}{
		{"pure numeric", "042", true, "", 42, "042"},
		{"prefixed", "WEB-001", true, "WEB-", 1, "001"},
		{"single digit", "5", true, "", 5, "5"},
		{"empty", "", false, "", 0, ""},
		{"no digits", "abc", false, "", 0, ""},
		{"digits then letters", "12abc", false, "", 0, ""},
		{"mixed prefix", "task-v2-007", true, "task-v2-", 7, "007"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ok := parseID(tt.id)
			if ok != tt.wantOK {
				t.Fatalf("parseID(%q) ok = %v, want %v", tt.id, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if p.prefix != tt.prefix {
				t.Errorf("prefix = %q, want %q", p.prefix, tt.prefix)
			}
			if p.number != tt.number {
				t.Errorf("number = %d, want %d", p.number, tt.number)
			}
			if p.numStr != tt.numStr {
				t.Errorf("numStr = %q, want %q", p.numStr, tt.numStr)
			}
		})
	}
}

func TestDetectPrefix(t *testing.T) {
	tests := []struct {
		name   string
		parsed []parsedID
		want   string
	}{
		{
			"all same prefix",
			[]parsedID{
				{prefix: "WEB-"},
				{prefix: "WEB-"},
				{prefix: "WEB-"},
			},
			"WEB-",
		},
		{
			"majority wins",
			[]parsedID{
				{prefix: "CLI-"},
				{prefix: "CLI-"},
				{prefix: "WEB-"},
			},
			"CLI-",
		},
		{
			"no majority",
			[]parsedID{
				{prefix: "A-"},
				{prefix: "B-"},
				{prefix: "C-"},
				{prefix: "D-"},
			},
			"",
		},
		{
			"empty prefixes majority",
			[]parsedID{
				{prefix: ""},
				{prefix: ""},
				{prefix: "WEB-"},
			},
			"",
		},
		{
			"empty input",
			nil,
			"",
		},
		{
			"exactly half not majority",
			[]parsedID{
				{prefix: "A-"},
				{prefix: "A-"},
				{prefix: "B-"},
				{prefix: "B-"},
			},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectPrefix(tt.parsed)
			if got != tt.want {
				t.Errorf("detectPrefix() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatID(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		number  int
		padding int
		want    string
	}{
		{"basic padding", "", 42, 3, "042"},
		{"prefix with padding", "WEB-", 1, 3, "WEB-001"},
		{"overflow", "", 1000, 3, "1000"},
		{"large padding", "", 5, 5, "00005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatID(tt.prefix, tt.number, tt.padding)
			if got != tt.want {
				t.Errorf("formatID(%q, %d, %d) = %q, want %q", tt.prefix, tt.number, tt.padding, got, tt.want)
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name    string
		ids     []string
		wantID  string
		wantMax string
		wantPfx string
		wantPad int
	}{
		{
			"pure numeric sequential",
			[]string{"001", "002", "003"},
			"004", "003", "", 3,
		},
		{
			"with gaps",
			[]string{"001", "005", "010"},
			"011", "010", "", 3,
		},
		{
			"prefixed IDs",
			[]string{"WEB-001", "WEB-002", "WEB-003"},
			"WEB-004", "WEB-003", "WEB-", 3,
		},
		{
			"mixed prefix no majority",
			[]string{"WEB-001", "CLI-002", "API-003", "DB-004"},
			"005", "DB-004", "", 3,
		},
		{
			"empty input",
			[]string{},
			"001", "", "", 3,
		},
		{
			"single task",
			[]string{"042"},
			"043", "042", "", 3,
		},
		{
			"padding overflow",
			[]string{"997", "998", "999"},
			"1000", "999", "", 3,
		},
		{
			"non-parseable IDs ignored",
			[]string{"abc", "def", "003"},
			"004", "003", "", 3,
		},
		{
			"all non-parseable",
			[]string{"abc", "def"},
			"001", "", "", 3,
		},
		{
			"wider padding preserved",
			[]string{"00001", "00002", "00003"},
			"00004", "00003", "", 5,
		},
		{
			"single digit IDs",
			[]string{"1", "2", "3"},
			"004", "3", "", 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Calculate(tt.ids)
			if result.NextID != tt.wantID {
				t.Errorf("NextID = %q, want %q", result.NextID, tt.wantID)
			}
			if result.MaxID != tt.wantMax {
				t.Errorf("MaxID = %q, want %q", result.MaxID, tt.wantMax)
			}
			if result.Prefix != tt.wantPfx {
				t.Errorf("Prefix = %q, want %q", result.Prefix, tt.wantPfx)
			}
			if result.Padding != tt.wantPad {
				t.Errorf("Padding = %d, want %d", result.Padding, tt.wantPad)
			}
		})
	}
}
