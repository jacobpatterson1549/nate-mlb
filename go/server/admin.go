package server

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
	"golang.org/x/crypto/bcrypt"
)

var (
	adminActions = map[string](func(db.SportType, *http.Request) error){
		"friends":  updateFriends,
		"players":  updatePlayers,
		"years":    updateYears,
		"cache":    clearCache,
		"password": resetPassword,
	}
)

func handleAdminPostRequest(st db.SportType, r *http.Request) error {
	if err := verifyUserPassword(r); err != nil {
		return err
	}
	actionParam := r.FormValue("action")
	if action, ok := adminActions[actionParam]; ok {
		return action(st, r)
	}
	return errors.New("invalid admin action")
}

func handleAdminSearchRequest(st db.SportType, year int, r *http.Request) ([]request.PlayerSearchResult, error) {
	searchQuery := r.FormValue("q")
	if len(searchQuery) == 0 {
		return nil, errors.New("missing search query param: q")
	}
	playerTypeID := r.FormValue("pt")
	if len(playerTypeID) == 0 {
		return nil, errors.New("missing player type query param: pt")
	}
	playerTypeIDI, err := strconv.Atoi(playerTypeID)
	if err != nil {
		return nil, fmt.Errorf("converting playerTypeID %v' to number: %w", playerTypeID, err)
	}
	playerType := db.PlayerType(playerTypeIDI)
	activePlayersOnly := r.FormValue("apo")
	activePlayersOnlyB := activePlayersOnly == "on"

	searcher, ok := request.Searchers[playerType]
	if !ok {
		return nil, fmt.Errorf("finding searcher for playerType %v", playerType)
	}
	return searcher.PlayerSearchResults(playerType, searchQuery, year, activePlayersOnlyB)
}

func handleAdminPasswordRequest(r *http.Request) error {
	password := r.FormValue("password")
	if len(password) == 0 {
		return errors.New("missing form password form param")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	return db.AddUser("admin", hashedPassword)
}

func updatePlayers(st db.SportType, r *http.Request) error {
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

	err := db.SavePlayers(st, players)
	if err != nil {
		return err
	}
	return db.ClearStat(st)
}

func updateFriends(st db.SportType, r *http.Request) error {
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

	err := db.SaveFriends(st, friends)
	if err != nil {
		return err
	}
	return db.ClearStat(st)
}

func updateYears(st db.SportType, r *http.Request) error {
	var years []db.Year
	for _, y := range r.Form["year"] {
		year, err := getYear(r, y)
		if err != nil {
			return err
		}
		years = append(years, year)
	}

	// (does not forcefully update cache if active year changed)
	return db.SaveYears(st, years)
}

func clearCache(st db.SportType, r *http.Request) error {
	request.ClearCache()
	return db.ClearStat(st)
}

func resetPassword(st db.SportType, r *http.Request) error {
	username := r.FormValue("username")
	newPassword := r.FormValue("newPassword")
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	return db.SetUserPassword(username, hashedPassword)
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
		return fmt.Errorf("verifying password: %w", err)
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hashedPassword), nil
}

func getPlayer(r *http.Request, id, displayOrder string) (db.Player, error) {
	var player db.Player

	IDI, err := strconv.Atoi(id)
	if err != nil {
		return player, fmt.Errorf("converting player id '%v' to number: %w", id, err)
	}
	player.ID = db.ID(IDI)

	displayOrderI, err := strconv.Atoi(displayOrder)
	if err != nil {
		return player, fmt.Errorf("converting player display order '%v' to number: %w", displayOrder, err)
	}
	player.DisplayOrder = displayOrderI

	playerType := r.FormValue(fmt.Sprintf("player-%s-player-type", id))
	playerTypeI, err := strconv.Atoi(playerType)
	if err != nil {
		return player, fmt.Errorf("converting player type '%v' to number: %w", playerType, err)
	}
	player.PlayerType = db.PlayerType(playerTypeI)

	sourceID := r.FormValue(fmt.Sprintf("player-%s-source-id", id))
	sourceIDI, err := strconv.Atoi(sourceID)
	if err != nil {
		return player, fmt.Errorf("converting player source id '%v' to number: %w", sourceID, err)
	}
	player.SourceID = db.SourceID(sourceIDI)

	friendID := r.FormValue(fmt.Sprintf("player-%s-friend-id", id))
	friendIDI, err := strconv.Atoi(friendID)
	if err != nil {
		return player, fmt.Errorf("converting player friend id '%v' to number: %w", friendID, err)
	}
	player.FriendID = db.ID(friendIDI)

	return player, nil
}

func getFriend(r *http.Request, id, displayOrder string) (db.Friend, error) {
	var friend db.Friend

	friendIDI, err := strconv.Atoi(id)
	if err != nil {
		return friend, fmt.Errorf("converting friend id '%v' to number: %w", id, err)
	}
	friend.ID = db.ID(friendIDI)

	friendDisplayOrderI, err := strconv.Atoi(displayOrder)
	if err != nil {
		return friend, fmt.Errorf("converting friend display order '%v' to number: %w", displayOrder, err)
	}
	friend.DisplayOrder = friendDisplayOrderI

	friend.Name = r.FormValue(fmt.Sprintf("friend-%s-name", id))

	return friend, nil
}

func getYear(r *http.Request, yearS string) (db.Year, error) {
	var year db.Year

	yearI, err := strconv.Atoi(yearS)
	if err != nil {
		return year, fmt.Errorf("converting year '%v to number: %w", year, err)
	}
	year.Value = yearI

	yearActive := r.FormValue("year-active")
	year.Active = yearS == yearActive

	return year, nil
}
