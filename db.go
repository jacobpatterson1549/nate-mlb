package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// IDs of constant db enums
const (
	playerTypeTeam     = 1
	playerTypeHitting  = 2
	playerTypePitching = 3
)

func getDb() (*sql.DB, error) {
	driverName := "postgres"
	datasourceName := os.Getenv("DATABASE_URL")
	return sql.Open(driverName, datasourceName)
}

func getFriendPlayerInfo() (FriendPlayerInfo, error) {
	fpi := FriendPlayerInfo{}

	db, err := getDb()
	if err != nil {
		return fpi, nil
	}
	defer db.Close()

	friends, err := getFriends(db)
	if err != nil {
		return fpi, err
	}
	playerTypes, err := getPlayerTypes(db)
	if err != nil {
		return fpi, err
	}
	players, err := getPlayers(db)
	if err != nil {
		return fpi, err
	}
	activeYear, err := getActiveYear()
	if err != nil {
		return fpi, err
	}

	fpi.friends = friends
	fpi.playerTypes = playerTypes
	fpi.players = players
	fpi.year = activeYear
	return fpi, nil
}

func getEtlStats() (EtlStats, error) {
	var es EtlStats

	db, err := getDb()
	if err != nil {
		return es, nil
	}
	defer db.Close()

	var year int
	var etlJSON sql.NullString
	row := db.QueryRow("SELECT year, etl_json FROM stats WHERE active")
	err = row.Scan(&year, &etlJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("no active year")
		}
		return es, err
	}
	fetchStats := true
	currentTime := getUtcTime()
	if etlJSON.Valid {
		err = json.Unmarshal([]byte(etlJSON.String), &es)
		if err != nil {
			return es, err
		}
		fetchStats = es.isStale(currentTime)
	}
	if fetchStats {
		scoreCategories, err := getStats()
		if err != nil {
			return es, err
		}
		es.ScoreCategories = scoreCategories
		es.EtlTime = currentTime
		etlJSON, err := json.Marshal(es)
		if err != nil {
			return es, err
		}
		result, err := db.Exec("UPDATE stats SET etl_json = $1 WHERE year = $2", etlJSON, year)
		if err != nil {
			return es, err
		}
		err = expectSingleRowAffected(result)
	}
	return es, err
}

func nullEtlJSON() error {
	db, err := getDb()
	if err == nil {
		defer db.Close()
		_, err = db.Exec("UPDATE stats SET etl_json = NULL WHERE active")
	}
	return err
}

func getActiveYear() (int, error) {
	var activeYear int

	db, err := getDb()
	if err != nil {
		return activeYear, nil
	}
	defer db.Close()

	row := db.QueryRow("SELECT year FROM stats WHERE active")
	err = row.Scan(&activeYear)
	if err == sql.ErrNoRows {
		err = errors.New("no active year")
	}
	return activeYear, err
}

func getYears() ([]Year, error) {
	years := []Year{}

	db, err := getDb()
	if err != nil {
		return years, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT year, active FROM stats ORDER BY year ASC")
	if err != nil {
		return years, fmt.Errorf("Error reading years: %q", err)
	}
	defer rows.Close()

	activeYearFound := false
	var active sql.NullBool
	i := 0
	for rows.Next() {
		years = append(years, Year{})
		err = rows.Scan(&years[i].Value, &active)
		if err != nil {
			return years, fmt.Errorf("Problem reading data: %q", err)
		}
		if active.Valid && active.Bool {
			if activeYearFound {
				return years, errors.New("multiple active years in db")
			}
			activeYearFound = true
			years[i].Active = true
		}
		i++
	}
	if !activeYearFound && len(years) > 0 {
		return years, errors.New("no active year in db")
	}
	return years, nil
}

func setYears(activeYear int, years []int) error {
	// TODO: use same db connection for getting existing years, updating/inserting/deleting ones
	currentYears, err := getYears()
	if err != nil {
		return err
	}
	currentYearsMap := make(map[int]bool)
	for _, year := range currentYears {
		currentYearsMap[year.Value] = true
	}

	insertYears := []int{}
	activeYearPresent := false
	for _, year := range years {
		if year == activeYear {
			activeYearPresent = true
		}
		if _, ok := currentYearsMap[year]; ok {
			delete(currentYearsMap, year) // no need to update
		} else {
			insertYears = append(insertYears, year)
		}
	}
	if len(years) > 0 && !activeYearPresent {
		return fmt.Errorf("active year %d not present in years: %q", activeYear, years)
	}
	deleteYears := currentYearsMap

	db, err := getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var result sql.Result
	for year := range deleteYears {
		if err == nil {
			result, err = tx.Exec(
				"DELETE FROM stats WHERE year = $1",
				year)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for _, year := range insertYears {
		if err == nil {
			result, err = tx.Exec(
				"INSERT INTO stats (year) VALUES ($1)",
				year)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	// remove active year
	if err == nil && len(years) > 0 {
		result, err = tx.Exec("UPDATE stats SET active = NULL WHERE active")
	}
	// set active year
	if err == nil && len(years) > 0 {
		// TODO: make "func affectOneRow(tx *sql.Tx, sql string) error" function
		result, err = tx.Exec(
			"UPDATE stats SET active = TRUE WHERE year = $1",
			activeYear)
		if err != nil {
			err = expectSingleRowAffected(result)
		}
	}
	if err == nil {
		err = tx.Commit()
	} else {
		if err2 := tx.Rollback(); err2 != nil {
			err = fmt.Errorf("Error: %s, ROLLBACK ERROR: %s", err.Error(), err2.Error())
		}
	}

	return err
}

// TODO: use shared logic to request friends, playerTypes, players (but with helper mapper functions)
func getFriends(db *sql.DB) ([]Friend, error) {
	rows, err := db.Query("SELECT f.id, f.display_order, f.name FROM friends AS f JOIN stats AS s ON f.year = s.year WHERE s.active ORDER BY f.display_order ASC")
	if err != nil {
		return nil, fmt.Errorf("Error reading friends: %q", err)
	}
	defer rows.Close()

	friends := []Friend{}
	i := 0
	for rows.Next() {
		friends = append(friends, Friend{})
		err = rows.Scan(&friends[i].id, &friends[i].displayOrder, &friends[i].name)
		if err != nil {
			return nil, fmt.Errorf("Problem reading data: %q", err)
		}
		i++
	}
	return friends, nil
}

func getPlayerTypes(db *sql.DB) ([]PlayerType, error) {
	rows, err := db.Query("SELECT id, name, description FROM player_types ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("Error reading playerTypes: %q", err)
	}
	defer rows.Close()

	playerTypes := []PlayerType{}
	i := 0
	for rows.Next() {
		playerTypes = append(playerTypes, PlayerType{})
		err = rows.Scan(&playerTypes[i].id, &playerTypes[i].name, &playerTypes[i].description)
		if err != nil {
			return nil, fmt.Errorf("Problem reading data: %q", err)
		}
		i++
	}
	return playerTypes, nil
}

func getPlayers(db *sql.DB) ([]Player, error) {
	rows, err := db.Query("SELECT p.id, p.display_order, p.player_type_id, p.player_id, p.friend_id FROM players AS p JOIN stats AS s ON p.year = s.year WHERE s.active ORDER BY p.player_type_id, p.friend_id, p.display_order")
	if err != nil {
		return nil, fmt.Errorf("Error reading players: %q", err)
	}
	defer rows.Close()

	players := []Player{}
	i := 0
	for rows.Next() {
		players = append(players, Player{})
		err = rows.Scan(&players[i].id, &players[i].displayOrder, &players[i].playerTypeID, &players[i].playerID, &players[i].friendID)
		if err != nil {
			return nil, fmt.Errorf("problem reading data: %q", err)
		}
		i++
	}
	return players, nil
}

func getUserPassword(username string) (string, error) {
	var v string

	db, err := getDb()
	if err != nil {
		return v, nil
	}
	defer db.Close()

	row := db.QueryRow("SELECT password FROM users WHERE username = $1", username)
	return v, row.Scan(&v)
}

func setUserPassword(username, password string) error {
	db, err := getDb()
	if err != nil {
		return err
	}

	result, err := db.Exec("UPDATE users SET password = $1 WHERE username = $2", password, username)
	if err != nil {
		return err
	}
	return expectSingleRowAffected(result)
}

func expectSingleRowAffected(r sql.Result) error {
	rows, err := r.RowsAffected()
	if err == nil && rows != 1 {
		return fmt.Errorf("expected to updated 1 row, but updated %d", rows)
	}
	return err
}

func setFriends(futureFriends []Friend) error {
	db, err := getDb()
	if err != nil {
		return nil
	}
	defer db.Close()

	friends, err := getFriends(db)
	if err != nil {
		return err
	}
	previousFriends := make(map[int]Friend)
	for _, friend := range friends {
		previousFriends[friend.id] = friend
	}

	insertFriends := []Friend{}
	updateFriends := []Friend{}
	for _, friend := range futureFriends {
		previousFriend, ok := previousFriends[friend.id]
		if !ok {
			insertFriends = append(insertFriends, friend)
		} else {
			if friend.displayOrder != previousFriend.displayOrder || friend.name != previousFriend.name {
				updateFriends = append(updateFriends, friend)
			}
		}
		delete(previousFriends, friend.id)
	}
	deleteFriends := previousFriends

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var result sql.Result
	for _, friend := range insertFriends {
		if err == nil {
			result, err = tx.Exec(
				"INSERT INTO friends (display_order, name, year) SELECT $1, $2, year FROM stats AS s WHERE s.active",
				friend.displayOrder,
				friend.name)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for _, friend := range updateFriends {
		if err == nil {
			result, err = tx.Exec(
				"UPDATE friends SET display_order = $1, name = $2 WHERE id = $3",
				friend.displayOrder,
				friend.name,
				friend.id)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for friendID := range deleteFriends {
		if err == nil {
			result, err = tx.Exec(
				"DELETE FROM friends WHERE id = $1",
				friendID)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	if err == nil {
		err = tx.Commit()
	} else if err2 := tx.Rollback(); err2 != nil {
		err = fmt.Errorf("Error: %s, ROLLBACK ERROR: %s", err.Error(), err2.Error())
	}

	return err
}

func setPlayers(futurePlayers []Player) error {
	db, err := getDb()
	if err != nil {
		return nil
	}
	defer db.Close()

	players, err := getPlayers(db)
	if err != nil {
		return err
	}
	previousPlayers := make(map[int]Player)
	for _, player := range players {
		previousPlayers[player.id] = player
	}

	insertPlayers := []Player{}
	updatePlayers := []Player{}
	for _, player := range futurePlayers {
		previousPlayer, ok := previousPlayers[player.id]
		if !ok {
			insertPlayers = append(insertPlayers, player)
		} else if player.displayOrder != previousPlayer.displayOrder { // can only update display order
			updatePlayers = append(updatePlayers, player)
		}
		delete(previousPlayers, player.id)
	}
	deletePlayers := previousPlayers

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var result sql.Result
	for _, player := range insertPlayers {
		if err == nil {
			result, err = tx.Exec(
				"INSERT INTO players (display_order, player_type_id, player_id, friend_id, year) SELECT $1, $2, $3, $4, year FROM stats AS s WHERE s.active",
				player.displayOrder,
				player.playerTypeID,
				player.playerID,
				player.friendID)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for _, player := range updatePlayers {
		if err == nil {
			result, err = tx.Exec(
				"UPDATE players SET display_order = $1 WHERE id = $2",
				player.displayOrder,
				player.id)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for playerID := range deletePlayers {
		if err == nil {
			result, err = tx.Exec(
				"DELETE FROM players WHERE id = $1",
				playerID)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	if err == nil {
		err = tx.Commit()
	} else {
		if err2 := tx.Rollback(); err2 != nil {
			err = fmt.Errorf("Error: %s, ROLLBACK ERROR: %s", err.Error(), err2.Error())
		}
	}

	return err
}

func getUtcTime() time.Time {
	return time.Now().UTC()
}

// SETS EtlRefreshTime and determines if it before the current time
func (es *EtlStats) isStale(currentTime time.Time) bool {
	previousHonoluluMidnight := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 10, 0, 0, 0, currentTime.Location())
	if previousHonoluluMidnight.After(currentTime) {
		previousHonoluluMidnight = previousHonoluluMidnight.Add(-24 * time.Hour)
	}
	es.EtlRefreshTime = previousHonoluluMidnight

	return es.EtlTime.Before(es.EtlRefreshTime)
}

// FriendPlayerInfo contain all the pool items for each Friend.
type FriendPlayerInfo struct {
	friends     []Friend
	playerTypes []PlayerType
	players     []Player
	year        int
}

// Year contains a year that has been set for stats and whether it is active
type Year struct {
	Value  int
	Active bool
}

// Friend contains the name of the person in the pool.
type Friend struct {
	id           int
	displayOrder int
	name         string
}

// PlayerType contain a name of a pool item.
type PlayerType struct {
	id          int
	name        string
	description string
}

// Player maps a player (of a a specific PlayerType) to a Friend.
type Player struct {
	id           int
	displayOrder int
	playerTypeID int
	playerID     int
	friendID     int
}

// EtlStats contain some score categories that were stored at a specific time
type EtlStats struct {
	EtlTime         time.Time
	EtlRefreshTime  time.Time
	ScoreCategories []ScoreCategory
}
