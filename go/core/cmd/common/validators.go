package common

import (
	"net/mail"
	"os"
)

// FileExistsValidator checks if a file exists.
func FileExistsValidator(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// EmailValidator checks if an email address is a valid format.
func EmailValidator(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
