package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func CurrentDir() string {
	currentDir, _ := os.Getwd()
	return currentDir
}

func EnsureDir(path ...string) error {
	return os.MkdirAll(filepath.Join(path...), os.ModePerm)
}

func StringPtr(s string) *string {
	return &s
}

// only use if you want to exit the program
func CheckErrX(err error, message string) {
	if err == nil {
		return
	}

	fmt.Println(message+":", err)
	os.Exit(1)
}
