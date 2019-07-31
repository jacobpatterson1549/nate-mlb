package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func getFriendPlayerInfo() (FriendPlayerInfo, error) {
	driverName := "postgres"
	datasourceName := os.Getenv("DATABASE_URL")
	db, err := sql.Open(driverName, datasourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	friends, err := getFriends(db)
	if err != nil {
		return FriendPlayerInfo{}, err
	}
	playerTypes, err := getPlayerTypes(db)
	if err != nil {
		return FriendPlayerInfo{}, err
	}
	players, err := getPlayers(db)
	if err != nil {
		return FriendPlayerInfo{}, err
	}

	return FriendPlayerInfo{
		friends,
		playerTypes,
		players,
	}, nil
}

// TODO: use shared logic to request friends, playerTypes, players (but with helper mapper functions)
func getFriends(db *sql.DB) ([]Friend, error) {
	rows, err := db.Query("SELECT id, name FROM friends ORDER BY display_order ASC")
	if err != nil {
		return nil, fmt.Errorf("Error reading friends: %q", err)
	}
	defer rows.Close()

	friends := []Friend{}
	i := 0
	for rows.Next() {
		friends = append(friends, Friend{})
		err = rows.Scan(&friends[i].id, &friends[i].name)
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
	rows, err := db.Query("SELECT player_type_id, player_id, friend_id FROM players ORDER BY player_type_id, friend_id, display_order")
	if err != nil {
		return nil, fmt.Errorf("Error reading playerTypes: %q", err)
	}
	defer rows.Close()

	players := []Player{}
	i := 0
	for rows.Next() {
		players = append(players, Player{})
		err = rows.Scan(&players[i].playerTypeID, &players[i].playerID, &players[i].friendID)
		if err != nil {
			return nil, fmt.Errorf("Problem reading data: %q", err)
		}
		i++
	}
	return players, nil
}

// FriendPlayerInfo contain all the pool items for each Friend.
type FriendPlayerInfo struct {
	friends     []Friend
	playerTypes []PlayerType
	players     []Player
}

// Friend contains the name of the person in the pool.
type Friend struct {
	id   int
	name string
}

// PlayerType contain a name of a pool item.
type PlayerType struct {
	id   int
	name string
}

// Player maps a player (of a a specific PlayerType) to a Friend.
type Player struct {
	playerTypeID int
	playerID     int
	friendID     int
}
