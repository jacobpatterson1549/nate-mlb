package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func handleAdminRequest(r *http.Request) error {
	action := r.FormValue("action")
	switch action {
	case "password":
		return setPassword(r)
	case "cache":
		return clearCache(r)
	case "friend-names":
		return updateFriendNames(r)
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

func updateFriendNames(r *http.Request) error {
	currentPassword := r.FormValue("currentPassword")
	if err := verifyPassword(currentPassword); err != nil {
		return err
	}

	friendCount := r.FormValue("friend-count")
	friendCountI, err := strconv.Atoi(friendCount)
	if err != nil {
		return fmt.Errorf("Expected number for friend-count, but got %q", friendCount)
	}
	friendNames := make([]string, friendCountI)
	for i := 0; i < friendCountI; i++ {
		friendNames[i] = r.FormValue(fmt.Sprintf("friend-name-%d", i))
	}

	return errors.New("TOOD: save friend names")
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
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
