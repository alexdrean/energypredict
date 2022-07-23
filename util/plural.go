package util

import (
	"strings"
)

func Plural(s string, n int) string {
	if n == 1 || n == -1 {
		return strings.ReplaceAll(s, "(s)", "s")
	} else {
		return strings.ReplaceAll(s, "(s)", "")
	}
}
