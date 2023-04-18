package util

import "strings"

func GetKeywordAndContent(msg string) (string, string) {
	msg = trimSpace(msg)
	split := strings.SplitN(msg, " ", 2)
	if len(split) == 2 {
		return strings.ToLower(split[0]), trimSpace(split[1])
	}

	return strings.ToLower(msg), ""
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
