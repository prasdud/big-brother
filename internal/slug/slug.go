// Package slug generates stable, URL-safe identifiers from human names.
package slug

import "strings"

// Make converts a name into a slug: ASCII lowercase alphanumeric words
// separated by single hyphens. Non-ASCII characters act as separators.
func Make(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(name) {
		isAlnum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlnum {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
