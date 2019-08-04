package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
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
	case "friends":
		return updateFriends(r)
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

func updateFriends(r *http.Request) error {
	currentPassword := r.FormValue("currentPassword")
	if err := verifyPassword(currentPassword); err != nil {
		return err
	}

	friends := []Friend{}
	re := regexp.MustCompile("^friend-([0-9]+)-display-order$")
	for k, v := range r.Form {
		if idMatches := re.FindStringSubmatch(k); len(idMatches) > 0 {
			friendIDi, err := strconv.Atoi(idMatches[1])
			if err != nil {
				return err
			}
			friendDisplayOrderI, err := strconv.Atoi(v[0])
			if err != nil {
				return err
			}
			friendName := r.Form.Get(fmt.Sprintf("friend-%d-name", friendIDi))
			friends = append(friends, Friend{
				id:           friendIDi,
				displayOrder: friendDisplayOrderI,
				name:         friendName,
			})
		}
	}

	err := setFriends(friends)
	if err == nil {
		err = setKeyStoreValue("etl", "")
	}
	return err
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
