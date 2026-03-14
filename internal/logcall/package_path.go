package logcall

import "strings"

func matchesPackagePath(actual, expected string) bool {
	return actual == expected || strings.HasSuffix(actual, "/vendor/"+expected)
}
