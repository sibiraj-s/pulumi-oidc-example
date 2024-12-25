package main

import (
	"os"
	"path/filepath"
)

// in real world, this funcion would fetch the token from the OIDC provider
func GetOIDCToken() string {
	// return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...."
	token := os.Getenv("OIDC_TOKEN")
	WriteToFile(TokenFilePath(), token)
	return token
}

func TokenFilePath() string {
	return filepath.Join(CurrentDir(), "token.txt")
}

func GetRoleArn() string {
	return os.Getenv("ROLE_ARN")
}
