package utils

import (
	"os"
	"regexp"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func IsStringEmpty(text string) bool {
	return len(text) == 0
}

// Replace strings using regex
func ReplaceAll(text string, pattern string, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(text, replacement)
}
