package main

import (
	"encoding/json"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func adminHashPassword(password string) (string, error) {
	passwordBytes := []byte(password)
	// salt and hash the password:
	hash, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func adminSetPassword(b []byte) error {
	pr := PasswordReset{}
	err := json.Unmarshal(b, &pr)
	if err != nil {
		return err
	}

	if err = verifyPassword(pr.CurrentPassword); err != nil {
		return err
	}

	hashedPassword, err := adminHashPassword(pr.NewPassword)
	if err != nil {
		return err
	}

	return setKeyStoreValue("admin", hashedPassword)
}

func adminSetFriends(b []byte) error {
	friends := []string{}
	err := json.Unmarshal(b, &friends)
	if err != nil {
		return err
	}

	return errors.New("Not implemented (TODO: set friends)")
}

func adminSetPlayers(b []byte) error {
	return errors.New("Not implemented (TODO: set players)")
}

func adminClearCache(b []byte) error {
	if len(b) != 0 {
		return errors.New("No body expected")
	}

	return errors.New("Not implemented (TODO: clearCache)")
}

func verifyPassword(password string) error {
	hashedPassword, err := getKeyStoreValue("admin")
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// PasswordReset cis the request to reset the admin password
type PasswordReset struct {
	CurrentPassword string `json:"old"`
	NewPassword     string `json:"new"`
}
