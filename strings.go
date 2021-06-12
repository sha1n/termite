package termite

import (
	"fmt"
)

// TruncateString returns a string that is at most maxLen long.
// If s is longer than maxLen, it is trimmed to (maxLen - 2) and three dots are appended.
func TruncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		if maxLen > 1 {
			return fmt.Sprintf("%s..", s[:maxLen-2])
		}
		return ""
	}
	return s
}
