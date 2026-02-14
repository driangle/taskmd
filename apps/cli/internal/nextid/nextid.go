package nextid

import "fmt"

// Result holds the computed next ID and related metadata.
type Result struct {
	NextID  string `json:"next_id" yaml:"next_id"`
	MaxID   string `json:"max_id" yaml:"max_id"`
	Prefix  string `json:"prefix" yaml:"prefix"`
	Padding int    `json:"padding" yaml:"padding"`
	Total   int    `json:"total" yaml:"total"`
}

type parsedID struct {
	original string
	prefix   string
	number   int
	numStr   string
}

// Calculate determines the next available ID from a list of existing IDs.
// It finds the maximum numeric suffix and returns max+1, preserving any
// common prefix and zero-padding.
func Calculate(ids []string) Result {
	var parsed []parsedID
	for _, id := range ids {
		if p, ok := parseID(id); ok {
			parsed = append(parsed, p)
		}
	}

	if len(parsed) == 0 {
		return Result{
			NextID:  "001",
			Padding: 3,
			Total:   len(ids),
		}
	}

	maxNum := 0
	maxParsed := parsed[0]
	for _, p := range parsed {
		if p.number > maxNum {
			maxNum = p.number
			maxParsed = p
		}
	}

	prefix := detectPrefix(parsed)
	padding := max(len(maxParsed.numStr), 3)

	nextNum := maxNum + 1
	nextID := formatID(prefix, nextNum, padding)

	return Result{
		NextID:  nextID,
		MaxID:   maxParsed.original,
		Prefix:  prefix,
		Padding: padding,
		Total:   len(ids),
	}
}

// parseID extracts the trailing numeric portion and any prefix from an ID.
// Returns false if the ID contains no digits at the end.
func parseID(id string) (parsedID, bool) {
	if id == "" {
		return parsedID{}, false
	}

	// Scan backward to find where trailing digits start
	i := len(id) - 1
	for i >= 0 && id[i] >= '0' && id[i] <= '9' {
		i--
	}

	numStr := id[i+1:]
	if numStr == "" {
		return parsedID{}, false
	}

	num := 0
	for _, ch := range numStr {
		num = num*10 + int(ch-'0')
	}

	return parsedID{
		original: id,
		prefix:   id[:i+1],
		number:   num,
		numStr:   numStr,
	}, true
}

// detectPrefix returns the most common prefix if it appears in more than
// 50% of the parsed IDs. Otherwise returns "".
func detectPrefix(parsed []parsedID) string {
	if len(parsed) == 0 {
		return ""
	}

	counts := make(map[string]int)
	for _, p := range parsed {
		counts[p.prefix]++
	}

	bestPrefix := ""
	bestCount := 0
	for prefix, count := range counts {
		if count > bestCount {
			bestCount = count
			bestPrefix = prefix
		}
	}

	if bestCount*2 > len(parsed) {
		return bestPrefix
	}
	return ""
}

// formatID assembles a prefix with a zero-padded number.
func formatID(prefix string, number int, padding int) string {
	return fmt.Sprintf("%s%0*d", prefix, padding, number)
}
