package common

import "os"

// FileExistsValidator checks if a file exists.
func FileExistsValidator(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
