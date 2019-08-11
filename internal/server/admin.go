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
		"password": resetPassword,
		"cache":    clearCache,
		"friends":  updateFriends,
		"players":  updatePlayers,
		"years":    updateYears,
	}
)

func handleAdminRequest(r *http.Request) error {
	actionParam := r.FormValue("action")
	if action, ok := adminActions[actionParam]; ok {
		return action(r)
	}
	return errors.New("invalid admin action")
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
	return db.SetUserPassword("admin", hashedPassword)
}

func clearCache(r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}

	return db.NullEtlJSON()
}

func updateFriends(r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}

	friends := []db.Friend{}
	re := regexp.MustCompile("^friend-([0-9]+)-display-order$")
	for k, v := range r.Form {
		if matches := re.FindStringSubmatch(k); len(matches) > 0 {
			friendIDi, err := strconv.Atoi(matches[1])
			if err != nil {
				return fmt.Errorf("problem converting %v to number: %v", matches[1], err)
			}
			friendDisplayOrderI, err := strconv.Atoi(v[0])
			if err != nil {
				return fmt.Errorf("problem converting friend display order '%v' to number: %v", v[0], err)
			}
			friendName := r.Form.Get(fmt.Sprintf("friend-%d-name", friendIDi))
			friends = append(friends, db.Friend{
				ID:           friendIDi,
				DisplayOrder: friendDisplayOrderI,
				Name:         friendName,
			})
		}
	}

	err := db.SetFriends(friends)
	if err != nil {
		return err
	}
	return db.NullEtlJSON()
}

func updatePlayers(r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}

	players := []db.Player{}
	re := regexp.MustCompile("^player-([0-9]+)-display-order$")
	for k, v := range r.Form {
		if matches := re.FindStringSubmatch(k); len(matches) > 0 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				return fmt.Errorf("problem converting %v to number: %v", matches[1], err)
			}
			playerDisplayOrderI, err := strconv.Atoi(v[0])
			if err != nil {
				return fmt.Errorf("problem converting player display order '%v' to number: %v", v[0], err)
			}
			playerTypeID := r.Form.Get(fmt.Sprintf("player-%d-player-type-id", id))
			var playerTypeIDI int
			if len(playerTypeID) > 0 { // not specified when updating existing player
				playerTypeIDI, err = strconv.Atoi(playerTypeID)
				if err != nil {
					return fmt.Errorf("problem converting player type id '%v' to number: %v", playerTypeID, err)
				}
			}
			playerID := r.Form.Get(fmt.Sprintf("player-%d-player-id", id))
			playerIDI, err := strconv.Atoi(playerID)
			if err != nil {
				return fmt.Errorf("problem converting player id '%v' to number: %v", playerID, err)
			}
			friendID := r.Form.Get(fmt.Sprintf("player-%d-friend-id", id))
			var friendIDI int
			if len(friendID) > 0 { // not specified when updating existing player
				friendIDI, err = strconv.Atoi(friendID)
				if err != nil {
					return fmt.Errorf("problem converting player friend id '%v' to number: %v", friendID, err)
				}
			}
			players = append(players, db.Player{
				ID:           id,
				DisplayOrder: playerDisplayOrderI,
				PlayerTypeID: playerTypeIDI,
				PlayerID:     playerIDI,
				FriendID:     friendIDI,
			})
		}
	}

	err := db.SetPlayers(players)
	if err != nil {
		return err
	}
	return db.NullEtlJSON()
}

func updateYears(r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}

	activeYear := r.Form.Get("year-active")
	if len(activeYear) == 0 {
		return errors.New("missing value for year-active")
	}
	activeYearI, err := strconv.Atoi(activeYear)
	if err != nil {
		return fmt.Errorf("problem converting active year (%v) to number: %v", activeYear, err)
	}
	years := r.Form["year"]
	yearsI := make([]int, len(years))
	for i, year := range years {
		yearI, err := strconv.Atoi(year)
		if err != nil {
			return fmt.Errorf("problem converting year (%v) to number: %v", year, err)
		}
		yearsI[i] = yearI
	}

	// (does not forcefully update cache if active year changed)
	return db.SetYears(activeYearI, yearsI)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("problem hashing password: %v", err)
	}
	return string(hashedPassword), nil
}

func verifyUserPassword(r *http.Request) error {
	username := r.FormValue("username")
	password := r.FormValue("password")
	hashedPassword, err := db.GetUserPassword(username)
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
