package nextid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
	"time"
)

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

// GeneratePrefixed produces the next sequential ID with the given prefix.
// It filters existing IDs by prefix, finds the max numeric suffix, and
// returns prefix + zero-padded(max+1).
func GeneratePrefixed(existingIDs []string, prefix string, padding int) string {
	maxNum := 0
	for _, id := range existingIDs {
		p, ok := parseID(id)
		if !ok || !strings.EqualFold(p.prefix, prefix) {
			continue
		}
		if p.number > maxNum {
			maxNum = p.number
		}
	}
	return formatID(prefix, maxNum+1, padding)
}

// GenerateRandom produces a random base-36 alphanumeric lowercase ID of the
// given length. It retries on collision with existingIDs (max 100 attempts).
func GenerateRandom(existingIDs []string, length int) (string, error) {
	existing := make(map[string]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		existing[id] = struct{}{}
	}

	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	charsetLen := big.NewInt(int64(len(charset)))

	for attempt := 0; attempt < 100; attempt++ {
		buf := make([]byte, length)
		for i := range buf {
			idx, err := rand.Int(rand.Reader, charsetLen)
			if err != nil {
				return "", fmt.Errorf("crypto/rand failed: %w", err)
			}
			buf[i] = charset[idx.Int64()]
		}
		id := string(buf)
		if _, taken := existing[id]; !taken {
			return id, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique ID after 100 attempts")
}

// crockfordBase32 is the Crockford Base32 alphabet (lowercase).
const crockfordBase32 = "0123456789abcdefghjkmnpqrstvwxyz"

const (
	// ulidTimestampChars is how many characters encode the 48-bit timestamp.
	ulidTimestampChars = 10
	// ulidMinRandomChars is how many random characters a shortened ULID keeps.
	// Each character carries 5 bits, so 4 chars ≈ 1M values per timestamp bucket —
	// enough for collision retries to succeed instead of regenerating the same string.
	ulidMinRandomChars = 4
	// ulidMaxAttempts bounds the collision-retry loop.
	ulidMaxAttempts = 100
	// ulidRetryBudget bounds the total time spent waiting for the timestamp
	// component to advance between colliding attempts.
	ulidRetryBudget = 500 * time.Millisecond
)

// GenerateULID produces a ULID: 48-bit millisecond timestamp + 80-bit crypto random,
// encoded as 26 Crockford Base32 characters (lowercase). When length > 0 and < 26,
// the result is shortened to that length by dropping low-order timestamp characters
// and trailing random characters, always keeping some randomness. It retries on
// collision with existingIDs (max 100 attempts).
func GenerateULID(existingIDs []string, length int) (string, error) {
	if length <= 0 || length > 26 {
		length = 26
	}

	existing := make(map[string]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		existing[id] = struct{}{}
	}

	// A shortened ULID only changes over time once its coarsest timestamp
	// character advances, so waiting between colliding attempts is what actually
	// widens the search space. Capped so a saturated ID space still fails fast.
	tick := ulidTimestampTick(length)
	budget := ulidRetryBudget

	for attempt := 0; attempt < ulidMaxAttempts; attempt++ {
		id, err := encodeULID(length)
		if err != nil {
			return "", err
		}
		if _, taken := existing[id]; !taken {
			return id, nil
		}
		if budget > 0 {
			wait := min(tick, budget)
			time.Sleep(wait)
			budget -= wait
		}
	}

	return "", fmt.Errorf("failed to generate unique ULID of length %d after %d attempts "+
		"(%d existing IDs); increase id.length in your config", length, ulidMaxAttempts, len(existingIDs))
}

// ulidTimestampWidth returns how many leading timestamp characters a ULID of the
// given length keeps. Shortening always keeps the leading timestamp characters so
// IDs stay roughly time-ordered, but never at the cost of all randomness.
func ulidTimestampWidth(length int) int {
	if length >= ulidTimestampChars+ulidMinRandomChars {
		return ulidTimestampChars
	}
	return max(1, length-ulidMinRandomChars)
}

// ulidTimestampTick returns how long the timestamp portion of a ULID of the given
// length stays constant. Each dropped timestamp character costs 5 bits of
// millisecond resolution.
func ulidTimestampTick(length int) time.Duration {
	dropped := ulidTimestampChars - ulidTimestampWidth(length)
	ms := int64(1) << (5 * dropped)
	// Clamp: with many dropped characters the tick is hours, which would overflow
	// a Duration and is far beyond anything worth waiting for anyway.
	if ms > int64(ulidRetryBudget/time.Millisecond) {
		return ulidRetryBudget
	}
	return time.Duration(ms) * time.Millisecond
}

// encodeULID generates a single ULID string of the given length.
func encodeULID(length int) (string, error) {
	ms := uint64(time.Now().UnixMilli())

	// Read 10 bytes of randomness (80 bits)
	var randomBytes [10]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", fmt.Errorf("crypto/rand failed: %w", err)
	}

	// Build the full 26-char ULID:
	// - 10 chars for 48-bit timestamp
	// - 16 chars for 80-bit random
	var buf [26]byte

	// Encode timestamp (48 bits) into 10 Crockford Base32 chars (big-endian)
	buf[0] = crockfordBase32[(ms>>45)&0x1F]
	buf[1] = crockfordBase32[(ms>>40)&0x1F]
	buf[2] = crockfordBase32[(ms>>35)&0x1F]
	buf[3] = crockfordBase32[(ms>>30)&0x1F]
	buf[4] = crockfordBase32[(ms>>25)&0x1F]
	buf[5] = crockfordBase32[(ms>>20)&0x1F]
	buf[6] = crockfordBase32[(ms>>15)&0x1F]
	buf[7] = crockfordBase32[(ms>>10)&0x1F]
	buf[8] = crockfordBase32[(ms>>5)&0x1F]
	buf[9] = crockfordBase32[ms&0x1F]

	// Encode random (80 bits = 10 bytes) into 16 Crockford Base32 chars
	// Pack into a uint64 + uint16 for bit manipulation
	hi := binary.BigEndian.Uint64(randomBytes[0:8]) // top 64 bits
	lo := uint64(binary.BigEndian.Uint16(randomBytes[8:10]))

	// 80 bits → 16 base32 chars (16 * 5 = 80 bits exactly)
	// Combine into a single 80-bit value spread across hi (bits 79-16) and lo (bits 15-0)
	buf[10] = crockfordBase32[(hi>>59)&0x1F]
	buf[11] = crockfordBase32[(hi>>54)&0x1F]
	buf[12] = crockfordBase32[(hi>>49)&0x1F]
	buf[13] = crockfordBase32[(hi>>44)&0x1F]
	buf[14] = crockfordBase32[(hi>>39)&0x1F]
	buf[15] = crockfordBase32[(hi>>34)&0x1F]
	buf[16] = crockfordBase32[(hi>>29)&0x1F]
	buf[17] = crockfordBase32[(hi>>24)&0x1F]
	buf[18] = crockfordBase32[(hi>>19)&0x1F]
	buf[19] = crockfordBase32[(hi>>14)&0x1F]
	buf[20] = crockfordBase32[(hi>>9)&0x1F]
	buf[21] = crockfordBase32[(hi>>4)&0x1F]
	// Last 4 bits of hi + first bit of lo
	buf[22] = crockfordBase32[((hi&0x0F)<<1)|(lo>>15)]
	buf[23] = crockfordBase32[(lo>>10)&0x1F]
	buf[24] = crockfordBase32[(lo>>5)&0x1F]
	buf[25] = crockfordBase32[lo&0x1F]

	tsWidth := ulidTimestampWidth(length)
	return string(buf[:tsWidth]) + string(buf[ulidTimestampChars:ulidTimestampChars+length-tsWidth]), nil
}

// formatID assembles a prefix with a zero-padded number.
func formatID(prefix string, number int, padding int) string {
	return fmt.Sprintf("%s%0*d", prefix, padding, number)
}
