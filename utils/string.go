package utils

import "strings"

func ReplaceLastString(text string, old string, newText string) string {
	i := strings.LastIndex(text, old)
	return text[:i] + strings.Replace(text[i:], old, newText, 1)
}
