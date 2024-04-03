package utils

import (
	"os"
	"strings"
)

// Contains проверяет: входит ли строка в массив строк?
func Contains(arr []string, str string, caseInsensitive bool) bool {
	lowerStr := strings.ToLower(str)
	for _, v := range arr {
		if (v == str) || (caseInsensitive && (strings.ToLower(v) == lowerStr)) {
			return true
		}
	}
	return false
}

// IsFileExists проверяет: существует ли файл?
func IsFileExists(path string) bool {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !fi.IsDir()
}
