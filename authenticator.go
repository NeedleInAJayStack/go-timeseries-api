package main

import (
	"fmt"
	"log"
)

// authenticator is an interface for validating usernames and passwords.
type authenticator interface {
	// authenticate returns true if the username and password are valid.
	authenticate(username string, password string) (bool, error)
}

// singleUserAuthenticator is a simple authenticator that only accepts a literal single username and password.
// In real applications, these should be injected by environment variables.
type singleUserAuthenticator struct {
	username string
	password string
}

func (a singleUserAuthenticator) authenticate(username string, password string) (bool, error) {
	if username != a.username {
		log.Printf("Invalid username: %s", username)
		return false, fmt.Errorf("Invalid username: %s", username)
	}
	if password != a.password {
		log.Printf("Invalid password")
		return false, fmt.Errorf("Invalid password")
	}
	return true, nil
}
