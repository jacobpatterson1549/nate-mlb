package server

import (
	"errors"
	"fmt"
	"nate-mlb/internal/db"
	"net/http"
	"regexp"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

var (
	adminActions = map[string](func(*http.Request) error){
		"friends":  updateFriends,
		"players":  updatePlayers,
		"years":    updateYears,
		"cache":    clearCache,
		"password": resetPassword,
	}
)

func handleAdminRequest(r *http.Request) error {
	actionParam := r.FormValue("action")
	if action, ok := adminActions[actionParam]; ok {
		return action(r)
	}
	return errors.New("invalid admin action")
}

func updatePlayers(r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}

	var players []db.Player
	re := regexp.MustCompile("^player-([0-9]+)-display-order$")
	for k, v := range r.Form {
		if matches := re.FindStringSubmatch(k); len(matches) > 1 {
			player, err := getPlayer(r, matches[1], v[0])
			if err != nil {
				return err
			}
			players = append(players, player)
		}
	}

	err := db.SavePlayers(players)
	if err != nil {
		return err
	}
	return db.ClearEtlStats()
}

func updateFriends(r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}

	var friends []db.Friend
	re := regexp.MustCompile("^friend-([0-9]+)-display-order$")
	for k, v := range r.Form {
		if matches := re.FindStringSubmatch(k); len(matches) > 1 {
			friend, err := getFriend(r, matches[1], v[0])
			if err != nil {
				return err
			}
			friends = append(friends, friend)
		}
	}

	err := db.SaveFriends(friends)
	if err != nil {
		return err
	}
	return db.ClearEtlStats()
}

func updateYears(r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}

	var years []db.Year
	for _, y := range r.Form["year"] {
		year, err := getYear(r, y)
		if err != nil {
			return err
		}
		years = append(years, year)
	}

	// (does not forcefully update cache if active year changed)
	return db.SaveYears(years)
}

func clearCache(r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}

	return db.ClearEtlStats()
}

func resetPassword(r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}

	newPassword := r.FormValue("newPassword")
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	return db.SavePassword("admin", hashedPassword)
}

func verifyUserPassword(r *http.Request) error {
	username := r.FormValue("username")
	password := r.FormValue("password")
	hashedPassword, err := db.GetPassword(username)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return errors.New("Incorrect Password")
	}
	if err != nil {
		return fmt.Errorf("problem verifying password: %v", err)
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("problem hashing password: %v", err)
	}
	return string(hashedPassword), nil
}

func getPlayer(r *http.Request, id, displayOrder string) (db.Player, error) {
	var player db.Player

	IDI, err := strconv.Atoi(id)
	if err != nil {
		return player, fmt.Errorf("problem converting %v to number: %v", id, err)
	}
	player.ID = IDI

	displayOrderI, err := strconv.Atoi(displayOrder)
	if err != nil {
		return player, fmt.Errorf("problem converting player display order '%v' to number: %v", displayOrder, err)
	}
	player.DisplayOrder = displayOrderI

	playerTypeID := r.Form.Get(fmt.Sprintf("player-%s-player-type-id", id))
	if len(playerTypeID) > 0 { // not specified when updating existing player
		playerTypeIDI, err := strconv.Atoi(playerTypeID)
		if err != nil {
			return player, fmt.Errorf("problem converting player type id '%v' to number: %v", playerTypeID, err)
		}
		player.PlayerType = db.PlayerType(playerTypeIDI)
	}

	playerID := r.Form.Get(fmt.Sprintf("player-%s-player-id", id))
	playerIDI, err := strconv.Atoi(playerID)
	if err != nil {
		return player, fmt.Errorf("problem converting player id '%v' to number: %v", playerID, err)
	}
	player.PlayerID = playerIDI

	friendID := r.Form.Get(fmt.Sprintf("player-%s-friend-id", id))
	if len(friendID) > 0 { // not specified when updating existing player
		friendIDI, err := strconv.Atoi(friendID)
		if err != nil {
			return player, fmt.Errorf("problem converting player friend id '%v' to number: %v", friendID, err)
		}
		player.FriendID = friendIDI
	}

	return player, nil
}

func getFriend(r *http.Request, id, displayOrder string) (db.Friend, error) {
	var friend db.Friend

	friendIDI, err := strconv.Atoi(id)
	if err != nil {
		return friend, fmt.Errorf("problem converting %v to number: %v", id, err)
	}
	friend.ID = friendIDI

	friendDisplayOrderI, err := strconv.Atoi(displayOrder)
	if err != nil {
		return friend, fmt.Errorf("problem converting friend display order '%v' to number: %v", displayOrder, err)
	}
	friend.DisplayOrder = friendDisplayOrderI

	friend.Name = r.Form.Get(fmt.Sprintf("friend-%s-name", id))

	return friend, nil
}

func getYear(r *http.Request, yearS string) (db.Year, error) {
	var year db.Year

	yearI, err := strconv.Atoi(yearS)
	if err != nil {
		return year, fmt.Errorf("problem converting year (%v) to number: %v", year, err)
	}
	year.Value = yearI

	yearActive := r.Form.Get("year-active")
	year.Active = yearS == yearActive

	return year, nil
}
