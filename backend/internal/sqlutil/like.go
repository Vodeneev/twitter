package sqlutil

import "strings"

// ContainsPattern returns a case-insensitive ILIKE pattern for substring search.
func ContainsPattern(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)
	return "%" + q + "%"
}
