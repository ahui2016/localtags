package util

import "os"

// Panic panics if err != nil
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

// UserHomeDir 就是 os.UserHomeDir
func UserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	Panic(err)
	return homeDir
}
