package tableprinter

import "strings"

// join slice of string but with line width limit and return slice of joined strings
func JoinWithLimit(elems []string, sep string, limit int) []string {
	var result []string
	currentLine := ""

	for _, keyword := range elems {
		if len(currentLine)+len(keyword)+len(sep) > limit {
			if currentLine != "" {
				result = append(result, currentLine)
			}
			currentLine = keyword
		} else {
			if currentLine != "" {
				currentLine += sep
			}
			currentLine += keyword
		}
	}

	if currentLine != "" {
		result = append(result, currentLine)
	}

	return result
}

// convert []string to []any for use in fmt.Printf
func convertToAny(slice []string) []any {
	result := make([]any, len(slice))
	for i, v := range slice {
		result[i] = v
	}

	return result
}

// replcae special chars to not printable equals
func escapeSpecialChars(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\b", "\\b")
	s = strings.ReplaceAll(s, "\f", "\\f")
	s = strings.ReplaceAll(s, "\v", "\\v")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\000", "\\0")
	return s
}
