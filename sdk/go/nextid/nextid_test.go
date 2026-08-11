package nextid

import (
	"regexp"
	"sort"
	"testing"
	"time"
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

func TestGeneratePrefixed(t *testing.T) {
	tests := []struct {
		name    string
		ids     []string
		prefix  string
		padding int
		want    string
	}{
		{
			"next after existing prefixed IDs",
			[]string{"WEB-001", "WEB-002", "WEB-003"},
			"WEB-", 3,
			"WEB-004",
		},
		{
			"empty existing IDs starts at 1",
			nil,
			"CLI-", 3,
			"CLI-001",
		},
		{
			"ignores IDs with different prefix",
			[]string{"WEB-010", "CLI-002", "CLI-005"},
			"CLI-", 3,
			"CLI-006",
		},
		{
			"respects padding",
			[]string{"T-01", "T-02"},
			"T-", 4,
			"T-0003",
		},
		{
			"case-insensitive prefix matching",
			[]string{"web-003"},
			"WEB-", 3,
			"WEB-004",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GeneratePrefixed(tt.ids, tt.prefix, tt.padding)
			if got != tt.want {
				t.Errorf("GeneratePrefixed() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateRandom_Format(t *testing.T) {
	id, err := GenerateRandom(nil, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(id) != 6 {
		t.Errorf("length = %d, want 6", len(id))
	}
	if !regexp.MustCompile(`^[0-9a-z]+$`).MatchString(id) {
		t.Errorf("id %q contains non-base36 characters", id)
	}
}

func TestGenerateRandom_RespectsLength(t *testing.T) {
	for _, length := range []int{4, 8, 12} {
		id, err := GenerateRandom(nil, length)
		if err != nil {
			t.Fatalf("unexpected error for length %d: %v", length, err)
		}
		if len(id) != length {
			t.Errorf("length = %d, want %d", len(id), length)
		}
	}
}

func TestGenerateRandom_CollisionAvoidance(t *testing.T) {
	existing := []string{"abc123", "def456", "ghi789"}
	for range 50 {
		id, err := GenerateRandom(existing, 6)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, ex := range existing {
			if id == ex {
				t.Errorf("generated id %q collides with existing", id)
			}
		}
	}
}

func TestGenerateRandom_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{})
	for range 100 {
		id, err := GenerateRandom(nil, 8)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, exists := seen[id]; exists {
			t.Errorf("duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestGenerateULID_Format(t *testing.T) {
	id, err := GenerateULID(nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(id) != 26 {
		t.Errorf("length = %d, want 26", len(id))
	}
	// Crockford Base32 lowercase: no i, l, o, u
	if !regexp.MustCompile(`^[0-9a-hjkmnp-tv-z]+$`).MatchString(id) {
		t.Errorf("id %q contains invalid Crockford Base32 characters", id)
	}
}

func TestGenerateULID_Sortability(t *testing.T) {
	id1, err := GenerateULID(nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	id2, err := GenerateULID(nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id1 >= id2 {
		t.Errorf("expected id1 < id2 for sortability, got %q >= %q", id1, id2)
	}
}

func TestGenerateULID_CollisionAvoidance(t *testing.T) {
	// Generate one ID, then verify new IDs don't collide
	first, err := GenerateULID(nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	existing := []string{first}
	for range 50 {
		id, err := GenerateULID(existing, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == first {
			t.Errorf("generated id %q collides with existing", id)
		}
	}
}

func TestGenerateULID_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{})
	for range 100 {
		id, err := GenerateULID(nil, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, exists := seen[id]; exists {
			t.Errorf("duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestGenerateULID_RespectsLength(t *testing.T) {
	for _, length := range []int{8, 12, 20, 26} {
		id, err := GenerateULID(nil, length)
		if err != nil {
			t.Fatalf("unexpected error for length %d: %v", length, err)
		}
		if len(id) != length {
			t.Errorf("length = %d, want %d", len(id), length)
		}
	}
}

func TestGenerateULID_OrderedBatch(t *testing.T) {
	// Generate a batch with small delays and verify they sort correctly
	var ids []string
	for range 10 {
		id, err := GenerateULID(nil, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ids = append(ids, id)
		time.Sleep(time.Millisecond)
	}

	sorted := make([]string, len(ids))
	copy(sorted, ids)
	sort.Strings(sorted)

	for i := range ids {
		if ids[i] != sorted[i] {
			t.Errorf("IDs not in sorted order at index %d: generated %q, sorted %q", i, ids[i], sorted[i])
		}
	}
}

// TestGenerateULID_ShortLengthBatch guards the bug where shortened ULIDs were
// pure timestamp: generating several IDs back-to-back at a short length used to
// exhaust all retries on identical strings.
func TestGenerateULID_ShortLengthBatch(t *testing.T) {
	for _, length := range []int{5, 8, 9, 12} {
		var ids []string
		for i := range 20 {
			id, err := GenerateULID(ids, length)
			if err != nil {
				t.Fatalf("length %d, iteration %d: unexpected error: %v", length, i, err)
			}
			if len(id) != length {
				t.Fatalf("length %d: got id %q of length %d", length, id, len(id))
			}
			ids = append(ids, id)
		}
	}
}

func TestGenerateULID_ShortLengthHasRandomness(t *testing.T) {
	// At length 9 the last characters must vary within a single timestamp tick.
	seen := make(map[string]struct{})
	for range 50 {
		id, err := GenerateULID(nil, 9)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seen[id] = struct{}{}
	}
	if len(seen) < 40 {
		t.Errorf("expected mostly distinct IDs, got %d unique out of 50", len(seen))
	}
}

func TestGenerateULID_ShortLengthKeepsTimePrefix(t *testing.T) {
	// The leading characters still encode time, so IDs stay coarsely ordered.
	a, err := GenerateULID(nil, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := GenerateULID(nil, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a[:ulidTimestampWidth(9)] != b[:ulidTimestampWidth(9)] {
		t.Errorf("timestamp prefixes differ within the same tick: %q vs %q", a, b)
	}
}

func TestGenerateULID_ExhaustedSpaceFailsFast(t *testing.T) {
	// A saturated space must still return an error rather than sleeping forever.
	existing := []string{}
	for _, c := range crockfordBase32 {
		existing = append(existing, string(c))
	}
	start := time.Now()
	if _, err := GenerateULID(existing, 1); err == nil {
		t.Fatal("expected error for exhausted 1-character space")
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("took %v, expected fast failure", elapsed)
	}
}
