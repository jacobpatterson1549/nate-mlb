package main

import (
	"encoding/json"
	"errors"
)

func adminSetPassword(b []byte) error {
	pr := PasswordReset{}
	err := json.Unmarshal(b, &pr)
	if err != nil {
		return err
	}

	return errors.New("Not implemented (TODO: set password)")
}

func adminSetYears(b []byte) error {
	y := Years{}
	err := json.Unmarshal(b, &y)
	if err != nil {
		return err
	}

	return errors.New("Not implemented (TODO: set years)")
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

// PasswordReset cis the request to reset the admin password
type PasswordReset struct {
	CurrentPassword string `json:"old"`
	NewPassword     string `json:"new"`
}

// Years is the request to update the years for different pools
type Years struct {
	Years       []int `json:"years"`
	DisplayYear int   `json:"display"`
}
