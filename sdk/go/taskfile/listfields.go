package taskfile

import "strings"

// listShape describes how a frontmatter list value is written on disk.
type listShape int

const (
	// shapeNone means the key is present but carries no list value we manage.
	shapeNone listShape = iota
	// shapeFlow is a flow sequence — ["a", "b"] — whose "[" may open on the key
	// line or, as Prettier emits once the line grows too long, on a later one.
	shapeFlow
	// shapeBlock is a block sequence of "- item" lines.
	shapeBlock
)

// emptyMode selects what an empty result does to the key.
type emptyMode int

const (
	// removeKey drops the key entirely — how pr, touches and dependencies behave.
	removeKey emptyMode = iota
	// keepEmptyKey leaves "key: []" behind — how tags behaves.
	keepEmptyKey
)

// listField locates a frontmatter list field and the full extent of its value.
//
// endIdx is exclusive, so lines[keyIdx:endIdx] is everything the field owns.
// Tracking the real extent is the point of this type: a flow sequence can span
// several lines, and rewriting only the key line leaves the rest behind as an
// orphan that corrupts the document.
type listField struct {
	keyIdx int
	endIdx int
	shape  listShape
	values []string
}

// findListField locates fieldName within the frontmatter and measures its value.
//
// The second return is false when the key is absent. A key with an unmanaged
// scalar value returns true with shapeNone, so callers can tell "not there" from
// "there but not a list".
func findListField(lines []string, openIdx, closeIdx int, fieldName string) (listField, bool) {
	prefix := fieldName + ":"
	keyIdx := -1
	for i := openIdx + 1; i < closeIdx; i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), prefix) {
			keyIdx = i
			break
		}
	}
	if keyIdx < 0 {
		return listField{}, false
	}

	f := listField{keyIdx: keyIdx, endIdx: keyIdx + 1, shape: shapeNone}
	rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[keyIdx]), prefix))

	// Flow sequence opening on the key line.
	if strings.HasPrefix(rest, "[") {
		end, inner := scanFlow(lines, keyIdx, closeIdx, strings.Index(lines[keyIdx], "["))
		f.endIdx, f.shape, f.values = end, shapeFlow, parseFlowValues(inner)
		return f, true
	}

	// Any other value on the key line is a scalar, not a list.
	if rest != "" {
		return f, true
	}

	// Flow sequence opening on a later line — the shape Prettier emits, and the
	// one that used to fall through to the block branch and corrupt the file.
	if keyIdx+1 < closeIdx {
		if next := lines[keyIdx+1]; strings.HasPrefix(strings.TrimSpace(next), "[") {
			end, inner := scanFlow(lines, keyIdx+1, closeIdx, strings.Index(next, "["))
			f.endIdx, f.shape, f.values = end, shapeFlow, parseFlowValues(inner)
			return f, true
		}
	}

	// Block sequence.
	end := keyIdx + 1
	var values []string
	for end < closeIdx && strings.HasPrefix(strings.TrimSpace(lines[end]), "- ") {
		item := strings.TrimPrefix(strings.TrimSpace(lines[end]), "- ")
		values = append(values, strings.Trim(strings.TrimSpace(item), `"'`))
		end++
	}
	if end > keyIdx+1 {
		f.endIdx, f.shape, f.values = end, shapeBlock, values
	}
	return f, true
}

// scanFlow walks a flow sequence from its opening "[" to the matching "]", which
// may sit several lines below. It returns the exclusive end line index and the
// raw text between the outermost brackets.
//
// Brackets inside quotes are ignored so a value containing "[" cannot end the
// scan early.
func scanFlow(lines []string, startLine, closeIdx, openCol int) (int, string) {
	var inner strings.Builder
	depth := 0
	var quote byte

	for i := startLine; i < closeIdx; i++ {
		line := lines[i]
		col := 0
		if i == startLine {
			col = openCol
		}
		for ; col < len(line); col++ {
			c := line[col]
			if quote != 0 {
				if c == quote {
					quote = 0
				}
				inner.WriteByte(c)
				continue
			}
			switch c {
			case '"', '\'':
				quote = c
				inner.WriteByte(c)
			case '[':
				depth++
				if depth > 1 {
					inner.WriteByte(c)
				}
			case ']':
				depth--
				if depth == 0 {
					return i + 1, inner.String()
				}
				inner.WriteByte(c)
			default:
				inner.WriteByte(c)
			}
		}
		// A line break inside a flow sequence separates tokens.
		inner.WriteByte(' ')
	}

	// Unterminated — treat the remainder of the frontmatter as the value rather
	// than silently keeping half of it.
	return closeIdx, inner.String()
}

// parseFlowValues splits the inside of a flow sequence into its entries,
// tolerating the trailing comma Prettier emits.
func parseFlowValues(inner string) []string {
	if strings.TrimSpace(inner) == "" {
		return nil
	}
	var values []string
	for _, part := range strings.Split(inner, ",") {
		part = strings.Trim(strings.TrimSpace(part), `"'`)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

// listFieldLines renders the replacement lines for a field, preserving the shape
// it already had so a block sequence stays a block sequence.
func listFieldLines(fieldName string, shape listShape, values []string, mode emptyMode) []string {
	if len(values) == 0 {
		if mode == keepEmptyKey {
			return []string{fieldName + ": []"}
		}
		return nil
	}
	if shape == shapeBlock {
		lines := make([]string, 0, len(values)+1)
		lines = append(lines, fieldName+":")
		for _, v := range values {
			lines = append(lines, "  - "+v)
		}
		return lines
	}
	return []string{FormatInlineList(fieldName, values)}
}

// replaceListField splices newLines over the field's full extent and returns the
// adjusted frontmatter close index.
func replaceListField(lines []string, closeIdx int, f listField, newLines []string) ([]string, int) {
	result := make([]string, 0, len(lines)-(f.endIdx-f.keyIdx)+len(newLines))
	result = append(result, lines[:f.keyIdx]...)
	result = append(result, newLines...)
	result = append(result, lines[f.endIdx:]...)
	return result, closeIdx + len(newLines) - (f.endIdx - f.keyIdx)
}
