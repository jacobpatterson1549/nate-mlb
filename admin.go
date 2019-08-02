package main

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func handleAdminRequest(r *http.Request) error {
	switch r.FormValue("action") {
	case "password":
		return setPassword(r)
	case "cache":
		return clearCache(r)
	default:
		return errors.New("invalid admin action")
	}
}

func setPassword(r *http.Request) error {
	newPassword := r.FormValue("newPassword")
	currentPassword := r.FormValue("currentPassword")

	if err := verifyPassword(currentPassword); err != nil {
		return err
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	return setKeyStoreValue("admin", hashedPassword)
}

func clearCache(r *http.Request) error {
	currentPassword := r.FormValue("currentPassword")
	if err := verifyPassword(currentPassword); err != nil {
		return err
	}

	return setKeyStoreValue("etl", "")
}

func hashPassword(password string) (string, error) {
	passwordBytes := []byte(password)
	hash, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
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
