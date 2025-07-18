package main

import (
	"os"
)

// in real world, this funcion would fetch the token from the OIDC provider
func GetOIDCToken() string {
	// return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...."
	token := os.Getenv("OIDC_TOKEN")
	return token
}

func GetRoleArn() string {
	return os.Getenv("ROLE_ARN")
}
