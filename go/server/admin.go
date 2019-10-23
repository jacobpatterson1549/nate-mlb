package server

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

type (
	adminDatastore interface {
		SaveYears(st db.SportType, futureYears []db.Year) error
		SaveFriends(st db.SportType, futureFriends []db.Friend) error
		SavePlayers(st db.SportType, futurePlayers []db.Player) error
		ClearStat(st db.SportType) error
		SetUserPassword(username string, p db.Password) error
		IsCorrectUserPassword(username string, p db.Password) (bool, error)
	}
	adminCache interface {
		Clear()
	}
)

var (
	playerDisplayOrderRE = regexp.MustCompile("^player-([0-9]+)-display-order$")
	friendDisplayOrderRE = regexp.MustCompile("^friend-([0-9]+)-display-order$")
)

func handleAdminPostRequest(ds adminDatastore, c adminCache, st db.SportType, r *http.Request) error {
	if err := verifyUserPassword(ds, r); err != nil {
		return err
	}
	actionParam := r.FormValue("action")
	var adminAction func(ds adminDatastore, st db.SportType, r *http.Request) error
	switch actionParam {
	case "friends":
		adminAction = updateFriends
	case "players":
		adminAction = updatePlayers
	case "years":
		adminAction = updateYears
	case "cache":
		adminAction = clearStat
		c.Clear()
	case "password":
		adminAction = resetPassword
	default:
		return fmt.Errorf("invalid admin action: %v", actionParam)
	}
	return adminAction(ds, st, r)
}

func handleAdminSearchRequest(year int, searchers map[db.PlayerType]request.Searcher, r *http.Request) ([]request.PlayerSearchResult, error) {
	searchQuery := r.FormValue("q")
	if len(searchQuery) == 0 {
		return nil, fmt.Errorf("missing search query param: q")
	}
	playerTypeID := r.FormValue("pt")
	if len(playerTypeID) == 0 {
		return nil, fmt.Errorf("missing player type query param: pt")
	}
	playerTypeIDI, err := strconv.Atoi(playerTypeID)
	if err != nil {
		return nil, fmt.Errorf("converting playerTypeID %v' to number: %w", playerTypeID, err)
	}
	playerType := db.PlayerType(playerTypeIDI)
	activePlayersOnly := r.FormValue("apo")
	activePlayersOnlyB := activePlayersOnly == "on"

	searcher, ok := searchers[playerType]
	if !ok {
		return nil, fmt.Errorf("no searcher for playerType %v", playerType)
	}
	return searcher.Search(playerType, year, searchQuery, activePlayersOnlyB)
}

func updatePlayers(ds adminDatastore, st db.SportType, r *http.Request) error {
	var players []db.Player
	for k, v := range r.Form {
		if matches := playerDisplayOrderRE.FindStringSubmatch(k); len(matches) > 1 {
			player, err := getPlayer(st, r, matches[1], v[0])
			if err != nil {
				return err
			}
			players = append(players, player)
		}
	}

	err := ds.SavePlayers(st, players)
	if err != nil {
		return err
	}
	return ds.ClearStat(st)
}

func updateFriends(ds adminDatastore, st db.SportType, r *http.Request) error {
	var friends []db.Friend

	for k, v := range r.Form {
		if matches := friendDisplayOrderRE.FindStringSubmatch(k); len(matches) > 1 {
			friend, err := getFriend(r, matches[1], v[0])
			if err != nil {
				return err
			}
			friends = append(friends, friend)
		}
	}

	err := ds.SaveFriends(st, friends)
	if err != nil {
		return err
	}
	return ds.ClearStat(st)
}

func updateYears(ds adminDatastore, st db.SportType, r *http.Request) error {
	var years []db.Year
	for _, y := range r.Form["year"] {
		year, err := getYear(r, y)
		if err != nil {
			return err
		}
		years = append(years, year)
	}

	return ds.SaveYears(st, years)
}

func clearStat(ds adminDatastore, st db.SportType, r *http.Request) error {
	return ds.ClearStat(st)
}

func resetPassword(ds adminDatastore, st db.SportType, r *http.Request) error {
	username := r.FormValue("username")
	newPassword := r.FormValue("newPassword")
	return ds.SetUserPassword(username, db.Password(newPassword))
}

func verifyUserPassword(ds adminDatastore, r *http.Request) error {
	username := r.FormValue("username")
	password := r.FormValue("password")
	correctPassword, err := ds.IsCorrectUserPassword(username, db.Password(password))
	if err != nil {
		return fmt.Errorf("verifying password: %w", err)
	}
	if !correctPassword {
		return fmt.Errorf("Incorrect Password")
	}
	return nil
}

func getPlayer(st db.SportType, r *http.Request, id, displayOrder string) (db.Player, error) {
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
