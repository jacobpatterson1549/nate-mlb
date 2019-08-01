package main

import (
	"errors"
	"net/http"

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

func adminSetPassword(r *http.Request) error {
	newPassword := r.FormValue("newPassword")
	currentPassword := r.FormValue("currentPassword")

	if err := verifyPassword(currentPassword); err != nil {
		return err
	}

	hashedPassword, err := adminHashPassword(newPassword)
	if err != nil {
		return err
	}

	return setKeyStoreValue("admin", hashedPassword)
}

func verifyPassword(password string) error {
	hashedPassword, err := getKeyStoreValue("admin")
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return errors.New("Incorrect Password")
	}
	return err
}
