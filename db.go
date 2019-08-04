package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
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

	fpi.friends = friends
	fpi.playerTypes = playerTypes
	fpi.players = players
	return fpi, nil
}

// TODO: use shared logic to request friends, playerTypes, players (but with helper mapper functions)
func getFriends(db *sql.DB) ([]Friend, error) {
	rows, err := db.Query("SELECT id, display_order, name FROM friends ORDER BY display_order ASC")
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
	rows, err := db.Query("SELECT id, name FROM player_types ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("Error reading playerTypes: %q", err)
	}
	defer rows.Close()

	playerTypes := []PlayerType{}
	i := 0
	for rows.Next() {
		playerTypes = append(playerTypes, PlayerType{})
		err = rows.Scan(&playerTypes[i].id, &playerTypes[i].name)
		if err != nil {
			return nil, fmt.Errorf("Problem reading data: %q", err)
		}
		i++
	}
	return playerTypes, nil
}

func getPlayers(db *sql.DB) ([]Player, error) {
	rows, err := db.Query("SELECT id, display_order, player_type_id, player_id, friend_id FROM players ORDER BY player_type_id, friend_id, display_order")
	if err != nil {
		return nil, fmt.Errorf("Error reading playerTypes: %q", err)
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

func getKeyStoreValue(key string) (string, error) {
	var v string

	db, err := getDb()
	if err != nil {
		return v, nil
	}
	defer db.Close()

	row := db.QueryRow("SELECT v FROM key_store WHERE k = $1", key)
	err = row.Scan(&v)
	return v, err // TODO: Can `return v, row.Scan(&v)` be used?
}

func setKeyStoreValue(key string, value string) error {
	db, err := getDb()
	if err != nil {
		return err
	}

	result, err := db.Exec("UPDATE key_store SET v = $1 WHERE k = $2", value, key)
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
				"INSERT INTO friends (display_order, name) VALUES ($1, $2)",
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
				"INSERT INTO players (display_order, playerTypeID, playerID, friendID) VALUES ($1, $2, $3, $4)",
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

// FriendPlayerInfo contain all the pool items for each Friend.
type FriendPlayerInfo struct {
	friends     []Friend
	playerTypes []PlayerType
	players     []Player
}

// Friend contains the name of the person in the pool.
type Friend struct {
	id           int
	displayOrder int
	name         string
}

// PlayerType contain a name of a pool item.
type PlayerType struct {
	id   int
	name string
}

// Player maps a player (of a a specific PlayerType) to a Friend.
type Player struct {
	id           int
	displayOrder int
	playerTypeID int
	playerID     int
	friendID     int
}
