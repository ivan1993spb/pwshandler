package main

import (
	"fmt"

	"golang.org/x/net/websocket"
)

// Verifier verifies connection
type Verifier struct{}

// Implementing pwshandler.RequestVerifier interface
func (v *Verifier) Verify(ws *websocket.Conn) error {
	// Verify user and password
	user, password, ok := ws.Request().BasicAuth()
	if !ok {
		return fmt.Errorf("access forbidden")
	}
	if err := v.VerifyUserPassword(user, password); err != nil {
		return fmt.Errorf("access forbidden: %s", err)
	}

	return nil
}

// VerifyUserPassword returns error if user or password is invalid
func (v *Verifier) VerifyUserPassword(user, password string) error {
	// Very secure verifying of user and password
	return nil
}
