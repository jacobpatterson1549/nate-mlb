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
	actionParam := r.FormValue("action")
	actions := map[string](func(*http.Request) error){
		"password": setPassword,
		"cache":    clearCache,
		"friends":  updateFriends,
		"players":  updatePlayers,
		"years":    updateYears,
	}
	if action, ok := actions[actionParam]; ok {
		return action(r)
	}
	return errors.New("invalid admin action")
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

	return nullEtlJSON()
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
	if err != nil {
		return err
	}
	return nullEtlJSON()
}

func updatePlayers(r *http.Request) error {
	currentPassword := r.FormValue("currentPassword")
	if err := verifyPassword(currentPassword); err != nil {
		return err
	}

	players := []Player{}
	re := regexp.MustCompile("^player-([0-9]+)-display-order$")
	for k, v := range r.Form {
		if idMatches := re.FindStringSubmatch(k); len(idMatches) > 0 {
			ID, err := strconv.Atoi(idMatches[1])
			if err != nil {
				return err
			}
			playerDisplayOrderI, err := strconv.Atoi(v[0])
			if err != nil {
				return err
			}
			playerTypeID := r.Form.Get(fmt.Sprintf("player-%d-type-id", ID))
			playerTypeIDI := 0
			if len(playerTypeID) > 0 {
				playerTypeIDI, err = strconv.Atoi(playerTypeID)
				if err != nil {
					return err
				}
			}
			playerID := r.Form.Get(fmt.Sprintf("player-%d-player-id", ID))
			playerIDI, err := strconv.Atoi(playerID)
			if err != nil {
				return err
			}
			friendID := r.Form.Get(fmt.Sprintf("player-%d-friend-id", ID))
			friendIDI := 0
			if len(friendID) > 0 {
				friendIDI, err = strconv.Atoi(friendID)
				if err != nil {
					return err
				}
			}
			players = append(players, Player{
				id:           ID,
				displayOrder: playerDisplayOrderI,
				playerTypeID: playerTypeIDI, // only needed for new Players
				playerID:     playerIDI,
				friendID:     friendIDI, // only needed for new Players
			})
		}
	}

	err := setPlayers(players)
	if err != nil {
		return err
	}
	return nullEtlJSON()
}

func updateYears(r *http.Request) error {
	currentPassword := r.FormValue("currentPassword")
	if err := verifyPassword(currentPassword); err != nil {
		return err
	}

	activeYear := r.Form.Get("active-year")
	if len(activeYear) == 0 {
		return errors.New("missing value for active-year")
	}
	activeYearI, err := strconv.Atoi(activeYear)
	if err != nil {
		return err
	}
	years := r.Form["year"]
	yearsI := make([]int, len(years))
	for i, year := range years {
		yearI, err := strconv.Atoi(year)
		if err != nil {
			return err
		}
		yearsI[i] = yearI
	}

	// calling getEtlStats() will update the cache if needed (no player/friend data has changed)
	return setYears(activeYearI, yearsI)
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
